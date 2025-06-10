package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-kni/eco-goinfra/pkg/nodes"
	"github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/basic/internal/helpers"
	"github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/basic/internal/tsparams"
	"github.com/openshift-kni/eco-gotests/tests/internal/inittools"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	amdLabel      = "feature.node.kubernetes.io/amd-gpu"
	amdLabelValue = "true"
)

var _ = Describe("AMD GPU Basic Tests", Ordered, Label(tsparams.LabelSuite), func() {

	Context("AMD GPU Basic 01", Label(tsparams.LabelSuite+"-01"), func() {

		apiClient := inittools.APIClient
		workerLabelMap := inittools.GeneralConfig.WorkerLabelMap

		BeforeAll(func() {

			// Returns an inventory of worker nodes
			nodeBuilder, builderErr := nodes.List(
				apiClient,
				metav1.ListOptions{LabelSelector: labels.Set(workerLabelMap).String()})

			if builderErr != nil {
				Skip("Failed to get list of worker nodes")
			}

			if len(nodeBuilder) == 0 {
				Skip("Worker nodes list is empty")
			}

		})

		BeforeEach(func() {})

		AfterEach(func() {})

		AfterAll(func() {})

		It("Check AMD label was added by NFD", func() {

			amdLabelFound, err := helpers.AllNodeLabel(apiClient, amdLabel, amdLabelValue, workerLabelMap)

			Expect(err).ToNot(HaveOccurred(), "An error occurred while attempting to verify the AMD label by NFD: %v ", err)
			Expect(amdLabelFound).NotTo(BeFalse(), "AMD label check failed to match "+
				"label %s and label value %s on all nodes", amdLabel, amdLabelValue)
		})

	})

})
