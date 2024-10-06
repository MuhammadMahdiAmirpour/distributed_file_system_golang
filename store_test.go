package main

import (
	"bytes"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "mybestpictures"
	pathName := CASPathTransformFunc(key)
	println(pathName)
	expectedPathName := "7037c/79055/7f0d8/61c53/d3bbd/1fafe/02dc3/699e6"
	if pathName != expectedPathName {
		t.Errorf("PathTransformFunc returned %s instead of %s", pathName, expectedPathName)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}
	s := NewStore(opts)
	data := bytes.NewBufferString("some jpg bytes")
	if err := s.writeStream("bestpictureever", data); err != nil {
		t.Error(err)
	}
}
