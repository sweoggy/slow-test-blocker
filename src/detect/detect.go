package detect

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/urfave/cli/v2"
)

type BaselineEntries struct {
	BaselineEntry []BaselineEntry `json:"baseline_entries"`
}

type BaselineEntry struct {
	FullTestName string `json:"full_test_name"`
	MaxDuraction int64  `json:"max_duration"`
	TestTime     int64  `json:"test_time"`
}

func DetectTooSlowUsingArgumentAndFlags(cliContext *cli.Context, generateBaseline bool) error {
	maxDuration := cliContext.Duration("max-duration")
	if maxDuration == 0 {
		return cli.Exit("Max duration was set to 0, no analyze needed", 5)
	}
	baselinePath := cliContext.String("baseline-path")
	if baselinePath == "" {
		return cli.Exit("You must specify a non-empty path to base line file", 5)
	}

	reportFile, err := os.ReadFile(cliContext.Args().First())
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read JUnit report file: %s", err), 2)
	}

	reportContent, err := xmlquery.Parse(strings.NewReader(string(reportFile)))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to parse report file as XML %s", err), 3)
	}

	return DetectTooSlow(reportContent, maxDuration, generateBaseline, baselinePath)
}

func DetectTooSlow(reportContent *xmlquery.Node, maxDuration time.Duration, generateBaseline bool, baselinePath string) error {
	var testsToIgnore BaselineEntries
	if !generateBaseline {
		baselineFile, err := os.ReadFile(baselinePath)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to read baseline file: %s", err), 2)
		}

		json.Unmarshal(baselineFile, &testsToIgnore)
	}

	testCases, err := xmlquery.QueryAll(reportContent, "//testcase")
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to query test cases %s", err), 4)
	}

	fmt.Println(fmt.Sprintf("Found %d test cases to analyze. Starting analysis...\n", len(testCases)))
	var tooSlowTestDetected = false
	var baseLineEntrySlice []BaselineEntry
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
			if generateBaseline {
				fmt.Println(fmt.Sprintf("Detected too slow test %s, adding to baseline. Test time was: %dms", fullTestName, testTime.Milliseconds()))
				baseLineEntrySlice = append(baseLineEntrySlice, BaselineEntry{FullTestName: fullTestName, MaxDuraction: maxDuration.Milliseconds(), TestTime: testTime.Milliseconds()})
			} else {
				var isTestIgnored = false
				for _, baselineEntry := range testsToIgnore.BaselineEntry {
					if baselineEntry.FullTestName == fullTestName {
						isTestIgnored = true
						break
					}
				}
				if !isTestIgnored {
					fmt.Println(fmt.Sprintf("%s test was too slow! Test time was: %dms", fullTestName, testTime.Milliseconds()))
					tooSlowTestDetected = true
				} else {
					fmt.Println(fmt.Sprintf("%s test is too slow (%dms), but in baseline, ignoring", fullTestName, testTime.Milliseconds()))
				}
			}
		}
	}

	if generateBaseline {
		var newBaselineEntries = BaselineEntries{BaselineEntry: baseLineEntrySlice}
		newBaselineEntriesEncoded, err := json.Marshal(newBaselineEntries)
		if err != nil {
			return cli.Exit(fmt.Sprintf("\nFailed to encoded new baseline entries: %s", err), 1)
		}

		if os.WriteFile(baselinePath, newBaselineEntriesEncoded, 0744) != nil {
			return cli.Exit(fmt.Sprintf("\nFailed to save new baseline file: %s", err), 1)
		}

		fmt.Println(fmt.Sprintf("Successfully generate baseline file, see output above for which slow tests (Exceeding %dms) were added to file", maxDuration.Milliseconds()))
	} else {
		if tooSlowTestDetected {
			return cli.Exit(fmt.Sprintf("\nAt least one test is too slow (Exceeding %dms), see output above for details", maxDuration.Milliseconds()), 1)
		}
		fmt.Println("Successfully analyzed all test cases, all tests are within limits")
	}

	return nil
}
