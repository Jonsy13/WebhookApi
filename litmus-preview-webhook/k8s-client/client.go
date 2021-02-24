package client

import (
	"context"
	"os"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// GetKubeConfig is a function for creating kubeconfig
func GetKubeConfig() (*rest.Config, error) {
	KubeConfig := os.Getenv("KUBECONFIG")
	// Use in-cluster config if kubeconfig path is not specified
	if KubeConfig == "" {
		return rest.InClusterConfig()
	}

	return clientcmd.BuildConfigFromFlags("", KubeConfig)
}

// GetGenericK8sClient is a function for creatiing clientset
func GetGenericK8sClient() (*kubernetes.Clientset, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// genericPodDelete is used for deleting a pod with given label
func genericPodDelete(label, namespace string, clientset *kubernetes.Clientset) error {

	var name string
	//Getting all pods in litmus Namespace
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Errorf("failed to get pods in %v namespace, err: %v", err)
	}

	klog.Infof("[Info]: Selecting pod with label %v for reboot", label)
	for i := range pods.Items {
		if pods.Items[i].Labels["component"] == label {
			name = pods.Items[i].Name
		}
	}
	klog.Infof("[Info]: Deleting the Pod : %v", name)
	err = clientset.CoreV1().Pods("litmus").Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Errorf("failed to delete %v pod, err: %v", name, err)
	}
	klog.Infof("[Info]: %v pod deleted successfully \n", name)
	return nil
}

// DeletePod is function for deleting a pod whose image is updated based on label
func DeletePod(repoName, namespace string, clientset *kubernetes.Clientset) error {

	//Setting the label for deleting a Pod
	switch repoName {
	case "litmusportal-frontend":
		if err := genericPodDelete("litmusportal-frontend", namespace, clientset); err != nil {
			return errors.Errorf("Failed to delete pod for repo %v, err: %v", repoName, err)
		}
	case "litmusportal-server", "litmusportal-auth-server":
		if err := genericPodDelete("litmusportal-server", namespace, clientset); err != nil {
			return errors.Errorf("Failed to delete pod for repo %v, err: %v", repoName, err)
		}
	default:
		return errors.Errorf("No Appropriate Operation Found!")
	}
	return nil
}
