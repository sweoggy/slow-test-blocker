package main

import (
	"log"
	"os"
	"time"

	"github.com/sweoggy/slow-test-blocker/src/detect"
	"github.com/urfave/cli/v2"
)

func main() {
	analyzeFlags := getAnalyzeFlags()

	app := &cli.App{
		Name:  "slow-test-blocker",
		Usage: "Utility for blocking slow test entering your code base",
		Commands: []*cli.Command{
			{
				Name:      "analyze",
				Usage:     "Analyze a JUnit report for slow tests",
				ArgsUsage: "[report.xml]",
				Action: func(cliContext *cli.Context) error {
					return detect.DetectTooSlow(cliContext, false)
				},
				Flags: analyzeFlags,
			},
			{
				Name:      "generate-baseline",
				Usage:     "Generate a baseline for tests ",
				ArgsUsage: "[report.xml]",
				Action: func(cliContext *cli.Context) error {
					return detect.DetectTooSlow(cliContext, true)
				},
				Flags: analyzeFlags,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getAnalyzeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.DurationFlag{Name: "max-duration", Usage: "Max duration a test is allowed to take before returning failure exit code", Value: time.Millisecond * 500},
		&cli.StringFlag{Name: "baseline-path", Usage: "Path to baseline path from current working directory", Value: "slow-test-blocker.baseline.json"},
	}
}
