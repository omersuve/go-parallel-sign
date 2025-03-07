package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
)

// Holds the RSA key pair for a client
type ClientKeys struct {
    PrivateKey *rsa.PrivateKey
    PublicKey  *rsa.PublicKey
}

// Creates a new RSA key pair
func GenerateKeys() (*ClientKeys, error) {
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048) // 2048-bit key
    if err != nil {
        return nil, err
    }
    return &ClientKeys{
        PrivateKey: privateKey,
        PublicKey:  &privateKey.PublicKey,
    }, nil
}

// Signs the number with the clients private key
func Sign(num int32, priv *rsa.PrivateKey) ([]byte, error) {
	if priv == nil || priv.N == nil {
        return nil, errors.New("invalid private key: nil or uninitialized")
    }
    // Convert int32 to bytes (e.g., big-endian)
    msg := make([]byte, 4)
    binary.BigEndian.PutUint32(msg, uint32(num))
    
    // Hash the message with SHA-256
    hash := sha256.Sum256(msg)
    
    // Sign the hash
    signature, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
    if err != nil {
        return nil, err
    }
    return signature, nil
}

// Verifies the signature using the clients public key
func Verify(num int32, signature []byte, pub *rsa.PublicKey) bool {
    // Convert int32 to bytes (same as Sign)
    msg := make([]byte, 4)
    binary.BigEndian.PutUint32(msg, uint32(num))
    
    // Hash the message
    hash := sha256.Sum256(msg)
    
    // Verify the signature
    err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], signature)
    return err == nil
}

// Serializes the public key to bytes
func PublicKey2Bytes(pub *rsa.PublicKey) ([]byte, error) {
    pubBytes, err := x509.MarshalPKIXPublicKey(pub)
    if err != nil {
        return nil, err
    }
    return pem.EncodeToMemory(&pem.Block{
        Type:  "PUBLIC KEY",
        Bytes: pubBytes,
    }), nil
}

// Deserializes the public key from bytes
func ParsePublicKey(pubBytes []byte) (*rsa.PublicKey, error) {
    block, _ := pem.Decode(pubBytes)
    if block == nil {
        return nil, errors.New("failed to decode PEM block")
    }
    pub, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, err
    }
    rsaPub, ok := pub.(*rsa.PublicKey)
    if !ok {
        return nil, errors.New("not an RSA public key")
    }
    return rsaPub, nil
}