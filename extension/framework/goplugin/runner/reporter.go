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
)

type Reporter struct {
	summaries map[string]types.SuiteSummary
}

func (reporter *Reporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (reporter *Reporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {

}

func (reporter *Reporter) SpecWillRun(specSummary *types.SpecSummary) {

}

func (reporter *Reporter) SpecDidComplete(specSummary *types.SpecSummary) {

}

func (reporter *Reporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {

}

func (reporter *Reporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	reporter.summaries[summary.SuiteDescription] = *summary
}

func (reporter *Reporter) Prepare(description string) {
	reporter.summaries[description] = types.SuiteSummary{
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
	fmt.Println("Testing report")
	fmt.Println()

	for _, summary := range reporter.summaries {
		fmt.Println("Suite description:", summary.SuiteDescription)
		fmt.Println("Suite succeeded:", summary.SuiteSucceeded)
		fmt.Println("Suite ID:", summary.SuiteID)
		fmt.Println("Number of specs before parallelization:", summary.NumberOfSpecsBeforeParallelization)
		fmt.Println("Number of total specs:", summary.NumberOfTotalSpecs)
		fmt.Println("Number of specs that will be run:", summary.NumberOfSpecsThatWillBeRun)
		fmt.Println("Number of pending specs:", summary.NumberOfPendingSpecs)
		fmt.Println("Number of skipped specs:", summary.NumberOfSkippedSpecs)
		fmt.Println("Number of passed specs:", summary.NumberOfPassedSpecs)
		fmt.Println("Number of failed specs:", summary.NumberOfFailedSpecs)
		fmt.Println("Run time:", summary.RunTime)
		fmt.Println()
	}
}

func NewReporter() *Reporter {
	return &Reporter{
		summaries: map[string]types.SuiteSummary{},
	}
}
