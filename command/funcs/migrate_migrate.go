package funcs

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/sacloud/cloud-plan-migrate/command"
	"github.com/sacloud/cloud-plan-migrate/command/params"
	"github.com/sacloud/cloud-plan-migrate/iaas"
	"github.com/sacloud/cloud-plan-migrate/migrate"
	"github.com/sacloud/libsacloud/sacloud"
)

func MigrateMigrate(ctx command.Context, params *params.MigrateMigrateParam) error {

	rawClient := ctx.GetAPIClient()
	client := iaas.NewClient(rawClient)

	// validate server status
	for _, serverID := range params.IDs {
		server, err := client.ServerByID(serverID)
		if err != nil {
			return fmt.Errorf("Migrate is failed: %s", err)
		}

		if len(server.Disks) == 0 {
			return fmt.Errorf("Server[%q] don't have any disks", serverID)
		}

		if server.GetServerPlan().Generation != sacloud.PlanG1 {
			return fmt.Errorf("Server[%q] is already use plan-gen2", serverID)
		}
	}

	// prepare params
	logfile, err := openLogFile()
	if err != nil {
		return fmt.Errorf("Migrate is failed: %s", err)
	}
	defer logfile.Close()

	logger := log.New(logfile, "", log.LstdFlags)
	options := &migrate.Options{
		DisableBoot:    params.DisableReboot,
		DeleteDisks:    params.CleanupDisk,
		MaxWorkerCount: 10, // TODO 設定変更可能に
		Logger:         logger,
	}

	// prepare migration
	migration, err := migrate.NewMigration(client, params.IDs, options)
	if err != nil {
		return fmt.Errorf("Migrate is failed: %s", err)
	}

	// exec migration
	var doneC = make(chan bool)
	tickC := time.NewTicker(time.Second).C

	go func() {
		migration.Apply()
		doneC <- true
	}()

	// wait and printing
	for {
		select {
		case <-tickC:
			outputMigrationStatus(migration.Working())
			outputMigrationErrors(migration.HasErrors())
		case <-doneC:
			outputMigrationStatus(migration.Working())

			fmt.Fprintln(command.GlobalOption.Out, "")

			c := color.New(color.FgHiGreen)
			c.Fprintln(command.GlobalOption.Out, "=== Migration finished ===")

			fmt.Fprintln(command.GlobalOption.Out, "")
			outputMigrationErrors(migration.HasErrors())
			return nil
		}
	}
}

func openLogFile() (*os.File, error) {
	name := fmt.Sprintf("migrate-%s.log", time.Now().Format("20060102-150405"))
	return os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
}

var out = bufio.NewWriter(command.GlobalOption.Out)
var screen = new(bytes.Buffer)

func outputMigrationStatus(status []*migrate.ServerStatus) {

	out.WriteString("\033[1;1H") // position(line3-1)
	out.WriteString("\033[0J")   // clear after cursor
	screen.Reset()

	if len(status) > 0 {
		table := tablewriter.NewWriter(screen)
		table.SetHeader([]string{"Server", "Shutdown", "Disk", "PlanChange", "Boot", "Cleanup"})
		//table.SetAutoMergeCells(true)
		//table.SetRowLine(true)
		table.SetAutoFormatHeaders(false)
		table.SetColMinWidth(1, 12)
		table.SetColMinWidth(2, 24)
		table.SetColMinWidth(3, 10)
		table.SetColMinWidth(4, 12)
		table.SetColMinWidth(5, 10)

		for _, s := range status {
			data := buildOutputDataFromStatus(s)
			table.AppendBulk(data)
		}

		table.Render() // write to buf
		fmt.Fprintln(screen, "")
	}

	out.WriteString(screen.String())
	out.Flush()
}

func buildOutputDataFromStatus(s *migrate.ServerStatus) [][]string {
	var data [][]string

	for _, d := range s.Disks {
		data = append(data, []string{
			s.ServerID(),
			s.ShutdownStatus(),
			d.CloneStatus(),
			s.MigrationStatus(),
			s.BootStatus(),
			d.DeleteStatus(),
		})
	}

	return data
}

func outputMigrationErrors(errs []*migrate.ServerStatus) {
	if len(errs) == 0 {
		return
	}

	screen.Reset()

	cTitle := color.New(color.BgRed)
	cBody := color.New(color.FgRed)
	cTitle.Fprintln(screen, "*** Errors ***")
	for _, e := range errs {
		cBody.Fprintf(screen, "  Server[%s:%s] Error: %s\n", e.ServerID(), e.ServerName(), e.Err)
	}

	out.WriteString(screen.String())
	out.Flush()
}
