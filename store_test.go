package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "mybestpictures"
	pathKey := CASPathTransformFunc(key)
	expectedFilename := "7037c790557f0d861c53d3bbd1fafe02dc3699e6"
	expectedPathName := "7037c/79055/7f0d8/61c53/d3bbd/1fafe/02dc3/699e6"
	if pathKey.PathName != expectedPathName {
		t.Errorf("got %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedFilename {
		t.Errorf("got %s want %s", pathKey.FileName, expectedFilename)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")
		if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}
		_, r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}
		b, _ := io.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("got %s want %s", b, data)
		}
		if err := s.Delete(key); err != nil {
			t.Error(err)
		}
		if ok := s.Has(key); ok {
			t.Errorf("expected to not have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
