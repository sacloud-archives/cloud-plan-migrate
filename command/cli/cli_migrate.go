package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sacloud/cloud-plan-migrate/command"
	"github.com/sacloud/cloud-plan-migrate/command/funcs"
	"github.com/sacloud/cloud-plan-migrate/command/params"
	"gopkg.in/urfave/cli.v2"
)

var MigrateCommand *cli.Command

func init() {
	migrateParam := params.NewMigrateMigrateParam()

	cliCommand := &cli.Command{
		Name:  "migrate",
		Usage: "Migrate server/disk plan",
		Action: func(c *cli.Context) error {
			// Set option values
			if c.IsSet("selector") {
				migrateParam.Selector = c.StringSlice("selector")
			}
			if c.IsSet("assumeyes") {
				migrateParam.Assumeyes = c.Bool("assumeyes")
			}
			if c.IsSet("cleanup-disk") {
				migrateParam.CleanupDisk = c.Bool("cleanup-disk")
			}
			if c.IsSet("disable-reboot") {
				migrateParam.DisableReboot = c.Bool("disable-reboot")
			}
			if c.IsSet("id") {
				migrateParam.ID = c.Int64("id")
			}

			if c.NArg() == 0 && len(migrateParam.Selector) == 0 {
				cli.ShowAppHelp(c)
				return nil
			}

			// interactive input when API Keys are empty
			if isTerminal() {
				c := color.New(color.BgMagenta)
				if command.GlobalOption.AccessToken == "" {
					// read input
					var input string
					c.Fprintln(command.GlobalOption.Out, "\nYour API AccessToken is not set")
					fmt.Fprintf(command.GlobalOption.Out, "\t%s: ", "Enter your token")
					fmt.Fscanln(command.GlobalOption.In, &input)
					command.GlobalOption.AccessToken = input
				}
				if command.GlobalOption.AccessTokenSecret == "" {
					// read input
					var input string
					c.Fprintln(command.GlobalOption.Out, "\nYour API AccessTokenSecret is not set")
					fmt.Fprintf(command.GlobalOption.Out, "\t%s: ", "Eneter your secret")
					fmt.Fscanln(command.GlobalOption.In, &input)
					command.GlobalOption.AccessTokenSecret = input
				}
			}

			// Validate global params
			if errors := command.GlobalOption.Validate(false); len(errors) > 0 {
				return command.FlattenErrorsWithPrefix(errors, "GlobalOptions")
			}

			// Validate specific for each command params
			if errors := migrateParam.Validate(); len(errors) > 0 {
				return command.FlattenErrorsWithPrefix(errors, "Options")
			}

			// create command context
			ctx := command.NewContext(c, c.Args().Slice(), migrateParam)

			apiClient := ctx.GetAPIClient().Server
			ids := []int64{}

			if c.NArg() == 0 {

				if len(migrateParam.Selector) == 0 {
					return fmt.Errorf("ID or Name argument or --selector option is required")
				}
				apiClient.Reset().Limit(1000)
				res, err := apiClient.Find()
				if err != nil {
					return fmt.Errorf("Find ID is failed: %s", err)
				}
				for _, v := range res.Servers {
					if hasTags(&v, migrateParam.Selector) {
						ids = append(ids, v.GetID())
					}
				}
				if len(ids) == 0 {
					return fmt.Errorf("Find ID is failed: Not Found[with search param tags=%s]", migrateParam.Selector)
				}

			} else {

				for _, arg := range c.Args().Slice() {

					for _, a := range strings.Split(arg, "\n") {
						idOrName := a
						if id, ok := toSakuraID(idOrName); ok {
							ids = append(ids, id)
						} else {
							apiClient.Reset()
							apiClient.Limit(1000)
							apiClient.SetFilterBy("Name", idOrName)
							res, err := apiClient.Find()
							if err != nil {
								return fmt.Errorf("Find ID is failed: %s", err)
							}
							if res.Count == 0 {
								return fmt.Errorf("Find ID is failed: Not Found[with search param %q]", idOrName)
							}
							for _, v := range res.Servers {
								if len(migrateParam.Selector) == 0 || hasTags(&v, migrateParam.Selector) {
									ids = append(ids, v.GetID())
								}
							}
						}
					}

				}

			}

			ids = command.UniqIDs(ids)
			if len(ids) == 0 {
				return fmt.Errorf("Target resource is not found")
			}
			migrateParam.IDs = ids

			// confirm
			if !migrateParam.Assumeyes {
				if !isTerminal() {
					return fmt.Errorf("When using redirect/pipe, specify --assumeyes(-y) option")
				}
				if !command.ConfirmContinue("migrate", ids...) {
					return nil
				}
			}

			return funcs.MigrateMigrate(ctx, migrateParam)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "token",
				Usage:       "API Token of SakuraCloud",
				EnvVars:     []string{"SAKURACLOUD_ACCESS_TOKEN"},
				DefaultText: "none",
				Destination: &command.GlobalOption.AccessToken,
			},
			&cli.StringFlag{
				Name:        "secret",
				Usage:       "API Secret of SakuraCloud",
				EnvVars:     []string{"SAKURACLOUD_ACCESS_TOKEN_SECRET"},
				DefaultText: "none",
				Destination: &command.GlobalOption.AccessTokenSecret,
			},
			&cli.StringFlag{
				Name:        "zone",
				Usage:       "Target zone of SakuraCloud",
				Value:       command.DefaultZone,
				DefaultText: command.DefaultZone,
				Destination: &command.GlobalOption.Zone,
				Hidden:      true,
			},
			&cli.IntFlag{
				Name:        "timeout",
				Usage:       "Number of timeout minutes for polling functions",
				EnvVars:     []string{"SAKURACLOUD_TIMEOUT"},
				Value:       60 * 24, // 24h
				Destination: &command.GlobalOption.Timeout,
				Hidden:      true,
			},
			&cli.StringFlag{
				Name:        "accept-language",
				Usage:       "Accept-Language Header",
				EnvVars:     []string{"SAKURACLOUD_ACCEPT_LANGUAGE"},
				Destination: &command.GlobalOption.AcceptLanguage,
				Hidden:      true,
			},
			&cli.IntFlag{
				Name:        "retry-max",
				Usage:       "Number of API-Client retries",
				EnvVars:     []string{"SAKURACLOUD_RETRY_MAX"},
				Destination: &command.GlobalOption.RetryMax,
				Value:       10,
				Hidden:      true,
			},
			&cli.Int64Flag{
				Name:        "retry-interval",
				Usage:       "API client retry interval seconds",
				EnvVars:     []string{"SAKURACLOUD_RETRY_INTERVAL"},
				Destination: &command.GlobalOption.RetryIntervalSec,
				Value:       5,
				Hidden:      true,
			},
			&cli.BoolFlag{
				Name:        "no-color",
				Usage:       "Flag of not using ANSI color output",
				EnvVars:     []string{"NO_COLOR"},
				Destination: &command.GlobalOption.NoColor,
				Hidden:      true,
			},
			&cli.StringFlag{
				Name:        "api-root-url",
				EnvVars:     []string{"cloud-plan-migrate_API_ROOT_URL"},
				Destination: &command.GlobalOption.APIRootURL,
				Hidden:      true,
			},
			&cli.StringSliceFlag{
				Name:   "zones",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:        "trace",
				Usage:       "Flag of SakuraCloud debug-mode",
				EnvVars:     []string{"SAKURACLOUD_TRACE_MODE"},
				Destination: &command.GlobalOption.TraceMode,
				Value:       false,
				Hidden:      true,
			},
			&cli.BoolFlag{
				Name:  "cleanup-disk",
				Usage: "If true, delete target disk after migration",
			},
			&cli.BoolFlag{
				Name:  "disable-reboot",
				Usage: "If true, don't boot target server after migration",
			},
			&cli.StringSliceFlag{
				Name:  "selector",
				Usage: "Set target filter by tag",
			},
			&cli.BoolFlag{
				Name:    "assumeyes",
				Aliases: []string{"y"},
				Usage:   "Assume that the answer to any question which would be asked is yes",
			},
			&cli.Int64Flag{
				Name:   "id",
				Usage:  "Set target ID",
				Hidden: true,
			},
		},
	}
	MigrateCommand = cliCommand
}
