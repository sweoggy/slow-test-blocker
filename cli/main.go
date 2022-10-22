package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "slow-test-blocker",
		Usage: "Utility for blocking slow test entering code base",
		Commands: []*cli.Command{
			{
				Name:      "analyze",
				Usage:     "Analyze a JUnit report for slow tests",
				ArgsUsage: "[report.xml]",
				Action: func(cliContext *cli.Context) error {
					maxDuration := cliContext.Duration("max-duration")
					if maxDuration == 0 {
						return cli.Exit("Max duration was set to 0, no analyze needed", 5)
					}

					reportFile, err := os.ReadFile(cliContext.Args().First())
					if err != nil {
						return cli.Exit(fmt.Sprintf("Failed to read JUnit report file %s", err), 2)
					}

					reportContent, err := xmlquery.Parse(strings.NewReader(string(reportFile)))
					if err != nil {
						return cli.Exit(fmt.Sprintf("Failed to parse report file as XML %s", err), 3)
					}

					testCases, err := xmlquery.QueryAll(reportContent, "//testcase")
					if err != nil {
						return cli.Exit(fmt.Sprintf("Failed to query test cases %s", err), 4)
					}

					fmt.Println(fmt.Sprintf("Found %d test cases to analyze. Starting analysis...\n", len(testCases)))
					var tooSlowTestDetected = false
					for _, testCase := range testCases {
						testClass := testCase.SelectAttr("class")
						// Some test suites, such as Cypress, does not have the class attribute, fallback to classname
						var testName string
						var fullTestName string
						if testClass == "" {
							testName = testCase.SelectAttr("classname")
							fullTestName = testName
						} else {
							testName = testCase.SelectAttr("name")
							fullTestName = testClass + "::" + testName
						}
						testTime, err := time.ParseDuration(fmt.Sprintf("%ss", testCase.SelectAttr("time")))

						if err != nil {
							fmt.Println(fmt.Sprintf("Failed to parser duration for %s, error %s", fullTestName, err))

							continue
						}

						if testTime > maxDuration {
							tooSlowTestDetected = true
							fmt.Println(fmt.Sprintf("%s test was too slow! Test time was: %dms", fullTestName, testTime.Milliseconds()))
						}
					}

					if tooSlowTestDetected == true {
						return cli.Exit(fmt.Sprintf("\nAt least one test is too slow, see output above for details"), 1)
					}

					fmt.Println("Successfully analyzed all test cases, all tests are within limits")

					return nil
				},
				Flags: []cli.Flag{
					&cli.DurationFlag{Name: "max-duration", Usage: "Max duration a test is allowed to take before returning failure exit code", Value: time.Millisecond * 500},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
