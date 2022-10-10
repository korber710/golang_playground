package main

import (
	"log"
	"os"
	"time"

	"github.com/korber710/golang_playground/robot-trigger/pkg/run"
	"github.com/urfave/cli/v2"
)

func main() {
	var verbosePrints bool

	app := &cli.App{
		Name:     "robot-trigger",
		Version:  "v0.0.1",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Steve Korber",
				Email: "korbersa@outlook.com",
			},
		},
		Copyright: "(c) 2022 Korber Solutions",
		Usage:     "trigger robot framework from go!",
		Commands: []*cli.Command{
			&cli.Command{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Executes a test suite in robot",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Destination: &verbosePrints},
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, DefaultText: "./test/test1.robot"},
				},
				Action: func(cCtx *cli.Context) error {
					robotFile := cCtx.String("file")
					err := run.Run(verbosePrints, robotFile)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
