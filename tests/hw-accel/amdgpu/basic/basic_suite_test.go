package amdgpu

import (
	"github.com/openshift-kni/eco-goinfra/pkg/reportxml"
	"github.com/openshift-kni/eco-gotests/tests/internal/reporter"
	"runtime"
	"testing"

	"github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/basic/internal/tsparams"
	_ "github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/basic/tests"
	. "github.com/openshift-kni/eco-gotests/tests/internal/inittools"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _, currentFile, _, _ = runtime.Caller(0)

func TestBasic(t *testing.T) {
	_, reporterConfig := GinkgoConfiguration()
	reporterConfig.JUnitReport = GeneralConfig.GetJunitReportPath(currentFile)

	RegisterFailHandler(Fail)
	RunSpecs(t, "AMD GPU Basic Suite", Label(tsparams.Labels...), reporterConfig)
}

var _ = ReportAfterSuite("", func(report Report) {
	reportxml.Create(
		report, GeneralConfig.GetReportPath(), GeneralConfig.TCPrefix)
})

var _ = JustAfterEach(func() {
	reporter.ReportIfFailed(
		CurrentSpecReport(), currentFile, tsparams.ReporterNamespacesToDump, tsparams.ReporterCRDsToDump)

})
