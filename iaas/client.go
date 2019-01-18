package iaas

import (
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

type Client interface {
	FindAll() ([]*sacloud.Server, error)
	// Find(param *FindParameter) ([]*sacloud.Server, error)
	ServerByID(id int64) (*sacloud.Server, error)
	DiskByID(id int64) (*sacloud.Disk, error)
	FindServerPlan(core int, memoryGB int) (*sacloud.ProductServer, error)

	Shutdown(id int64) (err error)
	DisconnectDisks(serverID int64) error
	CloneDisk(id int64) (progress <-chan interface{}, err error)
	ChangePlan(serverID int64, plan *sacloud.ProductServer) (*sacloud.Server, error)
	ConnectDisks(serverID int64, diskIDs []int64) error
	Boot(id int64) (err error)
	DeleteDisk(id int64) error
}

type FindParameter struct {
	Names []string
	Ids   []int64
	Tags  []string
}

func NewClient(apiClient *api.Client) Client {
	return &client{apiClient: apiClient}
}

type client struct {
	apiClient *api.Client
}

func (c *client) FindAll() ([]*sacloud.Server, error) {
	res, err := c.apiClient.Server.Reset().Limit(1000).Find()
	if err != nil {
		return nil, err
	}
	var servers []*sacloud.Server
	for _, s := range res.Servers {
		if s.ServerPlan.Generation != sacloud.PlanG2 {
			servers = append(servers, &s)
		}
	}
	return servers, nil
}

//func (c *client) Find(param *FindParameter) ([]*sacloud.Server, error) {
//
//}

func (c *client) ServerByID(id int64) (*sacloud.Server, error) {
	return c.apiClient.Server.Read(id)
}

func (c *client) DiskByID(id int64) (*sacloud.Disk, error) {
	return c.apiClient.Disk.Read(id)
}

func (c *client) FindServerPlan(core int, memoryGB int) (*sacloud.ProductServer, error) {
	return c.apiClient.Product.Server.GetBySpec(core, memoryGB, sacloud.PlanG2)
}

func (c *client) Shutdown(id int64) error {
	if _, err := c.apiClient.Server.Shutdown(id); err != nil {
		return err
	}
	return c.apiClient.Server.SleepUntilDown(id, c.apiClient.DefaultTimeoutDuration)
}

func (c *client) DisconnectDisks(serverID int64) error {
	server, err := c.ServerByID(serverID)
	if err != nil {
		return err
	}
	for _, disk := range server.Disks {
		if _, err := c.apiClient.Disk.DisconnectFromServer(disk.ID); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) CloneDisk(id int64) (<-chan interface{}, error) {
	sourceDisk, err := c.DiskByID(id)
	if err != nil {
		return nil, err
	}
	params := c.apiClient.Disk.New()
	params.SetDescription(sourceDisk.Description)
	if sourceDisk.HasIcon() {
		params.SetIconByID(sourceDisk.GetIconID())
	}
	params.Plan = sacloud.NewResource(sourceDisk.GetPlanID())
	params.SetSizeMB(sourceDisk.GetSizeMB())

	// DistantFromは未サポート
	// params.SetDistantFrom(sourceDisk.DistantFrom)

	params.SetName(sourceDisk.Name)
	params.SetTags(sourceDisk.Tags)
	params.SetDiskConnection(sacloud.EDiskConnection(sourceDisk.Connection))
	params.SetSourceDisk(id)

	disk, err := c.apiClient.Disk.Create(params)
	if err != nil {
		return nil, err
	}

	compC, progC, errC := c.apiClient.Disk.AsyncSleepWhileCopying(disk.ID, c.apiClient.DefaultTimeoutDuration)
	progress := make(chan interface{})

	go func() {
		for {
			select {
			case disk := <-compC:
				progress <- disk
				close(progress)
				return
			case disk := <-progC:
				progress <- disk
			case err := <-errC:
				progress <- err
				close(progress)
				return
			}
		}
	}()

	return progress, err
}

func (c *client) ChangePlan(serverID int64, plan *sacloud.ProductServer) (*sacloud.Server, error) {
	return c.apiClient.Server.ChangePlan(serverID, plan)
}

func (c *client) ConnectDisks(serverID int64, diskIDs []int64) error {
	for _, diskID := range diskIDs {
		if _, err := c.apiClient.Disk.ConnectToServer(diskID, serverID); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) Boot(id int64) error {
	if _, err := c.apiClient.Server.Boot(id); err != nil {
		return err
	}
	return c.apiClient.Server.SleepUntilUp(id, c.apiClient.DefaultTimeoutDuration)
}

func (c *client) DeleteDisk(id int64) error {
	_, err := c.apiClient.Disk.Delete(id)
	return err
}
