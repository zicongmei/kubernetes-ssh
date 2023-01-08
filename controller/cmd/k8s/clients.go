package k8s

import (
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type Clients struct {
	clientSet        *kubernetes.Clientset
	controllerClient ctrl.Client
}

func New(kubeconfigPath string) Clients {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		glog.Fatalf("fail to build k8s config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("fail to build k8s client set: %v", err)
	}
	controllerClient, err := ctrl.New(config, ctrl.Options{})
	if err != nil {
		glog.Fatalf("fail to build k8s controller client: %v", err)
	}
	return Clients{
		clientSet:        clientset,
		controllerClient: controllerClient,
	}
}

func (c *Clients) GetClientSet() *kubernetes.Clientset {
	return c.clientSet
}

func (c *Clients) GetControllerClient() ctrl.Client {
	return c.controllerClient
}
