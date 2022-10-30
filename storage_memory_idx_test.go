package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewIndexInMemory(t *testing.T) {

	i, err := newIndexInMemory("./testdata/0")
	assert.NoError(t, err)
	assert.NotEqual(t, i, nil)

	_, err = os.Stat("./testdata/0.dat")
	assert.True(t, err == nil || os.IsExist(err))
	_, err = os.Stat("./testdata/0.idx")
	assert.True(t, err == nil || os.IsExist(err))
	_, err = os.Stat("./testdata/0.meta")
	assert.True(t, err == nil || os.IsExist(err))
}

func Test_PutAndGet_Once(t *testing.T) {
	os.Remove("./testdata/1.dat")
	os.Remove("./testdata/1.idx")
	os.Remove("./testdata/1.meta")

	i, err := newIndexInMemory("./testdata/1")

	assert.NoError(t, err)
	assert.NotEqual(t, i, nil)

	err = i.Put(0, []byte("hello world"))
	assert.NoError(t, err)

	elem, ok, err := i.Get(0)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, elem.Data, []byte("hello world"))
}
