package main

import (
	"bytes"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "mybestpictures"
	pathKey := CASPathTransformFunc(key)
	expectedOriginal := "7037c790557f0d861c53d3bbd1fafe02dc3699e6"
	expectedPathName := "7037c/79055/7f0d8/61c53/d3bbd/1fafe/02dc3/699e6"
	if pathKey.PathName != expectedPathName {
		t.Errorf("got %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedOriginal {
		t.Errorf("got %s want %s", pathKey.FileName, expectedOriginal)
	}
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "mybestpictures"
	data := []byte("some jpg bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "mybestpictures"
	data := []byte("some jpg bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	if ok := s.Has(key); !ok {
		t.Errorf("expected to have key %s", key)
	}
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("got %s want %s", b, data)
	}
	s.Delete(key)
}
