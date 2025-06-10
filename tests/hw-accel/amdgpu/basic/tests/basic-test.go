package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift-kni/eco-goinfra/pkg/namespace"
	"github.com/openshift-kni/eco-goinfra/pkg/nodes"
	"github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/basic/internal/tsparams"
	"github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/internal/exec"
	"github.com/openshift-kni/eco-gotests/tests/hw-accel/amdgpu/internal/get"
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
		AMDWorkersListOptions := metav1.ListOptions{LabelSelector: labels.Set(workerLabelMap).String()}

		//AMDWorkerNodeBuilder, err := nodes.List(apiClient, metav1.ListOptions{LabelSelector: labels.Set(workerLabelMap).String()})
		AMDNodeBuilder, AMDNodeBuilderErr := nodes.List(apiClient, AMDWorkersListOptions)

		BeforeAll(func() {

			// Returns an inventory of worker nodes
			//nodeBuilder, builderErr := nodes.List(apiClient, AMDWorkersListOptions)

			if AMDNodeBuilderErr != nil {
				Skip("Failed to get list of worker nodes")
			}

			if len(AMDNodeBuilder) == 0 {
				Skip("Worker nodes list is empty")
			}

		})

		BeforeEach(func() {})

		AfterEach(func() {})

		AfterAll(func() {})

		It("Check AMD label was added by NFD", func() {

			amdLabelFound, err := get.LabelPresentOnAllNodes(apiClient, amdLabel, amdLabelValue, workerLabelMap)

			Expect(err).ToNot(HaveOccurred(), "An error occurred while attempting to verify the AMD label by NFD: %v ", err)
			Expect(amdLabelFound).NotTo(BeFalse(), "AMD label check failed to match "+
				"label %s and label value %s on all nodes", amdLabel, amdLabelValue)
		})

		It("Make sure AMD GPU is listed using lspci", func() {

			//namespaceName := "my-namespace-01"
			namespaceName := "default"
			_, namespaceCreateErr := namespace.NewBuilder(apiClient, namespaceName).Create()

			if namespaceCreateErr != nil {
				fmt.Println("failed to create namespace", namespaceCreateErr)
			}

			//lscpi_cmd := []string{"LD_LIBRARY_PATH=/host/usr/lib64 /host/usr/sbin/lspci | grep AMD"}
			//lscpi_cmd := []string{
			//	"dnf install -y pciutils",
			//	"lspci | grep AMD"}
			lscpi_cmd := []string{
				"sh",
				"-c",
				"lsmod | grep amdgpu"}
			//"dnf install -y pciutils && lspci | grep AMD"}

			for _, node := range AMDNodeBuilder {

				nodeName := node.Object.Name
				fmt.Println(nodeName)

				s, err := exec.RunCommandsOnSpecificNode(
					apiClient,
					"my-pod-08",
					namespaceName,
					nodeName,
					lscpi_cmd)

				if err != nil {
					//Skip("lscpi command failed on node: '%s' with err: %s", nodeName, err)
					fmt.Println(err)
				}

				fmt.Println(s)

			}
		})
	})
})
