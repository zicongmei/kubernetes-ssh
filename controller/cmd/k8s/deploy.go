package k8s

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
	"path"

	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	//image   = "linuxserver/openssh-server"
	image   = "sheixinsheisb/ssh-server"
	appPort = 22

	sshSecretName    = "ssh"
	sshVolumeName    = "ssh-volume"
	bootStraptVolume = "bootstrapt"

	bootstraptDir      = "/etc/kssh"
	bootstraptFileName = "bootstrapt.sh"

	sshSize = 4096
)

func DeployK8sObjects(
	cs *kubernetes.Clientset,
	namespace string,
	namePrefix string,
	podNum int) {
	objs := buildK8sObjects(namespace, namePrefix, podNum)
	ctx := context.Background()
	_, err := cs.CoreV1().Namespaces().Get(ctx, namespace, metaV1.GetOptions{})
	if errors.IsNotFound(err) {
		v1Namespace := &coreV1.Namespace{
			ObjectMeta: metaV1.ObjectMeta{
				Name: namespace,
			},
		}
		if _, err := cs.CoreV1().Namespaces().
			Create(ctx, v1Namespace, metaV1.CreateOptions{}); err != nil {
			glog.Fatalf("failed to create namespace: %v", err)
		}
	} else if err != nil {
		glog.Fatalf("failed to create namespace: %v", err)
	}
	for _, deploy := range objs.deployList {
		_, err := cs.AppsV1().Deployments(namespace).Get(ctx, deploy.Name, metaV1.GetOptions{})
		if errors.IsNotFound(err) {
			if _, err := cs.AppsV1().Deployments(namespace).
				Create(ctx, deploy, metaV1.CreateOptions{}); err != nil {
				glog.Fatalf("failed to create deployment %q: %v", deploy.Name, err)
			}
			glog.Infof("Deployment %q deployed.", deploy.Name)
		} else if err != nil {
			glog.Fatalf("failed to get deployment %q: %v", deploy.Name, err)
		}
	}

	for _, svc := range objs.svcList {
		_, err := cs.CoreV1().Services(namespace).Get(ctx, svc.Name, metaV1.GetOptions{})
		if errors.IsNotFound(err) {
			if _, err := cs.CoreV1().Services(namespace).
				Create(ctx, svc, metaV1.CreateOptions{}); err != nil {
				glog.Fatalf("failed to create service %q: %v", svc.Name, err)
			}
			glog.Infof("Service %q deployed.", svc.Name)
		} else if err != nil {
			glog.Fatalf("failed to get service %q: %v", svc.Name, err)
		}
	}

	for _, secret := range objs.secretList {
		_, err := cs.CoreV1().Secrets(namespace).Get(ctx, secret.Name, metaV1.GetOptions{})
		if errors.IsNotFound(err) {
			if _, err := cs.CoreV1().Secrets(namespace).
				Create(ctx, secret, metaV1.CreateOptions{}); err != nil {
				glog.Fatalf("failed to create secret %q: %v", secret.Name, err)
			}
			glog.Infof("Secret %q deployed.", secret.Name)
		} else if err != nil {
			glog.Fatalf("failed to get secret %q: %v", secret.Name, err)
		}
	}

	for _, configMap := range objs.configMapList {
		_, err := cs.CoreV1().ConfigMaps(namespace).Get(ctx, configMap.Name, metaV1.GetOptions{})
		if errors.IsNotFound(err) {
			if _, err := cs.CoreV1().ConfigMaps(namespace).
				Create(ctx, configMap, metaV1.CreateOptions{}); err != nil {
				glog.Fatalf("failed to create configMap %q: %v", configMap.Name, err)
			}
			glog.Infof("ConfigMap %q deployed.", configMap.Name)
		} else if err != nil {
			glog.Fatalf("failed to get configMap %q: %v", configMap.Name, err)
		}
	}
}

type buildK8sObjectsResponse struct {
	deployList    []*appsV1.Deployment
	svcList       []*coreV1.Service
	secretList    []*coreV1.Secret
	configMapList []*coreV1.ConfigMap
}

func buildK8sObjects(
	namespace string,
	namePrefix string,
	podNum int) buildK8sObjectsResponse {
	response := buildK8sObjectsResponse{}

	authorizedHosts := []byte{}
	for i := 0; i < podNum; i++ {
		name := fmt.Sprintf("%s-%d", namePrefix, i)
		response.deployList = append(response.deployList, createDeployment(namespace, name))
		response.svcList = append(response.svcList, createService(namespace, name))

		privateKey, publicKey := generateSSHKey()
		authorizedHosts = append(authorizedHosts, publicKey...)
		response.secretList = append(response.secretList, createSecret(namespace, name, privateKey, publicKey))
	}
	for _, secret := range response.secretList {
		secret.Data["authorized_keys"] = authorizedHosts
	}
	response.configMapList = createConfigMaps(namespace)
	return response
}

func createDeployment(
	namespace string,
	name string) *appsV1.Deployment {
	replica := int32(1)
	return &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: appsV1.DeploymentSpec{
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					"run": name,
				},
			},
			Replicas: &replica,
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						"run": name,
					},
				},
				Spec: coreV1.PodSpec{
					InitContainers: []coreV1.Container{
						{
							Name:    fmt.Sprintf("%s-init", name),
							Image:   image,
							Command: []string{"bash"},
							Args: []string{
								path.Join(bootstraptDir, bootstraptFileName),
							},
							VolumeMounts: []coreV1.VolumeMount{
								{
									Name:      sshVolumeName,
									MountPath: "/root/.ssh",
								},
								{
									Name:      sshSecretName,
									MountPath: "/tmp/ssh",
									ReadOnly:  true,
								},
								{
									Name:      bootStraptVolume,
									MountPath: bootstraptDir,
								},
							},
						},
					},
					Containers: []coreV1.Container{
						{
							Name:  name,
							Image: image,
							Ports: []coreV1.ContainerPort{
								{
									Name:          name,
									ContainerPort: appPort,
								},
							},
							VolumeMounts: []coreV1.VolumeMount{
								{
									Name:      sshVolumeName,
									MountPath: "/root/.ssh",
									// ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []coreV1.Volume{
						{
							Name: sshSecretName,
							VolumeSource: coreV1.VolumeSource{
								Secret: &coreV1.SecretVolumeSource{
									SecretName: name,
								},
							},
						},
						{
							Name: bootStraptVolume,
							VolumeSource: coreV1.VolumeSource{
								ConfigMap: &coreV1.ConfigMapVolumeSource{
									LocalObjectReference: coreV1.LocalObjectReference{
										Name: bootstraptFileName,
									},
								},
							},
						},
						{
							Name: sshVolumeName,
							VolumeSource: coreV1.VolumeSource{
								EmptyDir: &coreV1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

func createService(
	namespace string,
	name string) *coreV1.Service {
	return &coreV1.Service{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				"run": name,
			},
		},
		Spec: coreV1.ServiceSpec{
			Ports: []coreV1.ServicePort{
				{
					Port:     appPort,
					Protocol: coreV1.ProtocolTCP,
				},
			},
			ClusterIP: "None",
			Selector: map[string]string{
				"run": name,
			},
		},
	}
}

func createSecret(
	namespace string,
	name string,
	privateKey, publicKey []byte) *coreV1.Secret {
	return &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Type: coreV1.SecretTypeOpaque,
		Data: map[string][]byte{
			"id_rsa":     privateKey,
			"id_rsa.pub": publicKey,
		},
	}
}

//go:embed pod-bootstrapt.sh
var podBootstrapt string

func createConfigMaps(namespace string) []*coreV1.ConfigMap {
	return []*coreV1.ConfigMap{
		{
			ObjectMeta: metaV1.ObjectMeta{
				Namespace: namespace,
				Name:      bootstraptFileName,
			},
			Data: map[string]string{
				bootstraptFileName: podBootstrapt,
			},
		},
	}
}

func generateSSHKey() ([]byte, []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, sshSize)
	if err != nil {
		glog.Fatalf("failed to generate private key: %v", err)
	}

	privateKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		glog.Fatalf("failed to generate public key: %v", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	return privateKeyBytes, publicKeyBytes
}
