package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	migrateCLI "github.com/sacloud/cloud-plan-migrate/command/cli"
	"github.com/sacloud/cloud-plan-migrate/version"
	"gopkg.in/urfave/cli.v2"
)

var (
	appName      = "cloud-plan-migrate"
	appUsage     = "Command line tool for migrating server/disk resource-plan to gen2"
	appCopyright = "Copyright (C) 2019 Kazumichi Yamamoto."
)

func main() {

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sigChan
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(color.Error, color.HiBlueString("signal received; shutting down"))
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
	}()

	app := &cli.App{
		Name:      appName,
		Usage:     appUsage,
		HelpName:  appName,
		Copyright: appCopyright,
		Version:   version.FullVersion(),
		Flags:     migrateCLI.MigrateCommand.Flags,
		Action:    migrateCLI.MigrateCommand.Action,
	}

	cli.AppHelpTemplate = helpTemplate

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

var helpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[options] <Resource ID or Name...>{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}

OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}

COPYRIGHT:
   {{.Copyright}}
`
