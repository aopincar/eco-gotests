package assisted

import (
	"context"
	"fmt"
	"time"

	"math/rand"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-gotests/pkg/clients"
	"github.com/openshift-kni/eco-gotests/pkg/msg"
	hiveextV1Beta1 "github.com/openshift/assisted-service/api/hiveextension/v1beta1"
	agentInstallV1Beta1 "github.com/openshift/assisted-service/api/v1beta1"
	hiveV1 "github.com/openshift/hive/apis/hive/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	agentInfraEnvLabel = "infraenvs.agent-install.openshift.io"
	agentBMHLabel      = "agent-install.openshift.io/bmh"
)

// InfraEnvBuilder provides struct for the infraenv object containing connection to
// the cluster and the infraenv definitions.
type InfraEnvBuilder struct {
	Definition *agentInstallV1Beta1.InfraEnv
	Object     *agentInstallV1Beta1.InfraEnv
	errorMsg   string
	apiClient  *clients.Settings
}

// NewInfraEnvBuilder creates a new instance of InfraEnvBuilder.
func NewInfraEnvBuilder(apiClient *clients.Settings, name, nsname, psName string) *InfraEnvBuilder {
	glog.V(100).Infof(
		"Initializing new infraenv structure with the following params: "+
			"name: %s, namespace: %s, pull-secret: %s",
		name, nsname, psName)

	builder := InfraEnvBuilder{
		apiClient: apiClient,
		Definition: &agentInstallV1Beta1.InfraEnv{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
			Spec: agentInstallV1Beta1.InfraEnvSpec{
				PullSecretRef: &corev1.LocalObjectReference{
					Name: psName,
				},
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the infraenv is empty")

		builder.errorMsg = "infraenv 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the infraenv is empty")

		builder.errorMsg = "infraenv 'namespace' cannot be empty"
	}

	if psName == "" {
		glog.V(100).Infof("The pull-secret ref of the infraenv is empty")

		builder.errorMsg = "infraenv 'pull-secret' cannot be empty"
	}

	return &builder
}

// WithClusterRef sets the cluster reference to be used by the infraenv.
func (builder *InfraEnvBuilder) WithClusterRef(name, nsname string) *InfraEnvBuilder {
	glog.V(100).Infof("Adding clusterRef %s in namespace %s to InfraEnv %s", name, nsname, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if name == "" {
		glog.V(100).Infof("The name of the infraenv clusterRef is empty")

		builder.errorMsg = "infraenv clusterRef 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the infraenv clusterRef is empty")

		builder.errorMsg = "infraenv clusterRef 'namespace' cannot be empty"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.ClusterRef = &agentInstallV1Beta1.ClusterReference{
		Name:      name,
		Namespace: nsname,
	}

	return builder
}

// WithAdditionalNTPSource adds additional servers as NTP sources for the spoke cluster.
func (builder *InfraEnvBuilder) WithAdditionalNTPSource(ntpSource string) *InfraEnvBuilder {
	glog.V(100).Infof("Adding ntpSource %s to InfraEnv %s", ntpSource, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.AdditionalNTPSources = append(builder.Definition.Spec.AdditionalNTPSources, ntpSource)

	return builder
}

// WithSSHAuthorizedKey sets the authorized ssh key for accessing the nodes during discovery.
func (builder *InfraEnvBuilder) WithSSHAuthorizedKey(sshAuthKey string) *InfraEnvBuilder {
	glog.V(100).Infof("Adding sshAuthorizedKey %s to InfraEnv %s", sshAuthKey, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.SSHAuthorizedKey = sshAuthKey

	return builder
}

// WithAgentLabel adds labels to be applied to agents that boot from the infraenv.
func (builder *InfraEnvBuilder) WithAgentLabel(key, value string) *InfraEnvBuilder {
	glog.V(100).Infof("Adding agentLabel %s:%s to InfraEnv %s", key, value, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	if builder.Definition.Spec.AgentLabels == nil {
		builder.Definition.Spec.AgentLabels = make(map[string]string)
	}

	builder.Definition.Spec.AgentLabels[key] = value

	return builder
}

// WithProxy includes a proxy configuration to be used by the infraenv.
func (builder *InfraEnvBuilder) WithProxy(proxy agentInstallV1Beta1.Proxy) *InfraEnvBuilder {
	glog.V(100).Infof("Adding proxy %s to InfraEnv %s", proxy, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.Proxy = &proxy

	return builder
}

// WithNmstateConfigLabelSelector adds a selector for identifying
// nmstateconfigs that should be applied to this infraenv.
func (builder *InfraEnvBuilder) WithNmstateConfigLabelSelector(selector metaV1.LabelSelector) *InfraEnvBuilder {
	glog.V(100).Infof("Adding nmstateconfig selector %s to InfraEnv %s", &selector, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.NMStateConfigLabelSelector = selector

	return builder
}

// WithCPUType sets the cpu architecture for the discovery ISO.
func (builder *InfraEnvBuilder) WithCPUType(arch string) *InfraEnvBuilder {
	glog.V(100).Infof("Adding cpuArchitecture %s to InfraEnv %s", arch, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.CpuArchitecture = arch

	return builder
}

// WithIgnitionConfigOverride includes the specified ignitionconfigoverride for discovery.
func (builder *InfraEnvBuilder) WithIgnitionConfigOverride(override string) *InfraEnvBuilder {
	glog.V(100).Infof("Adding ignitionConfigOverride %s to InfraEnv %s", override, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IgnitionConfigOverride = override

	return builder
}

// WithIPXEScriptType modifies the IPXE script type generated by the infraenv.
func (builder *InfraEnvBuilder) WithIPXEScriptType(scriptType agentInstallV1Beta1.IPXEScriptType) *InfraEnvBuilder {
	glog.V(100).Infof("Adding ipxeScriptType %s to InfraEnv %s", scriptType, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IPXEScriptType = scriptType

	return builder
}

// WithKernelArgument appends kernel configurations to be configured by the infraenv.
func (builder *InfraEnvBuilder) WithKernelArgument(kernelArg agentInstallV1Beta1.KernelArgument) *InfraEnvBuilder {
	glog.V(100).Infof("Adding kernelArgument %s to InfraEnv %s", kernelArg, builder.Definition.Name)

	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.KernelArguments = append(builder.Definition.Spec.KernelArguments, kernelArg)

	return builder
}

// WaitForDiscoveryISOCreation waits the defined timeout for the discovery ISO to be generated.
func (builder *InfraEnvBuilder) WaitForDiscoveryISOCreation(timeout time.Duration) (*InfraEnvBuilder, error) {
	if builder.Definition == nil {
		glog.V(100).Infof("The infraenv is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("InfraEnv")
	}

	if builder.errorMsg != "" {
		return builder, nil
	}

	// Polls every retryInterval to determine if infraenv in desired state.
	var err error
	err = wait.PollImmediate(retryInterval, timeout, func() (bool, error) {
		builder.Object, err = builder.Get()

		if err != nil {
			return false, nil
		}

		return builder.Object.Status.CreatedTime != nil, nil

	})

	if err == nil {
		return builder, nil
	}

	return nil, err
}

// GetAllAgents returns a slice of agentBuilders of all agents belonging to the infraenv.
func (builder *InfraEnvBuilder) GetAllAgents() ([]*agentBuilder, error) {
	glog.V(100).Infof("Getting all agents from infraenv %s",
		builder.Definition.Name)

	if !builder.Exists() {
		return nil, fmt.Errorf("cannot get agents from non-existent infraenv")
	}

	agents, err := builder.GetAgentsByLabel(agentInfraEnvLabel, builder.Definition.Name)
	if err != nil {
		return nil, err
	}

	return agents, nil
}

// GetAgentsByRole returns a slice of agentBuilders of agents matching specified role belonging to the infraenv.
func (builder *InfraEnvBuilder) GetAgentsByRole(role string) ([]*agentBuilder, error) {
	glog.V(100).Infof("Getting agents from infraenv %s matching role %s",
		builder.Definition.Name, role)

	if !builder.Exists() {
		glog.V(100).Infof("Cannot get agents from non-existent infraenv: %s",
			role)

		return nil, fmt.Errorf("cannot get agents from non-existent infraenv")
	}

	var agents, agentsByRole []*agentBuilder

	agents, err := builder.GetAllAgents()
	if err != nil {
		return nil, err
	}

	for _, agent := range agents {
		if string(agent.Object.Status.Role) == role {
			agentsByRole = append(agentsByRole, agent)
		}
	}

	return agentsByRole, nil
}

// GetAgentByBMH returns an agentBuilder for the agent matching specified BMH belonging to the infraenv.
func (builder *InfraEnvBuilder) GetAgentByBMH(bmhName string) (*agentBuilder, error) {
	glog.V(100).Infof("Getting agent from infraenv %s matching bmh %s",
		builder.Definition.Name, bmhName)

	if !builder.Exists() {
		return nil, fmt.Errorf("cannot get agents from non-existent infraenv")
	}

	agents, err := builder.GetAgentsByLabel(agentBMHLabel, bmhName)
	if err != nil {
		return nil, err
	}

	switch len(agents) {
	case 1:
		return agents[0], nil
	case 0:
		glog.V(100).Infof("Found no agents referencing bmh %s", bmhName)

		return nil, fmt.Errorf("found no agents referencing bmh %s", bmhName)
	default:
		glog.V(100).Infof("Found multiple agent referencing bmh %s", bmhName)

		return nil, fmt.Errorf("found multiple agents referencing bmh %s", bmhName)
	}
}

// GetAgentByName returns an agentBuilder for the agent matching specified name belonging to the infraenv.
func (builder *InfraEnvBuilder) GetAgentByName(name string) (*agentBuilder, error) {
	glog.V(100).Infof("Getting agent from infraenv %s with name %s",
		builder.Definition.Name, name)

	if !builder.Exists() {
		return nil, fmt.Errorf("cannot get agents from non-existent infraenv")
	}

	agent, err := PullAgent(builder.apiClient, name, builder.Definition.Namespace)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

// GetAgentsByLabel returns a slice of agentBuilders for agents matching specified label.
func (builder *InfraEnvBuilder) GetAgentsByLabel(key, value string) ([]*agentBuilder, error) {
	glog.V(100).Infof("Getting agent matching label %s:%s",
		key, value)

	if !builder.Exists() {
		return nil, fmt.Errorf("cannot get agents from non-existent infraenv")
	}

	matchLabel := map[string]string{key: value}

	var agents agentInstallV1Beta1.AgentList

	err := builder.apiClient.List(context.TODO(), &agents, goclient.MatchingLabels(matchLabel))
	if err != nil {
		return nil, err
	}

	return builder.createBuilderListFromAgentList(agents.Items), nil
}

// WaitForAgentsToRegister waits the specified time for agents to register
// matching the provisioninRequirements of the related AgentClusterInstall.
func (builder *InfraEnvBuilder) WaitForAgentsToRegister(timeout time.Duration) ([]*agentBuilder, error) {
	if !builder.Exists() {
		return nil, fmt.Errorf("cannot get agents from non-existent infraenv")
	}

	agentclusterinstall, err := builder.GetAgentClusterInstallFromInfraEnv()

	if err != nil {
		return nil, err
	}

	var agentList []*agentBuilder

	agentCount := agentclusterinstall.Spec.ProvisionRequirements.ControlPlaneAgents +
		agentclusterinstall.Spec.ProvisionRequirements.WorkerAgents

	// Polls every retryInterval to determine if agent has registered.
	err = wait.PollImmediate(retryInterval, timeout, func() (bool, error) {

		agentList, err = builder.GetAllAgents()

		if err != nil {
			return false, err
		}

		return len(agentList) == agentCount, nil
	})

	return agentList, err
}

// WaitForMasterAgents waits the specified time for agents with the role master
// to register and match the ControlPlaneAgents count in ProvisionRequirements of the related AgentClusterInstall.
func (builder *InfraEnvBuilder) WaitForMasterAgents(timeout time.Duration) ([]*agentBuilder, error) {
	agentclusterinstall, err := builder.GetAgentClusterInstallFromInfraEnv()

	if err != nil {
		return nil, err
	}

	var agentList []*agentBuilder

	agentCount := agentclusterinstall.Spec.ProvisionRequirements.ControlPlaneAgents

	// Polls every retryInterval to determine if agent has registered.
	err = wait.PollImmediate(retryInterval, timeout, func() (bool, error) {
		agentList, err = builder.GetAgentsByRole("master")
		if err != nil {

			return false, err
		}

		return len(agentList) == agentCount, nil
	})

	return agentList, err
}

// WaitForMasterAgentCount waits the specified time for agents
// with the role master to register and match the specified count.
func (builder *InfraEnvBuilder) WaitForMasterAgentCount(count int, timeout time.Duration) ([]*agentBuilder, error) {
	var agentList []*agentBuilder

	// Polls every retryInterval to determine if agent has registered.
	err := wait.PollImmediate(retryInterval, timeout, func() (bool, error) {

		agentList, err := builder.GetAgentsByRole("master")
		if err != nil {

			return false, err
		}

		return len(agentList) == count, nil
	})

	return agentList, err
}

// GetRandomMasterAgent returns an agentBuilder of a random agent that has it's role set to master.
func (builder *InfraEnvBuilder) GetRandomMasterAgent() (*agentBuilder, error) {
	rand.Seed(time.Now().UnixNano())

	agentList, err := builder.GetAgentsByRole("master")
	if err != nil {
		return nil, err
	}

	if len(agentList) == 0 {
		return nil, fmt.Errorf("could not find any master agents")
	}

	randInt := rand.Intn(len(agentList))

	return agentList[randInt], nil
}

// WaitForWorkerAgents waits the specified time for agents with the role worker to register and match the WorkerAgents
// count in ProvisionRequirements of the related AgentClusterInstall.
func (builder *InfraEnvBuilder) WaitForWorkerAgents(timeout time.Duration) ([]*agentBuilder, error) {
	agentclusterinstall, err := builder.GetAgentClusterInstallFromInfraEnv()

	if err != nil {
		return nil, err
	}

	var agentList []*agentBuilder

	agentCount := agentclusterinstall.Spec.ProvisionRequirements.WorkerAgents

	// Polls every retryInterval to determine if agent has registered.
	err = wait.PollImmediate(retryInterval, timeout, func() (bool, error) {

		agentList, err = builder.GetAgentsByRole("worker")
		if err != nil {

			return false, err
		}

		return len(agentList) == agentCount, nil
	})

	return agentList, err
}

// WaitForWorkerAgentCount waits the specified time
// for agents with the role worker to register and match the specified count.
func (builder *InfraEnvBuilder) WaitForWorkerAgentCount(count int, timeout time.Duration) ([]*agentBuilder, error) {
	var agentList []*agentBuilder

	// Polls every retryInterval to determine if agent has registered.
	err := wait.PollImmediate(retryInterval, timeout, func() (bool, error) {

		agentList, err := builder.GetAgentsByRole("worker")
		if err != nil {

			return false, err
		}

		return len(agentList) == count, nil
	})

	return agentList, err
}

// GetRandomWorkerAgent returns an agentBuilder of a random agent that has it's role set to worker.
func (builder *InfraEnvBuilder) GetRandomWorkerAgent() (*agentBuilder, error) {
	rand.Seed(time.Now().UnixNano())

	agentList, err := builder.GetAgentsByRole("worker")
	if err != nil {
		return nil, err
	}

	if len(agentList) == 0 {
		return nil, fmt.Errorf("could not find any worker agents")
	}

	randInt := rand.Intn(len(agentList))

	return agentList[randInt], nil
}

// createBuilderListFromAgentList takes an Agent slice and transforms it into an *agentBuilder slice.
func (builder *InfraEnvBuilder) createBuilderListFromAgentList(agents []agentInstallV1Beta1.Agent) []*agentBuilder {
	var buliderList []*agentBuilder

	for _, agent := range agents {
		copiedAgent := agent
		buliderList = append(buliderList, newAgentBuilder(builder.apiClient, &copiedAgent))
	}

	return buliderList
}

// GetAgentClusterInstallFromInfraEnv returns the AgentClusterInstall that is referenced by this InfraEnv.
func (builder *InfraEnvBuilder) GetAgentClusterInstallFromInfraEnv() (*hiveextV1Beta1.AgentClusterInstall, error) {
	if !builder.Exists() {
		glog.V(100).Infof("Getting infraenv %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

		return nil, fmt.Errorf("cannot wait from agents to register with non-existent infraenv")
	}

	var clusterdeployment hiveV1.ClusterDeployment

	glog.V(100).Infof("Getting clusterdeployment %s in namespace %s",
		builder.Object.Spec.ClusterRef.Name, builder.Object.Spec.ClusterRef.Namespace)

	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Object.Spec.ClusterRef.Name,
		Namespace: builder.Object.Spec.ClusterRef.Namespace,
	}, &clusterdeployment)

	if err != nil {
		glog.V(100).Infof("Unable to get clusterdeployment %s referenced by infraenv %s",
			builder.Object.Spec.ClusterRef.Name, builder.Definition.Name)

		return nil, err
	}

	glog.V(100).Infof("Getting agentclusterinstall %s",
		clusterdeployment.Spec.ClusterInstallRef.Name)

	var agentclusterinstall hiveextV1Beta1.AgentClusterInstall

	err = builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      clusterdeployment.Spec.ClusterInstallRef.Name,
		Namespace: clusterdeployment.Namespace,
	}, &agentclusterinstall)

	if err != nil {
		glog.V(100).Infof("Unable to get agentclusterinstall %s referenced by clusterdeployment %s",
			clusterdeployment.Spec.ClusterInstallRef.Name, clusterdeployment.Name)

		return nil, err
	}

	return &agentclusterinstall, nil
}

// Get fetches the defined infraenv from the cluster.
func (builder *InfraEnvBuilder) Get() (*agentInstallV1Beta1.InfraEnv, error) {
	glog.V(100).Infof("Getting infraenv %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	infraEnv := &agentInstallV1Beta1.InfraEnv{}

	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, infraEnv)

	if err != nil {
		return nil, err
	}

	return infraEnv, err
}

// Create generates a infraenv on the cluster.
func (builder *InfraEnvBuilder) Create() (*InfraEnvBuilder, error) {
	glog.V(100).Infof("Creating the infraenv %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if builder.errorMsg != "" {
		return nil, fmt.Errorf(builder.errorMsg)
	}

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Update modifies an existing infraenv on the cluster.
func (builder *InfraEnvBuilder) Update(force bool) (*InfraEnvBuilder, error) {
	glog.V(100).Infof("Updating infraenv %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof("infraenv %s in namespace %s does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.errorMsg = "Cannot update non-existent infraenv"
	}

	if builder.errorMsg != "" {
		return nil, fmt.Errorf(builder.errorMsg)
	}

	err := builder.apiClient.Update(context.TODO(), builder.Definition)

	if err != nil {
		if force {
			glog.V(100).Infof(
				"Failed to update the infraenv object %s in namespace %s. "+
					"Note: Force flag set, executed delete/create methods instead",
				builder.Definition.Name, builder.Definition.Namespace,
			)

			_, err = builder.Delete()
			builder.Definition.ResourceVersion = ""

			if err != nil {
				glog.V(100).Infof(
					"Failed to update the infraenv object %s in namespace %s, "+
						"due to error in delete function",
					builder.Definition.Name, builder.Definition.Namespace,
				)

				return nil, err
			}

			return builder.Create()
		}
	}

	return builder, err
}

// Delete removes an infraenv from the cluster.
func (builder *InfraEnvBuilder) Delete() (*InfraEnvBuilder, error) {
	glog.V(100).Infof("Deleting the infraenv %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return builder, fmt.Errorf("infraenv cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {
		return builder, fmt.Errorf("cannot delete infraenv: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Exists checks if the defined infraenv has already been created.
func (builder *InfraEnvBuilder) Exists() bool {
	glog.V(100).Infof("Checking if infraenv %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}