// Copyright (C) 2017 NTT Innovation Institute, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runner

import (
	"fmt"

	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters/stenographer"
	"github.com/onsi/ginkgo/types"
)

const defaultStyle = "\x1b[0m"
const boldStyle = "\x1b[1m"
const redColor = "\x1b[91m"
const greenColor = "\x1b[32m"
const yellowColor = "\x1b[33m"
const cyanColor = "\x1b[36m"
const grayColor = "\x1b[90m"
const lightGrayColor = "\x1b[37m"

type Reporter struct {
	suites []types.SuiteSummary
	specs  []types.SpecSummary
}

func (reporter *Reporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (reporter *Reporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *Reporter) SpecWillRun(specSummary *types.SpecSummary) {
}

func (reporter *Reporter) SpecDidComplete(specSummary *types.SpecSummary) {
	reporter.specs = append(reporter.specs, *specSummary)
}

func (reporter *Reporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *Reporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	for i, _ := range reporter.suites {
		if reporter.suites[i].SuiteDescription == summary.SuiteDescription {
			reporter.suites[i] = *summary
			break
		}
	}
}

func (reporter *Reporter) Prepare(description string) {
	reporter.suites = append(reporter.suites, types.SuiteSummary{
		SuiteDescription: description,
		SuiteSucceeded:   false,
		SuiteID:          "undefined",
		NumberOfSpecsBeforeParallelization: 0,
		NumberOfTotalSpecs:                 0,
		NumberOfSpecsThatWillBeRun:         0,
		NumberOfPendingSpecs:               0,
		NumberOfSkippedSpecs:               0,
		NumberOfPassedSpecs:                0,
		NumberOfFailedSpecs:                0,
		RunTime:                            time.Duration(0),
	})
}

func (reporter *Reporter) Report() {
	fmt.Println("--------------------------------------------------------------------------------")

	fmt.Println("Failures:")
	fmt.Println()

	stenographer := stenographer.New(true)

	configSuccinct := false
	configFullTrace := true
	configNoisyPendings := false
	configSlowSpecThreshold := float64(2) // secs

	for _, spec := range reporter.specs {
		stenographer.AnnounceCapturedOutput(spec.CapturedOutput)

		switch spec.State {
		case types.SpecStatePassed:
			if spec.IsMeasurement {
				stenographer.AnnounceSuccesfulMeasurement(&spec, configSuccinct)
			} else if spec.RunTime.Seconds() >= configSlowSpecThreshold {
				stenographer.AnnounceSuccesfulSlowSpec(&spec, configSuccinct)
			} else {
				stenographer.AnnounceSuccesfulSpec(&spec)
			}

		case types.SpecStatePending:
			stenographer.AnnouncePendingSpec(&spec, configNoisyPendings && !configSuccinct)
		case types.SpecStateSkipped:
			stenographer.AnnounceSkippedSpec(&spec, configSuccinct, configFullTrace)
		case types.SpecStateTimedOut:
			stenographer.AnnounceSpecTimedOut(&spec, configSuccinct, configFullTrace)
		case types.SpecStatePanicked:
			stenographer.AnnounceSpecPanicked(&spec, configSuccinct, configFullTrace)
		case types.SpecStateFailed:
			stenographer.AnnounceSpecFailed(&spec, configSuccinct, configFullTrace)
		}
	}

	fmt.Println()
	fmt.Println("Report:")
	fmt.Println()

	// counters
	prevNumberOfTotalSpecs := 0
	prevNumberOfPendingSpecs := 0
	prevNumberOfSkippedSpecs := 0
	prevNumberOfPassedSpecs := 0
	prevNumberOfFailedSpecs := 0

	numberOfTotalSpecs := 0
	numberOfPendingSpecs := 0
	numberOfSkippedSpecs := 0
	numberOfPassedSpecs := 0
	numberOfFailedSpecs := 0
	runTime := time.Duration(0) * time.Nanosecond

	totalNumberOfTotalSpecs := 0
	totalNumberOfPendingSpecs := 0
	totalNumberOfSkippedSpecs := 0
	totalNumberOfPassedSpecs := 0
	totalNumberOfFailedSpecs := 0
	totalRunTime := time.Duration(0) * time.Nanosecond

	description := ""
	succeeded := false

	for index, suite := range reporter.suites {
		numberOfTotalSpecs = suite.NumberOfTotalSpecs - prevNumberOfTotalSpecs
		numberOfPendingSpecs = suite.NumberOfPendingSpecs - prevNumberOfPendingSpecs
		numberOfSkippedSpecs = suite.NumberOfSkippedSpecs - prevNumberOfSkippedSpecs
		numberOfPassedSpecs = suite.NumberOfPassedSpecs - prevNumberOfPassedSpecs
		numberOfFailedSpecs = suite.NumberOfFailedSpecs - prevNumberOfFailedSpecs
		runTime = suite.RunTime

		prevNumberOfTotalSpecs = suite.NumberOfTotalSpecs
		prevNumberOfPendingSpecs = suite.NumberOfPendingSpecs
		prevNumberOfSkippedSpecs = suite.NumberOfSkippedSpecs
		prevNumberOfPassedSpecs = suite.NumberOfPassedSpecs
		prevNumberOfFailedSpecs = suite.NumberOfFailedSpecs

		totalNumberOfTotalSpecs += numberOfTotalSpecs
		totalNumberOfPendingSpecs += numberOfPendingSpecs
		totalNumberOfSkippedSpecs += numberOfSkippedSpecs
		totalNumberOfPassedSpecs += numberOfPassedSpecs
		totalNumberOfFailedSpecs += numberOfFailedSpecs
		totalRunTime += runTime

		description = suite.SuiteDescription
		succeeded = suite.SuiteSucceeded

		fmt.Printf("[%4d] ", index+1)
		if succeeded {
			fmt.Printf(greenColor+"%-80s"+defaultStyle, description)
		} else {
			fmt.Printf(redColor+"%-80s"+defaultStyle, description)
		}
		fmt.Printf(cyanColor+"total: %-4d%8s"+defaultStyle, numberOfTotalSpecs, "")
		if succeeded {
			fmt.Printf(greenColor+"passed: %-4d%8s"+defaultStyle, numberOfPassedSpecs, "")
		} else {
			fmt.Printf("passed: %-4d%8s", numberOfPassedSpecs, "")
		}
		if !succeeded {
			fmt.Printf(redColor+"failed: %-4d%8s"+defaultStyle, numberOfFailedSpecs, "")
		} else {
			fmt.Printf("failed: %-4d%8s", numberOfFailedSpecs, "")
		}
		if numberOfSkippedSpecs > 0 {
			fmt.Printf(yellowColor+"skipped: %-4d%8s"+defaultStyle, numberOfSkippedSpecs, "")
		} else {
			fmt.Printf("skipped: %-4d%8s", numberOfSkippedSpecs, "")
		}
		if numberOfPendingSpecs > 0 {
			fmt.Printf(yellowColor + "pending: %-4d%8s" + defaultStyle, numberOfPendingSpecs, "")
		} else {
			fmt.Printf("pending: %-4d%8s", numberOfPendingSpecs, "")
		}
		fmt.Printf("run time: %s\n", runTime)
	}

	fmt.Println()

	fmt.Printf("[----] ")
	fmt.Printf(yellowColor+"%-80s"+defaultStyle, "SUMMARY")
	fmt.Printf(cyanColor+"total: %-4d%8s"+defaultStyle, totalNumberOfTotalSpecs, "")
	fmt.Printf("passed: %-4d%8s", totalNumberOfPassedSpecs, "")
	fmt.Printf("failed: %-4d%8s", totalNumberOfFailedSpecs, "")
	fmt.Printf("skipped: %-4d%8s", totalNumberOfSkippedSpecs, "")
	fmt.Printf("pending: %-4d%8s", totalNumberOfPendingSpecs, "")
	fmt.Printf("run time: %s\n", totalRunTime)
}

func NewReporter() *Reporter {
	return &Reporter{
		suites: []types.SuiteSummary{},
	}
}
