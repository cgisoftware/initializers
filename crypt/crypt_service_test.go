package crypt

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeAESKeyFile(t *testing.T, dir, name string) string {
	t.Helper()

	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(hex.EncodeToString(key)), 0o600); err != nil {
		t.Fatalf("failed to write AES key file: %v", err)
	}

	return path
}

func writeRSAKeyFiles(t *testing.T, dir string) (privatePath, publicPath string) {
	t.Helper()

	keyPair, err := GenerateRSAKeyPairDefault()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairDefault() error = %v", err)
	}

	privatePath = filepath.Join(dir, "rsa_private.pem")
	if err := os.WriteFile(privatePath, []byte(keyPair.PrivateKey), 0o600); err != nil {
		t.Fatalf("failed to write RSA private key file: %v", err)
	}

	publicPath = filepath.Join(dir, "rsa_public.pem")
	if err := os.WriteFile(publicPath, []byte(keyPair.PublicKey), 0o644); err != nil {
		t.Fatalf("failed to write RSA public key file: %v", err)
	}

	return privatePath, publicPath
}

func newTestCryptService(t *testing.T) CryptService {
	t.Helper()

	dir := t.TempDir()
	privatePath, publicPath := writeRSAKeyFiles(t, dir)
	masterKeyPath := writeAESKeyFile(t, dir, "master.key")
	rotationKeyPath := writeAESKeyFile(t, dir, "rotation.key")

	service, err := Initialize(privatePath, publicPath, masterKeyPath, rotationKeyPath)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	return service
}

func TestInitializeRequiresAllKeyPaths(t *testing.T) {
	dir := t.TempDir()
	privatePath, publicPath := writeRSAKeyFiles(t, dir)
	masterKeyPath := writeAESKeyFile(t, dir, "master.key")
	rotationKeyPath := writeAESKeyFile(t, dir, "rotation.key")

	tests := []struct {
		name               string
		rsaPrivateKeyPath  string
		rsaPublicKeyPath   string
		aesMasterKeyPath   string
		aesRotationKeyPath string
	}{
		{"rsa private ausente", "", publicPath, masterKeyPath, rotationKeyPath},
		{"rsa public ausente", privatePath, "", masterKeyPath, rotationKeyPath},
		{"aes master ausente", privatePath, publicPath, "", rotationKeyPath},
		{"aes rotation ausente", privatePath, publicPath, masterKeyPath, ""},
		{"todos ausentes", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Initialize(tt.rsaPrivateKeyPath, tt.rsaPublicKeyPath, tt.aesMasterKeyPath, tt.aesRotationKeyPath); err == nil {
				t.Fatal("Initialize() expected error, got nil")
			}
		})
	}
}

func TestInitializeRejectsMissingKeyFile(t *testing.T) {
	dir := t.TempDir()
	_, publicPath := writeRSAKeyFiles(t, dir)
	masterKeyPath := writeAESKeyFile(t, dir, "master.key")
	rotationKeyPath := writeAESKeyFile(t, dir, "rotation.key")

	if _, err := Initialize(filepath.Join(dir, "nao_existe.pem"), publicPath, masterKeyPath, rotationKeyPath); err == nil {
		t.Fatal("Initialize() expected error for missing private key file, got nil")
	}
}

func TestCryptServiceEncryptDecryptDataRoundTrip(t *testing.T) {
	service := newTestCryptService(t)

	data := "dados sensíveis via CryptService"
	encrypted, err := service.EncryptData(data)
	if err != nil {
		t.Fatalf("EncryptData() error = %v", err)
	}

	decrypted, err := service.DecryptData(encrypted)
	if err != nil {
		t.Fatalf("DecryptData() error = %v", err)
	}

	if string(decrypted) != data {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, data)
	}
}

func TestCryptServiceMasterKeySimpleRoundTrip(t *testing.T) {
	service := newTestCryptService(t)

	data := "outro dado sensível"
	encrypted, err := service.EncryptWithMasterKeySimple(data)
	if err != nil {
		t.Fatalf("EncryptWithMasterKeySimple() error = %v", err)
	}

	decrypted, err := service.DecryptWithMasterKeySimple(encrypted)
	if err != nil {
		t.Fatalf("DecryptWithMasterKeySimple() error = %v", err)
	}

	if string(decrypted) != data {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, data)
	}
}

func TestNewCryptManagerPasswordRoundTrip(t *testing.T) {
	service := newTestCryptService(t)
	manager := NewCryptManager(service)

	password := "minhaSenhaSegura123!"
	encrypted, err := manager.EncryptPassword(password)
	if err != nil {
		t.Fatalf("EncryptPassword() error = %v", err)
	}

	decrypted, err := manager.DecryptPassword(encrypted)
	if err != nil {
		t.Fatalf("DecryptPassword() error = %v", err)
	}

	if string(decrypted) != password {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, password)
	}
}

func TestNewCryptManagerSensitiveDataRoundTrip(t *testing.T) {
	service := newTestCryptService(t)
	manager := NewCryptManager(service)

	data := "CPF: 123.456.789-00"
	encrypted, err := manager.EncryptSensitiveData(data)
	if err != nil {
		t.Fatalf("EncryptSensitiveData() error = %v", err)
	}

	decrypted, err := manager.DecryptSensitiveData(encrypted)
	if err != nil {
		t.Fatalf("DecryptSensitiveData() error = %v", err)
	}

	if string(decrypted) != data {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, data)
	}
}

func TestGenerateTokenDecryptTokenRoundTrip(t *testing.T) {
	ctx := context.Background()
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	data := []byte("payload do token")
	token, err := GenerateToken(ctx, key, data)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	decrypted, err := DecryptToken(ctx, key, token)
	if err != nil {
		t.Fatalf("DecryptToken() error = %v", err)
	}

	if string(decrypted) != string(data) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, data)
	}
}

func TestDecryptTokenRejectsInvalidFormat(t *testing.T) {
	ctx := context.Background()
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	if _, err := DecryptToken(ctx, key, "token-sem-formato-esperado-com-tres-partes"); err == nil {
		t.Fatal("DecryptToken() expected error for malformed token, got nil")
	}
}

func TestDecryptTokenRejectsTamperedToken(t *testing.T) {
	ctx := context.Background()
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	token, err := GenerateToken(ctx, key, []byte("dado original"))
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		t.Fatalf("unexpected token format: %q", token)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("failed to decode ciphertext: %v", err)
	}
	ciphertext[len(ciphertext)-1] ^= 0xFF

	tamperedToken := base64.StdEncoding.EncodeToString(ciphertext) + "-" + parts[1]

	if _, err := DecryptToken(ctx, key, tamperedToken); err == nil {
		t.Fatal("DecryptToken() expected error for tampered token, got nil")
	}
}
