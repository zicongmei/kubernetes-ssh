package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	image = "linuxserver/openssh-server"
	port  = 2222
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
	cs := buildK8sClient(kubeconfigFlag)
	deployK8sObjects(cs, namespaceFlag, namePrefixFlag, podNumFlag)
}

func buildK8sClient(kubeconfigPath string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalf("fail to build k8s config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("fail to build k8s client set: %v", err)
	}
	return clientset
}

func deployK8sObjects(
	cs *kubernetes.Clientset,
	namespace string,
	namePrefix string,
	podNum int) {
	deployList, svcList := buildK8sObjects(namespace, namePrefix, podNum)
	ctx := context.Background()
	for _, deploy := range deployList {
		if _, err := cs.AppsV1().Deployments(namespace).
			Create(ctx, deploy, v1.CreateOptions{}); err != nil {
			log.Fatalf("failed to create deployment %q: %v", deploy.Name, err)
		}
	}
	for _, svc := range svcList {
		if _, err := cs.CoreV1().Services(namespace).
			Create(ctx, svc, v1.CreateOptions{}); err != nil {
			log.Fatalf("failed to create service %q: %v", svc.Name, err)
		}
	}

}

func buildK8sObjects(
	namespace string,
	namePrefix string,
	podNum int) ([]*appsV1.Deployment, []*coreV1.Service) {
	deployList := []*appsV1.Deployment{}
	svcList := []*coreV1.Service{}
	replica := int32(1)
	for i := 0; i < podNum; i++ {
		name := fmt.Sprintf("%s-%d", namePrefix, i)
		deployList = append(deployList, &appsV1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			},
			Spec: appsV1.DeploymentSpec{
				Selector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"run": name,
					},
				},
				Replicas: &replica,
				Template: coreV1.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{
							"run": name,
						},
					},
					Spec: coreV1.PodSpec{
						Containers: []coreV1.Container{
							{
								Name:  name,
								Image: image,
								Ports: []coreV1.ContainerPort{
									{
										Name:          name,
										ContainerPort: port,
									},
								},
							},
						},
					},
				},
			},
		})

		svcList = append(svcList, &coreV1.Service{
			ObjectMeta: v1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
				Labels: map[string]string{
					"run": name,
				},
			},
			Spec: coreV1.ServiceSpec{
				Ports: []coreV1.ServicePort{
					{
						Port:     port,
						Protocol: coreV1.ProtocolTCP,
					},
				},
				ClusterIP: "None",
				Selector: map[string]string{
					"run": name,
				},
			},
		})
	}
	return deployList, svcList
}
