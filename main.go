package main

import (
	"fmt"
	"os"

	"github.com/Boyuan-Chen/v3-migration/config"
	"github.com/Boyuan-Chen/v3-migration/flags"
	"github.com/Boyuan-Chen/v3-migration/migration"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/urfave/cli"
)

var (
	GitVersion = ""
	GitCommit  = ""
	GitDate    = ""
)

func main() {
	app := cli.NewApp()
	app.Flags = flags.Flags

	app.Version = GitVersion + "-" + params.VersionWithCommit(GitCommit, GitDate)
	app.Name = "boba-v3-migration"
	app.Usage = "Migrate legacy Boba block chain to erigon client"
	app.Description = "Configure with endpoints to send proposers to the execution engine"

	// Define the functionality of the application
	app.Action = func(ctx *cli.Context) error {
		if args := ctx.Args(); len(args) > 0 {
			return fmt.Errorf("invalid command: %q", args[0])
		}

		config := config.NewConfig(ctx)
		m, err := migration.NewMigration(config)
		if err != nil {
			return err
		}

		if err := m.Start(); err != nil {
			return err
		}

		m.Wait()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("application failed", "message", err)
	}
}
