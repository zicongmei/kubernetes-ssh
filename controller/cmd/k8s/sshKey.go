package k8s

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
)

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
