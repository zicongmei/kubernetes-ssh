package yamlDecoder_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/zicongmei/kubernetes-ssh/controller/cmd/k8s/yamlDecoder"
)

func TestDecoder(t *testing.T) {
	yaml := `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-demo
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 30Gi
  storageClassName: standard-rwo
---
kind: Pod
apiVersion: v1
metadata:
  name: pod-demo
spec:
  volumes:
    - name: pvc-demo-vol
      persistentVolumeClaim:
       claimName: pvc-demo
  containers:
    - name: pod-demo
      image: nginx
      ports:
        - containerPort: 80
          name: "http-server"
      volumeMounts:
        - mountPath: "/usr/share/nginx/html"
          name: pvc-demo-vol
`

	g := gomega.NewGomegaWithT(t)
	objs, err := yamlDecoder.Decode(yaml)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(len(objs)).To(gomega.Equal(2))

	g.Expect(objs[0].GetKind()).To(gomega.Equal("PersistentVolumeClaim"))
	g.Expect(objs[1].GetKind()).To(gomega.Equal("Pod"))
}
