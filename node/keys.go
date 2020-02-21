package node

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

const (
	privKeyFile = "priv.key"
	pubKeyFile  = "pub.key"
)

// WriteKeyPairIfNotExists writes new ecdsa keypair for node if not exists
func WriteKeyPairIfNotExists(c *Config) {
	if _, err := os.Stat(c.Node.Keyspath); os.IsNotExist(err) {
		if err := os.MkdirAll(c.Node.Keyspath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
		priv, pub := generateNewKeyPair()
		privPem, pubPem := EncodeKeyPair(priv, pub)
		log.Printf("path: %s", path.Join(c.Node.Keyspath, privKeyFile))
		if err := ioutil.WriteFile(path.Join(c.Node.Keyspath, privKeyFile), []byte(privPem), os.ModePerm); err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(path.Join(c.Node.Keyspath, pubKeyFile), []byte(pubPem), os.ModePerm); err != nil {
			log.Fatal(err)
		}
		log.Printf("waiting for all nodes to generate keypairs")
		time.Sleep(5 * time.Second)
	}
}

//LoadKeyPair loads private and public keys, with public PEM for messages
func LoadKeyPair(c *Config) (*ecdsa.PrivateKey, *ecdsa.PublicKey, string) {
	privPath := path.Join(c.Node.Keyspath, privKeyFile)
	pubPath := path.Join(c.Node.Keyspath, pubKeyFile)
	f, err := os.Open(privPath)
	if err != nil {
		log.Fatal(err)
	}
	privBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	fPub, err := os.Open(pubPath)
	if err != nil {
		log.Fatal(err)
	}
	pubBytes, err := ioutil.ReadAll(fPub)
	if err != nil {
		log.Fatal(err)
	}
	priv, pub := DecodeKeyPair(string(privBytes), string(pubBytes))
	return priv, pub, string(pubBytes)
}

func generateNewKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	c := elliptic.P384()
	priv, err := ecdsa.GenerateKey(c, rand.Reader)
	if err != nil {
		log.Fatal(err)
		return nil, nil
	}
	if !c.IsOnCurve(priv.PublicKey.X, priv.PublicKey.Y) {
		log.Fatal(err)
	}
	return priv, &priv.PublicKey
}

func EncodeKeyPair(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}

func DecodeKeyPair(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return privateKey, publicKey
}

func LoadPublicKey(keyDir string) *ecdsa.PublicKey {
	pubPath := path.Join(keyDir, pubKeyFile)
	f, err := os.Open(pubPath)
	if err != nil {
		log.Fatal(err)
	}
	pubBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	pub, err := DecodePublicKey(string(pubBytes))
	if err != nil {
		log.Fatal(err)
	}
	return pub
}

func DecodePublicKey(pemEncodedPub string) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	if blockPub != nil {
		x509EncodedPub := blockPub.Bytes
		genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
		publicKey := genericPublicKey.(*ecdsa.PublicKey)
		return publicKey, nil
	}
	return nil, errors.New("block pub is nil")
}
