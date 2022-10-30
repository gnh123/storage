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
