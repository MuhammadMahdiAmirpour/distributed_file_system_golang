package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// GenerateID creates a unique 32-byte hexadecimal identifier by generating random bytes and encoding them.
// Returns:
//   - A unique ID string or an empty string if an error occurs during random byte generation.
func GenerateID() string {
	buf := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

// HashKey creates an MD5 hash of the given key string and returns it as a hexadecimal string.
// Parameters:
//   - key: The string to be hashed.
//
// Returns:
//   - The MD5 hash of the input key as a hexadecimal string.
func HashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

// NewEncryptionKey generates a 32-byte random encryption key for AES encryption.
// Returns:
//   - A byte slice containing the generated encryption key or nil if an error occurs during key generation.
func NewEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, keyBuf)
	if err != nil {
		fmt.Printf("%s", err)
		return nil
	}
	return keyBuf
}

// copyStream performs encrypted copying from the source reader to the destination writer.
// It uses the provided cipher.Stream and encrypts/decrypts data as it reads from src and writes to dst.
// Parameters:
//   - stream: The cipher.Stream used for encryption or decryption.
//   - blockSize: The size of the cipher block, used to initialize the written bytes count.
//   - src: The data source to be copied from.
//   - dst: The data destination to be copied to.
//
// Returns:
//   - The total number of bytes written or an error if writing fails during processing.
func copyStream(stream cipher.Stream, blockSize int, src io.Reader, dst io.Writer) (int, error) {
	var (
		buf = make([]byte, 32*1024) // Buffer for copying data in chunks
		nw  = blockSize             // Initial byte count set to the block size
	)
	for {
		n, err := src.Read(buf) // Read data into buffer
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n]) // Apply XOR for encryption/decryption
			nn, err := dst.Write(buf[:n])     // Write processed data to destination
			if err != nil {
				return 0, err
			}
			nw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, nil
}

// CopyDecrypt decrypts data from src and writes the decrypted data to dst using AES-CTR mode.
// Parameters:
//   - key: The AES encryption key for decryption.
//   - src: The source from which encrypted data is read.
//   - dst: The destination where decrypted data will be written.
//
// Returns:
//   - The total number of bytes written or an error if decryption or writing fails.
func CopyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}
	stream := cipher.NewCTR(block, iv)
	return copyStream(stream, block.BlockSize(), src, dst)
}

// CopyEncrypt encrypts data from src and writes the encrypted data to dst using AES-CTR mode.
// It prepends the IV (initialization vector) to the output for use in decryption.
// Parameters:
//   - key: The AES encryption key for encryption.
//   - src: The source from which plain data is read.
//   - dst: The destination where encrypted data will be written.
//
// Returns:
//   - The total number of bytes written or an error if encryption or writing fails.
func CopyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}
	// Prepend the IV to the encrypted output
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}
	stream := cipher.NewCTR(block, iv)
	return copyStream(stream, block.BlockSize(), src, dst)
}
