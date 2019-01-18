package migrate

import (
	"testing"
	"time"

	"github.com/sacloud/libsacloud/sacloud"
	"github.com/stretchr/testify/assert"
)

var (
	serverID      = int64(1)
	currentDiskID = int64(2)
	emptyID       = int64(0)
)

func singleDiskServer() *sacloud.Server {
	server := &sacloud.Server{Resource: sacloud.NewResource(serverID)}
	disk := sacloud.Disk{Resource: sacloud.NewResource(currentDiskID)}
	disk.SizeMB = 20 * 1024
	server.Disks = []sacloud.Disk{disk}
	server.Instance = &sacloud.Instance{
		EServerInstanceStatus: &sacloud.EServerInstanceStatus{
			Status: "up",
		},
	}
	return server
}

func multipleDiskServer() *sacloud.Server {
	server := &sacloud.Server{Resource: sacloud.NewResource(serverID)}
	disk := sacloud.Disk{Resource: sacloud.NewResource(currentDiskID)}
	disk.SizeMB = 20 * 1024
	server.Disks = []sacloud.Disk{disk, disk}
	server.Instance = &sacloud.Instance{
		EServerInstanceStatus: &sacloud.EServerInstanceStatus{
			Status: "up",
		},
	}
	return server
}

type fakeClient struct {
	server *sacloud.Server
}

func (f *fakeClient) FindAll() ([]*sacloud.Server, error) {
	// not implements
	return nil, nil
}
func (f *fakeClient) ServerByID(id int64) (*sacloud.Server, error) {
	return f.server, nil
}
func (f *fakeClient) DiskByID(id int64) (*sacloud.Disk, error) {
	return nil, nil
}
func (f *fakeClient) FindServerPlan(core int, memoryGB int) (*sacloud.ProductServer, error) {
	return nil, nil
}
func (f *fakeClient) Shutdown(id int64) (err error) {
	return nil
}
func (f *fakeClient) DisconnectDisks(serverID int64) error {
	return nil
}
func (f *fakeClient) CloneDisk(id int64) (<-chan interface{}, error) {
	return nil, nil
}
func (f *fakeClient) ChangePlan(serverID int64, plan *sacloud.ProductServer) (*sacloud.Server, error) {
	return nil, nil
}
func (f *fakeClient) ConnectDisks(serverID int64, diskIDs []int64) error {
	return nil
}
func (f *fakeClient) Boot(id int64) error {
	return nil
}
func (f *fakeClient) DeleteDisk(id int64) error {
	return nil
}

func TestMigration_NewMigration(t *testing.T) {

	t.Run("constructor", func(t *testing.T) {

		fakeClient := &fakeClient{
			server: singleDiskServer(),
		}

		migration, err := NewMigration(fakeClient, []int64{serverID}, &Options{})
		assert.NotNil(t, migration)
		assert.NoError(t, err)

		assert.Len(t, migration.status, 1)
		status := migration.status[0]

		// check status
		assert.Equal(t, serverID, status.targetServerID)
		assert.Len(t, status.Disks, 1)
		diskStatus := status.Disks[0]

		assert.Equal(t, currentDiskID, diskStatus.originalID)
		assert.Equal(t, emptyID, diskStatus.clonedID)

		assert.NotNil(t, status.stepShutdown)
		assert.True(t, status.stepShutdown.needProcess)

		assert.NotNil(t, status.stepDisconnectDisks)
		assert.True(t, status.stepDisconnectDisks.needProcess)
		assert.NotNil(t, status.stepConnectDisks)
		assert.True(t, status.stepConnectDisks.needProcess)

		assert.NotNil(t, diskStatus.stepClone)
		assert.True(t, diskStatus.stepClone.needProcess)

		assert.NotNil(t, status.stepPlanMigrate)
		assert.True(t, status.stepPlanMigrate.needProcess)

		assert.NotNil(t, status.stepBoot)
		assert.True(t, status.stepBoot.needProcess)

		assert.NotNil(t, diskStatus.stepDelete.needProcess)
		assert.False(t, diskStatus.stepDelete.needProcess)
	})

	t.Run("with disks", func(t *testing.T) {

		expects := []struct {
			expectDisksLen int
			server         *sacloud.Server
		}{
			{
				expectDisksLen: 1,
				server:         singleDiskServer(),
			},
			{
				expectDisksLen: 2,
				server:         multipleDiskServer(),
			},
		}

		for _, expect := range expects {

			fakeClient := &fakeClient{
				server: expect.server,
			}

			migration, err := NewMigration(fakeClient, []int64{serverID}, &Options{})
			assert.NotNil(t, migration)
			assert.NoError(t, err)

			assert.Len(t, migration.status, 1)
			status := migration.status[0]

			assert.Len(t, status.Disks, expect.expectDisksLen)
		}
	})

	t.Run("with option", func(t *testing.T) {

		fakeClient := &fakeClient{
			server: singleDiskServer(),
		}

		migration, err := NewMigration(fakeClient, []int64{serverID}, &Options{
			DisableBoot: true,
			DeleteDisks: true,
		})
		assert.NoError(t, err)

		status := migration.status[0]
		assert.False(t, status.stepBoot.needProcess)
		assert.True(t, status.Disks[0].stepDelete.needProcess)
	})
}

func TestMigration_handleSteps(t *testing.T) {

	fakeClient := &fakeClient{
		server: singleDiskServer(),
	}
	migration, _ := NewMigration(fakeClient, []int64{serverID}, &Options{})

	called := false

	status := &ServerStatus{}
	step1 := &step{}
	step2 := &step{}
	steps := []*step{step1, step2}

	err := migration.handleSteps(func(status *ServerStatus) error {
		called = true
		time.Sleep(10 * time.Millisecond)
		return nil
	}, status, steps...)

	assert.NoError(t, err)
	assert.True(t, called)

	assert.True(t, step1.started)
	assert.True(t, step2.started)
	assert.True(t, step1.done)
	assert.True(t, step2.done)

	assert.True(t, step1.elapsed() > 10*time.Millisecond)
	assert.True(t, step2.elapsed() > 10*time.Millisecond)

}
