package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/muhammadmahdiamirpour/distributed-file-system/crypto"
)

const DefaultRootDirName = "dfs-net"

// CASPathTransformFunc generates a hash-based path for content-addressable storage (CAS).
// It hashes the input key with SHA-1 and splits it into subdirectories for a hierarchical path structure.
//
// Parameters:
//   - key: A unique identifier for the content.
//
// Returns: A PathKey struct with `PathName` as a folder structure and `FileName` as the full hash.
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

// PathTransformFunc defines a function signature for transforming keys into paths.
type PathTransformFunc func(string) PathKey

// PathKey represents the storage path structure with `PathName` as the directory path and `FileName` as the final file.
type PathKey struct {
	PathName string
	FileName string
}

// FirstPathName returns the first directory in the `PathName`.
func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// FullPath returns the full directory path including the file name.
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

// StoreOpts configures options for creating a new storage instance.
//
// Fields:
//   - Root: Root directory for storage.
//   - PathTransformFunc: Function to transform keys to paths.
type StoreOpts struct {
	Root              string
	PathTransformFunc PathTransformFunc
}

// DefaultPathTransformFunc is the default path transformer, storing files without path splitting.
var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{PathName: key, FileName: key}
}

// Store represents a storage system with a specified path structure and encryption options.
type Store struct {
	StoreOpts
}

// NewStore initializes and returns a new Store instance with the given options.
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = DefaultRootDirName
	}
	return &Store{StoreOpts: opts}
}

// Has checks if a file with the specified key exists in the store.
//
// Parameters:
//   - id: An identifier to create a unique path.
//   - key: The key to locate the file.
//
// Returns: True if the file exists, false otherwise.
func (s *Store) Has(id string, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, fs.ErrNotExist)
}

// Clear deletes all files in the root storage directory.
func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

// Delete removes the file corresponding to the specified key from storage.
//
// Parameters:
//   - id: An identifier to create a unique path.
//   - key: The key to locate the file.
func (s *Store) Delete(id string, key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.FileName)
	}()
	firstPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FirstPathName())
	return os.RemoveAll(firstPathWithRoot)
}

// Write saves the contents from the reader to storage, creating directories if necessary.
//
// Parameters:
//   - id: Identifier for the storage path.
//   - key: Key for locating the file.
//   - R: Reader for the file contents.
//
// Returns: Number of bytes written and any errors.
func (s *Store) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

// WriteDecrypt saves encrypted content from the reader, decrypting it with the provided key.
//
// Parameters:
//   - encKey: Key for decrypting the content.
//   - id: Identifier for the storage path.
//   - key: Key for locating the file.
//   - r: Reader for the encrypted content.
//
// Returns: Number of bytes written and any errors.
func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	n, err := crypto.CopyDecrypt(encKey, r, f)
	if err != nil {
		return 0, err
	}
	return int64(n), err
}

// openFileForWriting prepares the file for writing, creating the necessary directories.
//
// Parameters:
//   - id: Identifier for the storage path.
//   - key: Key for locating the file.
//
// Returns: File handle and any errors.
func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	path := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	return os.Create(fullPath)
}

// writeStream copy content from the reader to storage.
//
// Parameters:
//   - id: Identifier for the storage path.
//   - key: Key for locating the file.
//   - R: Reader for the file contents.
//
// Returns: Number of bytes written and any errors.
func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

// Read retrieves the content corresponding to the specified key from storage.
//
// Parameters:
//   - id: Identifier for the storage path.
//   - key: Key for locating the file.
//
// Returns: File size, a reader for the file content, and any errors.
func (s *Store) Read(id string, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

// readStream opens a file for reading from storage.
//
// Parameters:
//   - id: Identifier for the storage path.
//   - key: Key for locating the file.
//
// Returns: File size, a reader for the file content, and any errors.
func (s *Store) readStream(id string, key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPath := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	file, err := os.Open(fullPath)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
}
