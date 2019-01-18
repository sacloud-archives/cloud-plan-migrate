package migrate

import (
	"fmt"
	"sync"

	"github.com/sacloud/cloud-plan-migrate/iaas"
	"github.com/sacloud/libsacloud/sacloud"
)

type Options struct {
	DisableBoot    bool
	DeleteDisks    bool
	MaxWorkerCount int
	Logger         Logger
}

type Migration struct {
	client         iaas.Client
	status         []*ServerStatus
	working        []*ServerStatus
	lock           sync.Mutex
	maxWorkerCount int
}

func NewMigration(client iaas.Client, serverIDs []int64, options *Options) (*Migration, error) {

	var status []*ServerStatus
	if options == nil {
		options = &Options{}
	}

	for _, id := range serverIDs {

		s := &ServerStatus{targetServerID: id}

		server, err := client.ServerByID(id)
		if err != nil {
			return nil, err
		}

		s.serverName = server.Name
		serverLogPref := fmt.Sprintf(": Server[%d:%s] :", server.ID, server.Name)

		newPlan, err := client.FindServerPlan(server.GetCPU(), server.GetMemoryGB())
		if err != nil {
			return nil, err
		}
		s.newPlan = newPlan

		var disks []*DiskStatus
		for _, disk := range server.Disks {
			diskLogPref := fmt.Sprintf(":   Disk[%d:%s] :", disk.ID, server.Name) // サーバ名を利用

			d := &DiskStatus{
				originalID: disk.ID,
				sizeMB:     disk.GetSizeMB(),
				stepClone: &step{
					needProcess: true,
					logger:      options.Logger,
					logPrefix:   fmt.Sprintf("%s %s", diskLogPref, "Clone Disk"),
				},
				stepDelete: &step{
					needProcess: options.DeleteDisks,
					logger:      options.Logger,
					logPrefix:   fmt.Sprintf("%s %s", diskLogPref, "Delete Disk"),
				},
			}
			disks = append(disks, d)
		}
		s.Disks = disks
		if len(disks) > 0 {
			s.stepConnectDisks = &step{
				needProcess: true,
				logger:      options.Logger,
				logPrefix:   fmt.Sprintf("%s %s", serverLogPref, "Connect Disk"),
			}
			s.stepDisconnectDisks = &step{
				needProcess: true,
				logger:      options.Logger,
				logPrefix:   fmt.Sprintf("%s %s", serverLogPref, "Disconnect Disk"),
			}
		}
		s.stepShutdown = &step{
			needProcess: server.IsUp(),
			logger:      options.Logger,
			logPrefix:   fmt.Sprintf("%s %s", serverLogPref, "Shutdown Server"),
		}
		s.stepPlanMigrate = &step{
			needProcess: true,
			logger:      options.Logger,
			logPrefix:   fmt.Sprintf("%s %s", serverLogPref, "Migrate Server Plan"),
		}
		s.stepBoot = &step{
			needProcess: !options.DisableBoot,
			logger:      options.Logger,
			logPrefix:   fmt.Sprintf("%s %s", serverLogPref, "Boot Server"),
		}

		status = append(status, s)
	}

	return &Migration{
		client:         client,
		status:         status,
		maxWorkerCount: options.MaxWorkerCount,
	}, nil
}

func (m *Migration) Apply() {

	var wg sync.WaitGroup
	wg.Add(len(m.status))

	limit := make(chan struct{}, m.maxWorkerCount)

	for i := range m.status {
		go func(status *ServerStatus) {
			limit <- struct{}{}
			m.addWorking(status)
			m.applyServer(status)
			m.removeWorking(status)
			<-limit
			wg.Done()
		}(m.status[i])
	}

	wg.Wait()
}

func (m *Migration) Working() []*ServerStatus {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.working
}

func (m *Migration) HasErrors() []*ServerStatus {
	var errs []*ServerStatus
	for _, s := range m.status {
		if s.Err != nil {
			errs = append(errs, s)
		}
	}
	return errs
}

func (m *Migration) addWorking(status *ServerStatus) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.working = append(m.working, status)
}

func (m *Migration) removeWorking(status *ServerStatus) {
	m.lock.Lock()
	defer m.lock.Unlock()
	var result []*ServerStatus
	for _, v := range m.working {
		if v != status {
			result = append(result, v)
		}
	}
	m.working = result
}

func (m *Migration) applyServer(status *ServerStatus) {
	// shutdown(if need)
	if err := m.handleSteps(m.shutdownServer, status, status.stepShutdown); err != nil {
		return
	}

	// disconnect disk
	if err := m.handleSteps(m.disconnectDisks, status, status.stepDisconnectDisks); err != nil {
		return
	}

	// clone disk
	if err := m.handleSteps(m.cloneDisks, status, status.cloneDiskSteps()...); err != nil {
		return
	}

	// migrate server plan
	if err := m.handleSteps(m.migrateServerPlan, status, status.stepPlanMigrate); err != nil {
		return
	}

	// connect disk
	if err := m.handleSteps(m.connectDisks, status, status.stepConnectDisks); err != nil {
		return
	}

	// boot server
	if err := m.handleSteps(m.bootServer, status, status.stepBoot); err != nil {
		return
	}

	// delete disk
	if err := m.handleSteps(m.deleteDisks, status, status.deleteDiskSteps()...); err != nil {
		return
	}
}

func (m *Migration) handleSteps(stepFunc func(*ServerStatus) error, status *ServerStatus, steps ...*step) error {
	for _, step := range steps {
		step.start()
	}

	if err := stepFunc(status); err != nil {
		status.Err = err
		return err
	}

	for _, step := range steps {
		step.finalize()
	}

	return nil
}

func (m *Migration) shutdownServer(status *ServerStatus) error {
	if status.stepShutdown.needProcess {
		// shutdown
		if err := m.client.Shutdown(status.targetServerID); err != nil {
			status.stepShutdown.setError(err)
			return err
		}
	}
	return nil
}

func (m *Migration) disconnectDisks(status *ServerStatus) error {
	if status.stepDisconnectDisks.needProcess {
		if err := m.client.DisconnectDisks(status.targetServerID); err != nil {
			status.stepDisconnectDisks.setError(err)
			return err
		}
	}
	return nil
}

func (m *Migration) cloneDisks(status *ServerStatus) error {
	errC := make(chan error)
	doneC := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(len(status.Disks))

	for _, disk := range status.Disks {
		go func(status *DiskStatus) {
			if disk.stepClone.needProcess {
				progress, err := m.client.CloneDisk(status.originalID)
				if err != nil {
					disk.stepClone.setError(err)
					errC <- err
					return
				}

				var newDisk *sacloud.Disk
				for {
					res, ok := <-progress
					if !ok && newDisk != nil {
						status.clonedID = newDisk.ID
						status.migratedMB = newDisk.GetMigratedMB()
						status.stepClone.finalize()
						break
					}
					switch d := res.(type) {
					case *sacloud.Disk:
						status.clonedID = d.ID
						status.migratedMB = d.GetMigratedMB()
						newDisk = d
					case error:
						status.stepClone.setError(d)
						errC <- d
						return
					}
				}
			}
			wg.Done()
		}(disk)
	}

	go func() {
		wg.Wait()
		doneC <- true
	}()

	for {
		select {
		case err := <-errC:
			return err
		case <-doneC:
			return nil
		}
	}

}

func (m *Migration) migrateServerPlan(status *ServerStatus) error {
	if status.stepPlanMigrate.needProcess {
		newServer, err := m.client.ChangePlan(status.targetServerID, status.newPlan)
		if err != nil {
			status.stepPlanMigrate.setError(err)
			return err
		}
		status.migratedServerID = newServer.ID
	}
	return nil
}

func (m *Migration) connectDisks(status *ServerStatus) error {
	if status.stepConnectDisks.needProcess {
		if err := m.client.ConnectDisks(status.migratedServerID, status.clonedDiskIDs()); err != nil {
			status.stepConnectDisks.setError(err)
			return err
		}
	}
	return nil
}

func (m *Migration) bootServer(status *ServerStatus) error {
	if status.stepBoot.needProcess {
		// boot
		if err := m.client.Boot(status.migratedServerID); err != nil {
			status.stepBoot.setError(err)
			return err
		}
	}
	return nil
}

func (m *Migration) deleteDisks(status *ServerStatus) error {
	for _, disk := range status.Disks {
		if disk.stepDelete.needProcess {
			if err := m.client.DeleteDisk(disk.originalID); err != nil {
				disk.stepDelete.setError(err)
				return err
			}
		}
	}
	return nil
}
