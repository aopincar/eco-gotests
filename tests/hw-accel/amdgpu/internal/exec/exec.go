package exec

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

func boolPtr(b bool) *bool {
	return &b
}
func ptrInt64(i int64) *int64 {
	return &i
}

func GetPodLogs(clientset *clients.Settings, namespace, podName string) (string, error) {
	req := clientset.Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})

	logStream, err := req.Stream(context.TODO())
	if err != nil {
		return "", fmt.Errorf("error opening log stream: %v", err)
	}
	defer logStream.Close()

	var logs strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := logStream.Read(buf)
		if n > 0 {
			logs.WriteString(string(buf[:n]))
		}
		if err != nil {
			break
		}
	}

	return logs.String(), nil
}

// RunCommandsOnSpecificNode runs commands on a specific node by creating a pod on that node.
func RunCommandsOnSpecificNode(clientset *clients.Settings, podName, namespace, nodeName string, commands []string) (string, error) {
	// Validate input parameters
	if podName == "" || namespace == "" || nodeName == "" {
		return "", fmt.Errorf("podName, namespace, and nodeName cannot be empty")
	}
	// Check if pod already exists
	_, err := clientset.Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err == nil {
		return "", fmt.Errorf("pod %s already exists in namespace %s", podName, namespace)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			HostPID:       true,
			HostNetwork:   true,
			NodeName:      nodeName,
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{
				{
					Name: "host-root",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/",
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name: "debugger",

					Image:   "quay.io/aopincar/ecosys-amd/ubi9-tools:0.0.1",
					Command: commands,
					SecurityContext: &corev1.SecurityContext{
						Privileged: boolPtr(true),
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host-root",
							MountPath: "/host",
						},
					},
				},
			},
		},
	}

	_, err = clientset.Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create pod: %v", err)
	}

	glog.Info("Waiting for pod to complete...")
	watch, err := clientset.Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + podName,
		TimeoutSeconds: ptrInt64(120),
	})
	if err != nil {
		return "", fmt.Errorf("failed to watch pod: %v", err)
	}
	defer watch.Stop()

	completed := false
	var phase corev1.PodPhase

	for event := range watch.ResultChan() {
		p, ok := event.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		phase = p.Status.Phase
		if phase == corev1.PodSucceeded || phase == corev1.PodFailed {
			completed = true
			break
		}
	}

	if !completed {
		return "", fmt.Errorf("timed out waiting for pod to complete")
	}

	// Brief delay to ensure logs are fully available
	time.Sleep(5 * time.Second)

	logs, err := GetPodLogs(clientset, namespace, podName)
	if err != nil {
		return "", fmt.Errorf("pod completed with phase %s but failed to get logs: %v", phase, err)
	}

	if phase == corev1.PodFailed {
		return "", fmt.Errorf("command failed. Logs:\n%s", logs)
	}

	// Clean up
	err = clientset.Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		return "", fmt.Errorf("failed deleting pod:%s,%v", podName, err)
	}
	return logs, nil
}

//Image: "quay.io/wabouham/ecosys-nvidia/ubi9-tools:0.0.1",
