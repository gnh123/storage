package storage

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试初始化函数
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

// 测试put get函数，put进去，能get出来
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

func Test_PutAndGet(t *testing.T) {
	os.Remove("./testdata/11.dat")
	os.Remove("./testdata/11.idx")
	os.Remove("./testdata/11.meta")

	index, err := newIndexInMemory("./testdata/11")

	assert.NoError(t, err)
	assert.NotEqual(t, index, nil)

	for i := int64(0); i < 100; i++ {

		err = index.Put(i, []byte(fmt.Sprintf("hello world:%d", i)))
		assert.NoError(t, err)
	}

	for i := int64(0); i < 100; i++ {

		elem, ok, err := index.Get(i)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, elem.Data, []byte(fmt.Sprintf("hello world:%d", i)))
	}
}

// put delete get, 删除之后不能get出来
func Test_PutDeleteGet_Once(t *testing.T) {

	os.Remove("./testdata/2.dat")
	os.Remove("./testdata/2.idx")
	os.Remove("./testdata/2.meta")

	i, err := newIndexInMemory("./testdata/2")

	assert.NoError(t, err)
	assert.NotEqual(t, i, nil)

	err = i.Put(0, []byte("hello world"))
	assert.NoError(t, err)

	elem, ok, err := i.Get(0)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, elem.Data, []byte("hello world"))

	i.Delete(0)
	elem, ok, err = i.Get(0)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.NotEqual(t, elem.Data, []byte("hello world"))
}
