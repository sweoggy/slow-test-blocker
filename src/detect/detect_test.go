package detect

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/antchfx/xmlquery"
)

const baselineFilename = "slow-test-blocker.test.baseline.json"

func TestDetectTooSlowCypressTest(t *testing.T) {
	t.Cleanup(cleanup)

	reportFile, _ := os.ReadFile("../../fixtures/cypressresults.xml")

	reportContent, _ := xmlquery.Parse(strings.NewReader(string(reportFile)))

	err := DetectTooSlow(reportContent, time.Millisecond*500, false, baselineFilename)

	if err == nil {
		t.Fatal("Expected error as some tests are too slow")
	}

	err = DetectTooSlow(reportContent, time.Millisecond*500, true, baselineFilename)

	if err != nil {
		t.Fatalf("No error expected when generating baseline %s", err)
	}

	err = DetectTooSlow(reportContent, time.Millisecond*500, false, baselineFilename)

	if err != nil {
		t.Fatalf("No error expected when having a baseline %s", err)
	}
}

func TestDetectTooSlowPhpTest(t *testing.T) {
	t.Cleanup(cleanup)

	reportFile, _ := os.ReadFile("../../fixtures/php-test-results.xml")

	reportContent, _ := xmlquery.Parse(strings.NewReader(string(reportFile)))

	err := DetectTooSlow(reportContent, time.Millisecond*567496, false, baselineFilename)

	if err != nil {
		t.Fatalf("No error expected as the slowest test is equal to threshold. Error: %s", err)
	}
}

func cleanup() {
	_ = os.Remove(baselineFilename)
	_ = os.WriteFile(baselineFilename, nil, 0744)
}
