package crypt

import (
	"testing"
)

func TestMasterKeyEncryptDecryptRoundTrip(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	plaintext := []byte("dados confidenciais de teste")

	encrypted, err := EncryptWithMasterKey(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptWithMasterKey() error = %v", err)
	}

	decrypted, err := DecryptWithMasterKey(key, encrypted)
	if err != nil {
		t.Fatalf("DecryptWithMasterKey() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestMasterKeyDecryptRejectsTamperedCiphertext(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	encrypted, err := EncryptWithMasterKey(key, []byte("dado original"))
	if err != nil {
		t.Fatalf("EncryptWithMasterKey() error = %v", err)
	}

	tampered := append([]byte(nil), encrypted...)
	tampered[len(tampered)-1] ^= 0xFF

	if _, err := DecryptWithMasterKey(key, tampered); err == nil {
		t.Fatal("DecryptWithMasterKey() expected error for tampered ciphertext, got nil")
	}
}

func TestRotationKeyEncryptDecryptRoundTrip(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	plaintext := []byte("dados rotacionados de teste")

	encrypted, err := EncryptWithRotationKey(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptWithRotationKey() error = %v", err)
	}

	decrypted, err := DecryptWithRotationKey(key, encrypted)
	if err != nil {
		t.Fatalf("DecryptWithRotationKey() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestRotationKeyDecryptRejectsTamperedCiphertext(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	encrypted, err := EncryptWithRotationKey(key, []byte("dado original"))
	if err != nil {
		t.Fatalf("EncryptWithRotationKey() error = %v", err)
	}

	tampered := append([]byte(nil), encrypted...)
	tampered[len(tampered)-1] ^= 0xFF

	if _, err := DecryptWithRotationKey(key, tampered); err == nil {
		t.Fatal("DecryptWithRotationKey() expected error for tampered ciphertext, got nil")
	}
}

func TestGenerateRSAKeyPairRejectsSmallKeySize(t *testing.T) {
	if _, err := GenerateRSAKeyPair(1024); err == nil {
		t.Fatal("GenerateRSAKeyPair(1024) expected error for key size below minimum, got nil")
	}
}

func TestRSAKeyPairPEMRoundTrip(t *testing.T) {
	keyPair, err := GenerateRSAKeyPairDefault()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairDefault() error = %v", err)
	}

	privateKey, err := LoadRSAPrivateKeyFromPEM(keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("LoadRSAPrivateKeyFromPEM() error = %v", err)
	}

	publicKey, err := LoadRSAPublicKeyFromPEM(keyPair.PublicKey)
	if err != nil {
		t.Fatalf("LoadRSAPublicKeyFromPEM() error = %v", err)
	}

	plaintext := "mensagem de teste RSA"
	encrypted, err := HybridEncryptWithKeys(plaintext, publicKey)
	if err != nil {
		t.Fatalf("HybridEncryptWithKeys() error = %v", err)
	}

	decrypted, err := HybridDecryptWithKeys(encrypted, privateKey)
	if err != nil {
		t.Fatalf("HybridDecryptWithKeys() error = %v", err)
	}

	if string(decrypted) != plaintext {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestHybridEncryptDecryptRoundTrip(t *testing.T) {
	keyPair, err := GenerateRSAKeyPairDefault()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairDefault() error = %v", err)
	}

	publicKey, err := LoadRSAPublicKeyFromPEM(keyPair.PublicKey)
	if err != nil {
		t.Fatalf("LoadRSAPublicKeyFromPEM() error = %v", err)
	}

	privateKey, err := LoadRSAPrivateKeyFromPEM(keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("LoadRSAPrivateKeyFromPEM() error = %v", err)
	}

	plaintext := []byte("dados grandes o suficiente para justificar o modo híbrido")

	encrypted, err := HybridEncrypt(publicKey, plaintext)
	if err != nil {
		t.Fatalf("HybridEncrypt() error = %v", err)
	}

	decrypted, err := HybridDecrypt(privateKey, encrypted)
	if err != nil {
		t.Fatalf("HybridDecrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestLoadAESKeyFromPathRejectsMissingFile(t *testing.T) {
	if _, err := LoadAESKeyFromPath("/path/inexistente/master.key"); err == nil {
		t.Fatal("LoadAESKeyFromPath() expected error for missing file, got nil")
	}
}
