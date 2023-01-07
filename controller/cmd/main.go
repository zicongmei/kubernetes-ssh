package main

import (
	"flag"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/zicongmei/kubernetes-ssh/controller/cmd/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespaceFlag  string
	podNumFlag     int
	kubeconfigFlag string
	namePrefixFlag string
)

func init() {
	defaultKubeconfigPath := path.Join(os.Getenv("HOME"), ".kube/config")
	flag.StringVar(&namespaceFlag, "namespace", "default", "Namespace.")
	flag.StringVar(&namePrefixFlag, "name_prefix", "sample", "Prefix of names.")
	flag.IntVar(&podNumFlag, "pod_num", 2, "Number of pods.")
	flag.StringVar(&kubeconfigFlag, "kubeconfig", defaultKubeconfigPath, "Path to the kubeconfig.")
	flag.Parse()
}

func main() {
	flag.Set("logtostderr", "true")
	cs := buildK8sClient(kubeconfigFlag)
	k8s.DeployK8sObjects(cs, namespaceFlag, namePrefixFlag, podNumFlag)
}

func buildK8sClient(kubeconfigPath string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		glog.Fatalf("fail to build k8s config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("fail to build k8s client set: %v", err)
	}
	return clientset
}
