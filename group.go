package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
)

type Size int64

const (
	KB Size = 1024
	MB      = KB * 1024
	GB      = MB * 1024
	TB      = 1024 * GB
)

var (
	ErrIllegalKey = errors.New("Illegal key")
)

// 一组里面有多个引擎，一个引擎只存储32GB
type Group struct {
	datArr []Storager
	next   int32
}

func loadOrNewGroup(dir string, max Size) (g *Group, err error) {
	count := max / maxDatLimit
	if count == 0 {
		count = 1
	}

	g.datArr = make([]Storager, count)
	return
}

func (g *Group) Put(data []byte) (index string, err error) {

	groupIndex := atomic.LoadInt32(&g.next)
	var idx int
	for ; groupIndex < int32(len(g.datArr)); groupIndex++ {
		key := g.datArr[groupIndex].GetSeq()
		idx, err = g.datArr[groupIndex].Put(key, data)
		if err != nil && errors.Is(err, ErrFull) {
			continue
		}
	}

	if err != nil {
		return
	}

	return fmt.Sprintf("%d,%d", groupIndex, idx), nil
}

func (g *Group) checkIndex(key string) (groupIndex int, idx int, err error) {

	pos := strings.Index(key, ",")
	if pos == -1 {
		err = ErrIllegalKey
		return
	}

	groupIndexStr := key[:pos]
	groupIndex, err = strconv.Atoi(groupIndexStr)
	if err != nil {
		err = fmt.Errorf("%w %w", ErrIllegalKey, err)
		return
	}

	idxStr := key[pos+1:]
	idx, err = strconv.Atoi(idxStr)
	if err != nil {
		err = fmt.Errorf("%w %w", ErrIllegalKey, err)
		return
	}

	if groupIndex >= len(g.datArr) {
		err = fmt.Errorf("%w, groupIndex:%d > len(g.datArr:%d)", ErrIllegalKey, groupIndex, len(g.datArr))
		return
	}
	return
}

func (g *Group) Get(key string) (element Data, ok bool, err error) {
	groupIndex, idx, err := g.checkIndex(key)
	if err != nil {
		return
	}
	return g.datArr[groupIndex].Get(int64(idx))
}

func (g *Group) Delete(key string) (err error) {
	groupIndex, idx, err := g.checkIndex(key)
	if err != nil {
		return
	}
	return g.datArr[groupIndex].Delete(int64(idx))
}

func (g *Group) Close() (err error) {
	for _, s := range g.datArr {
		if err = s.Close(); err != nil {
			return err
		}
	}
	return
}
