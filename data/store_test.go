package data

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"

	pathKey := CASPathTransformFun(key)
	expectedOriginalKey := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathname := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	assert.Equal(t, pathKey.Filename, expectedOriginalKey)
	assert.Equal(t, pathKey.Pathname, expectedPathname)
}

func TestDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFun,
	}

	s := NewStore(opts)
	key := "momspecials"
	data := []byte("some jpg bytes")
	assert.Nil(t, s.writeStream(key, bytes.NewReader(data)))
	assert.Nil(t, s.Delete(key))
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	key := "momspecials"
	data := []byte("some jpg bytes")
	assert.Nil(t, s.writeStream(key, bytes.NewReader(data)))

	assert.True(t, s.Has(key))
	r, err := s.Read(key)
	assert.Nil(t, err)

	b, _ := io.ReadAll(r)
	assert.Equal(t, string(data), string(b))

	assert.Nil(t, s.Delete(key))
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFun,
	}

	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	assert.Nil(t, s.Clear())
}
