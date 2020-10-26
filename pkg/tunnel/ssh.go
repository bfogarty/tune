package tunnel

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
)


func generateKeyPair() ([]byte, []byte, error) {
	// generate a new 2048-bit RSA key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

        // encode the private key in PEM format
	privateKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	})

        // encode the public key for inclusion in authorized_keys
	publicKey, err := ssh.NewPublicKey(&rsaKey.PublicKey)
	if err != nil {
                return nil, nil, err
	}
        publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

        return privateKeyBytes, publicKeyBytes, nil
}

