package k8s

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"github.com/golang/glog"
	"github.com/zicongmei/kubernetes-ssh/controller/cmd/k8s/yamlDecoder"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//go:embed templates/pod-bootstrapt.sh
var bootstraptContent string

//go:embed templates/systemObjs.yaml
var systemTmpl string

//go:embed templates/podObjs.yaml
var perPodTaml string

const (
	bootstraptKey = "bootstrapt.sh"
)

type TemplateData struct {
	Namespace               string
	Name                    string
	BootstraptConfigMapName string
	BootstraptContent       string
	Image                   string
	Port                    int
	AuthorizedKeys          string
	SSHPrivateKey           string
	SSHPublicKey            string
}

func DeployYaml(
	clients Clients,
	namespace string,
	namePrefix string,
	podNum int) {
	ctx := context.Background()
	client := clients.GetControllerClient()

	allObjs := generateObjs(namespace, namePrefix, podNum)
	for _, o := range allObjs {
		err := client.Create(ctx, o)
		if err != nil {
			glog.Fatalf("failed to create %q object %q: %v", o.GetKind(), o.GetName(), err)
		}
		glog.Infof("created %q object %q", o.GetKind(), o.GetName())
	}
}

func generateObjs(
	namespace string,
	namePrefix string,
	podNum int) []*unstructured.Unstructured {
	systemObjs := generateSystemObjs(namespace)
	podObjs := generateAllPods(namespace, namePrefix, podNum)
	return append(systemObjs, podObjs...)
}

func generateSystemObjs(namespace string) []*unstructured.Unstructured {
	data := TemplateData{
		Namespace:               namespace,
		BootstraptContent:       strings.Join(strings.Split(bootstraptContent, "\n"), `\n`),
		BootstraptConfigMapName: bootstraptKey,
	}
	tmpl, err := template.New("tmpl").Parse(systemTmpl)
	if err != nil {
		glog.Fatalf("failed to parse system template: %v", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		glog.Fatalf("failed to execute system template: %v", err)
	}
	fmt.Println(buf.String())
	objs, err := yamlDecoder.Decode(buf.String())
	if err != nil {
		glog.Fatalf("failed to decode the system yaml: %v", err)
	}
	return objs
}

type sshKeys struct {
	authorizedHosts []byte
	allPrivateKeys  [][]byte
	allPublicKeys   [][]byte
}

func generateAllPods(
	namespace string,
	namePrefix string,
	podNum int) []*unstructured.Unstructured {

	keys := &sshKeys{
		authorizedHosts: make([]byte, 0),
		allPrivateKeys:  make([][]byte, 0),
		allPublicKeys:   make([][]byte, 0),
	}
	for i := 0; i < podNum; i++ {
		privateKey, publicKey := generateSSHKey()
		keys.authorizedHosts = append(keys.authorizedHosts, publicKey...)
		keys.allPrivateKeys = append(keys.allPrivateKeys, privateKey)
		keys.allPublicKeys = append(keys.allPublicKeys, publicKey)
	}

	objs := []*unstructured.Unstructured{}
	for i := 0; i < podNum; i++ {
		name := fmt.Sprintf("%s-%d", namePrefix, i)
		o := generateOnePodObjs(namespace, name, keys, i)
		objs = append(objs, o...)
	}
	return objs
}

func generateOnePodObjs(
	namespace string,
	name string,
	keys *sshKeys,
	index int) []*unstructured.Unstructured {
	data := TemplateData{
		Namespace:               namespace,
		Name:                    name,
		BootstraptConfigMapName: bootstraptKey,
		Image:                   image,
		Port:                    appPort,
		AuthorizedKeys:          base64.StdEncoding.EncodeToString(keys.authorizedHosts),
		SSHPrivateKey:           base64.StdEncoding.EncodeToString(keys.allPrivateKeys[index]),
		SSHPublicKey:            base64.StdEncoding.EncodeToString(keys.allPublicKeys[index]),
	}
	tmpl, err := template.New("tmpl").Parse(perPodTaml)
	if err != nil {
		glog.Fatalf("failed to parse pod template: %v", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		glog.Fatalf("failed to execute pod template: %v", err)
	}
	objs, err := yamlDecoder.Decode(buf.String())
	if err != nil {
		glog.Fatalf("failed to decode the pod yaml: %v", err)
	}
	return objs
}
