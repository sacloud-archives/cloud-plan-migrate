package migrate

import (
	"fmt"

	"github.com/sacloud/libsacloud/sacloud"
)

type ServerStatus struct {
	Disks []*DiskStatus

	stepShutdown        *step
	stepDisconnectDisks *step
	stepConnectDisks    *step
	stepPlanMigrate     *step
	stepBoot            *step

	targetServerID   int64
	serverName       string
	migratedServerID int64
	newPlan          *sacloud.ProductServer

	Err error
}

func (s *ServerStatus) ServerID() string {
	return fmt.Sprintf("%d", s.targetServerID)
}

func (s *ServerStatus) ServerName() string {
	return s.serverName
}

func (s *ServerStatus) ShutdownStatus() string {
	return s.stepShutdown.Status()
}

func (s *ServerStatus) MigrationStatus() string {
	return s.stepPlanMigrate.Status()
}

func (s *ServerStatus) BootStatus() string {
	return s.stepBoot.Status()
}

func (s *ServerStatus) setNewDiskID(old, new int64) {
	d := s.findDiskStatus(old)
	if d != nil {
		d.clonedID = new
	}
}

func (s *ServerStatus) findDiskStatus(oldDiskID int64) *DiskStatus {
	for _, d := range s.Disks {
		if d.originalID == oldDiskID {
			return d
		}
	}
	return nil
}

func (s *ServerStatus) clonedDiskIDs() []int64 {
	var ids []int64
	for _, d := range s.Disks {
		ids = append(ids, d.clonedID)
	}
	return ids
}

func (s *ServerStatus) cloneDiskSteps() []*step {
	var steps []*step
	for _, d := range s.Disks {
		steps = append(steps, d.stepClone)
	}
	return steps
}

func (s *ServerStatus) deleteDiskSteps() []*step {
	var steps []*step
	for _, d := range s.Disks {
		steps = append(steps, d.stepDelete)
	}
	return steps
}

type DiskStatus struct {
	stepClone  *step
	stepDelete *step

	originalID int64
	sizeMB     int
	migratedMB int
	clonedID   int64
}

func (d *DiskStatus) CloneStatus() string {
	if !d.stepClone.started {
		return d.stepClone.Status()
	}

	id := fmt.Sprintf("%d", d.originalID)
	if d.stepClone.done {
		id = fmt.Sprintf("%d(cloned)", d.clonedID)
	}

	if d.cloning() {
		return fmt.Sprintf("ID:%s(%ds)\n%s", id, int(d.stepClone.elapsed().Seconds()), d.migratedStatus())
	}
	return fmt.Sprintf("ID:%s", id)
}

func (d *DiskStatus) cloning() bool {
	return d.stepClone.needProcess && d.stepClone.started && !d.stepClone.done
}

func (d *DiskStatus) DeleteStatus() string {
	return d.stepDelete.Status()
}

func (d *DiskStatus) migratedStatus() string {
	return fmt.Sprintf("%7dMB/%7dMB", d.migratedMB, d.sizeMB)
}
