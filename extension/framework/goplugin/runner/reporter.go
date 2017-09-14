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
	suites map[string]types.SuiteSummary
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
	reporter.suites[summary.SuiteDescription] = *summary
}

func (reporter *Reporter) Prepare(description string) {
	reporter.suites[description] = types.SuiteSummary{
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
	}
}

func (reporter *Reporter) Report() {
	fmt.Println("----------------------------------------")

	fmt.Println("Failures:")
	fmt.Println()

	stenographer := stenographer.New(true)

	configSuccinct := false
	configFullTrace := true
	configNoisyPendings := false
	configSlowSpecThreshold := float64(2) // secs

	for _, spec := range reporter.specs {
		//if spec.State != types.SpecStatePassed {
		//	fmt.Println(redColor + "FAILED:" + defaultStyle, spec)
		//	fmt.Println()
		//}

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

	fmt.Println("Reports:")
	fmt.Println()

	index := 0

	for _, suite := range reporter.suites {
		index++
		fmt.Printf("[%4d] ", index)
		if suite.SuiteSucceeded {
			fmt.Printf(greenColor+"%-64s"+defaultStyle, suite.SuiteDescription)
		} else {
			fmt.Printf(redColor+"%-64s"+defaultStyle, suite.SuiteDescription)
		}
		fmt.Printf(cyanColor+"total: %-4d%8s"+defaultStyle, suite.NumberOfTotalSpecs, "")
		if suite.SuiteSucceeded {
			fmt.Printf(greenColor+"passed: %-4d%8s"+defaultStyle, suite.NumberOfPassedSpecs, "")
		} else {
			fmt.Printf("passed: %-4d%8s", suite.NumberOfPassedSpecs, "")
		}
		if !suite.SuiteSucceeded {
			fmt.Printf(redColor+"failed: %-4d%8s"+defaultStyle, suite.NumberOfFailedSpecs, "")
		} else {
			fmt.Printf("failed: %-4d%8s", suite.NumberOfFailedSpecs, "")
		}
		fmt.Printf("skipped: %-4d%8s", suite.NumberOfSkippedSpecs, "")
		fmt.Printf("pending: %-4d%8s", suite.NumberOfPendingSpecs, "")
		fmt.Printf("run time: %s\n", suite.RunTime)
	}
	fmt.Println()
}

func NewReporter() *Reporter {
	return &Reporter{
		suites: map[string]types.SuiteSummary{},
	}
}
