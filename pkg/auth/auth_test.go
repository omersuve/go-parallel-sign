package auth

import (
	"crypto/rsa"
	"testing"
)

func TestGenerateKeys(t *testing.T) {
	keys, err := GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys() failed: %v", err)
	}
	if keys.PrivateKey == nil {
		t.Errorf("GenerateKeys() produced nil PrivateKey")
	}
	if keys.PublicKey == nil {
		t.Errorf("GenerateKeys() produced nil PublicKey")
	}
	if keys.PublicKey != &keys.PrivateKey.PublicKey {
		t.Errorf("GenerateKeys() PublicKey does not match PrivateKey's public key")
	}
	if keys.PrivateKey.N.BitLen() != 2048 {
		t.Errorf("PrivateKey size = %d bits, want 2048", keys.PrivateKey.N.BitLen())
	}
}

func TestSignAndVerify(t *testing.T) {
	keys, err := GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys() failed: %v", err)
	}

	num := int32(42)

	sig, err := Sign(num, keys.PrivateKey)
	if err != nil {
		t.Fatalf("Sign(%d) failed: %v", num, err)
	}
	if len(sig) != 256 {
		t.Errorf("Sign(%d) signature length = %d, want 256", num, len(sig))
	}

	if !Verify(num, sig, keys.PublicKey) {
		t.Errorf("Verify(%d, valid signature) = false, want true", num)
	}

	if Verify(43, sig, keys.PublicKey) {
		t.Errorf("Verify(43, signature for %d) = true, want false", num)
	}

	badSig := make([]byte, 256)
	for i := range badSig {
		badSig[i] = 0xff
	}
	if Verify(num, badSig, keys.PublicKey) {
		t.Errorf("Verify(%d, invalid signature) = true, want false", num)
	}
}

func TestPublicKey2BytesAndParsePublicKey(t *testing.T) {
	keys, err := GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys() failed: %v", err)
	}

	pubBytes, err := PublicKey2Bytes(keys.PublicKey)
	if err != nil {
		t.Fatalf("PublicKey2Bytes() failed: %v", err)
	}
	if len(pubBytes) == 0 {
		t.Errorf("PublicKey2Bytes() produced empty bytes")
	}

	parsedPub, err := ParsePublicKey(pubBytes)
	if err != nil {
		t.Fatalf("ParsePublicKey() failed: %v", err)
	}
	if parsedPub.N.Cmp(keys.PublicKey.N) != 0 || parsedPub.E != keys.PublicKey.E {
		t.Errorf("ParsePublicKey() returned different key: got N=%v E=%d, want N=%v E=%d",
			parsedPub.N, parsedPub.E, keys.PublicKey.N, keys.PublicKey.E)
	}

	_, err = ParsePublicKey([]byte("invalid"))
	if err == nil {
		t.Errorf("ParsePublicKey(invalid) did not fail, want error")
	}
}

func TestSignWithInvalidKey(t *testing.T) {
	num := int32(42)

	// Test with nil private key
	_, err := Sign(num, nil)
	if err == nil {
		t.Errorf("Sign(%d) with nil key did not fail, want error", num)
	} else if err.Error() != "invalid private key: nil or uninitialized" {
		t.Errorf("Sign(%d) with nil key returned wrong error: got %v, want 'invalid private key: nil or uninitialized'", num, err)
	}

	// Test with uninitialized private key
	_, err = Sign(num, &rsa.PrivateKey{})
	if err == nil {
		t.Errorf("Sign(%d) with uninitialized key did not fail, want error", num)
	} else if err.Error() != "invalid private key: nil or uninitialized" {
		t.Errorf("Sign(%d) with uninitialized key returned wrong error: got %v, want 'invalid private key: nil or uninitialized'", num, err)
	}
}