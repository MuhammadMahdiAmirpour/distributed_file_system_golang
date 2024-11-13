package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCopyEncryptDecrypt tests encryption and decryption using CopyEncrypt and CopyDecrypt.
func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not Bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := NewEncryptionKey()

	// Check if the key generation was successful.
	assert.NotNil(t, key, "Encryption key should not be nil")

	// Encrypt the payload.
	_, err := CopyEncrypt(key, src, dst)
	assert.Nil(t, err, "CopyEncrypt should not return an error")

	// Decrypt the payload.
	out := new(bytes.Buffer)
	nw, err := CopyDecrypt(key, dst, out)
	assert.Nil(t, err, "CopyDecrypt should not return an error")

	// Check if the decrypted content matches the original payload.
	assert.Equal(t, len(payload)+aes.BlockSize, nw, "Decrypted output size should match original size plus IV")
	assert.Equal(t, payload, out.String(), "Decrypted payload should match the original")
}

// TestCopyDecryptWithInvalidKey tests decryption with an incorrect key.
func TestCopyDecryptWithInvalidKey(t *testing.T) {
	payload := "Test message"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := NewEncryptionKey()

	// Encrypt the payload.
	_, err := CopyEncrypt(key, src, dst)
	assert.Nil(t, err, "CopyEncrypt should not return an error")

	// Use a different key for decryption.
	wrongKey := NewEncryptionKey()
	out := new(bytes.Buffer)
	_, err = CopyDecrypt(wrongKey, dst, out)

	// Instead of checking for an error, verify if the decrypted content is incorrect.
	// In this case, we expect the decrypted output to not match the original payload.
	if out.String() == payload {
		t.Error("Decryption with the wrong key should not produce the original payload")
	}
}

// TestGenerateID tests if the GenerateID function creates unique 32-byte hexadecimal IDs.
func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	// Check if the IDs are of correct length and unique.
	assert.NotEqual(t, id1, id2, "Generated IDs should be unique")
	assert.Equal(t, 64, len(id1), "Generated ID should be a 32-byte hex string (64 characters)")
	assert.Equal(t, 64, len(id2), "Generated ID should be a 32-byte hex string (64 characters)")
}

// TestHashKey tests the HashKey function to ensure it creates consistent MD5 hashes.
func TestHashKey(t *testing.T) {
	key := "mySecretKey"
	expectedHash := md5Hash("mySecretKey") // Generate expected hash using the same MD5 process.

	// Ensure the HashKey produces the correct MD5 hash.
	assert.Equal(t, expectedHash, HashKey(key), "HashKey should return the correct MD5 hash")
}

// md5Hash is a helper function to generate an MD5 hash for test comparison.
func md5Hash(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// TestNewEncryptionKey tests the generation of a random encryption key.
func TestNewEncryptionKey(t *testing.T) {
	key1 := NewEncryptionKey()
	key2 := NewEncryptionKey()

	// Ensure keys are generated and are not nil.
	assert.NotNil(t, key1, "NewEncryptionKey should not return nil")
	assert.NotNil(t, key2, "NewEncryptionKey should not return nil")

	// Ensure the keys are of correct length (32 bytes) and unique.
	assert.Equal(t, 32, len(key1), "Encryption key should be 32 bytes")
	assert.Equal(t, 32, len(key2), "Encryption key should be 32 bytes")
	assert.NotEqual(t, key1, key2, "Each generated encryption key should be unique")
}

// TestCopyEncryptDecryptWithEmptyPayload tests the encryption and decryption of an empty payload.
func TestCopyEncryptDecryptWithEmptyPayload(t *testing.T) {
	payload := ""
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := NewEncryptionKey()

	// Encrypt the empty payload.
	_, err := CopyEncrypt(key, src, dst)
	assert.Nil(t, err, "CopyEncrypt should not return an error for empty payload")

	// Decrypt the empty payload.
	out := new(bytes.Buffer)
	nw, err := CopyDecrypt(key, dst, out)
	assert.Nil(t, err, "CopyDecrypt should not return an error for empty payload")

	// Check if the decrypted content matches the original empty payload.
	assert.Equal(t, aes.BlockSize, nw, "Decrypted output size should match the size of IV for empty payload")
	assert.Equal(t, payload, out.String(), "Decrypted payload should match the original empty payload")
}
