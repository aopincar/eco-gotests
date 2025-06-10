package tsparams

import (
	amdgpuv1 "github.com/ROCm/gpu-operator/api/v1alpha1"
	"github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/internal/amdgpuparams"
	"github.com/openshift-kni/eco-gotests/tests/hw-accel/internal/hwaccelparams"
	"github.com/openshift-kni/k8sreporter"
)

var (
	// Labels represents the range of labels that can be used for test cases selection.
	Labels = append(amdgpuparams.Labels, LabelSuite)

	// ReporterNamespacesToDump tells to the reporter from where to collect logs.
	ReporterNamespacesToDump = map[string]string{
		hwaccelparams.NFDNamespace: "nfd-operator",
		"amd-gpu-operator":         "openshift-amd-gpu",
	}

	// ReporterCRDsToDump tells to the reporter what CRs to dump.
	ReporterCRDsToDump = []k8sreporter.CRData{
		{Cr: &amdgpuv1.DeviceConfigList{}},
	}
)
