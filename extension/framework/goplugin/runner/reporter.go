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

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
	"time"
	"github.com/onsi/ginkgo/reporters/stenographer"
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
		SuiteSucceeded: false,
		SuiteID: "undefined",
		NumberOfSpecsBeforeParallelization: 0,
		NumberOfTotalSpecs: 0,
		NumberOfSpecsThatWillBeRun: 0,
		NumberOfPendingSpecs: 0,
		NumberOfSkippedSpecs: 0,
		NumberOfPassedSpecs: 0,
		NumberOfFailedSpecs: 0,
		RunTime: time.Duration(0),
	}
}

func (reporter *Reporter) Report() {
	fmt.Println("----------------------------------------")
	fmt.Println("Reports:")
	fmt.Println()

	for _, suite := range reporter.suites {
		fmt.Println("Suite description:", suite.SuiteDescription)
		fmt.Println("Suite succeeded:", suite.SuiteSucceeded)
		fmt.Println("Suite ID:", suite.SuiteID)
		fmt.Println("Number of specs before parallelization:", suite.NumberOfSpecsBeforeParallelization)
		fmt.Println("Number of total specs:", suite.NumberOfTotalSpecs)
		fmt.Println("Number of specs that will be run:", suite.NumberOfSpecsThatWillBeRun)
		fmt.Println("Number of pending specs:", suite.NumberOfPendingSpecs)
		fmt.Println("Number of skipped specs:", suite.NumberOfSkippedSpecs)
		fmt.Println("Number of passed specs:", suite.NumberOfPassedSpecs)
		fmt.Println("Number of failed specs:", suite.NumberOfFailedSpecs)
		fmt.Println("Run time:", suite.RunTime)
		fmt.Println()
	}
	fmt.Println()

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
}

func NewReporter() *Reporter {
	return &Reporter{
		suites: map[string]types.SuiteSummary{},
	}
}
