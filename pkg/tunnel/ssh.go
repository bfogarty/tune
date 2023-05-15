package tunnel

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
)

func generateKeyPair() ([]byte, []byte, error) {
	// generate a new ed25519 keypair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// encode the private key in PEM format
	privKeyPKCS8, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	privateKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyPKCS8,
	})

	// encode the public key for inclusion in authorized_keys
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)

	return privateKeyBytes, publicKeyBytes, nil
}
