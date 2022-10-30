package storage

import (
	"errors"
	"fmt"
	"os"
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
	ErrDirName    = errors.New("dir name is empty")
)

// 一组里面有多个引擎，一个引擎只存储32GB
type Group struct {
	datArr []Storager //一个组下面有多个存储引擎
	next   int32
}

func dirName(dir string) string {
	if !strings.HasSuffix(dir, "/") {
		return dir + "/"
	}
	return dir
}

// 加载
func loadOrNewGroup(dir string, max Size) (g *Group, err error) {
	// 检查dir否为空, 如果返回错误
	if len(dir) == 0 {
		return nil, ErrDirName
	}

	// 检查目录是否存在
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		// 不存在新建一个
		os.Mkdir(dir, 0755)
	}

	// 一个dat最多到32GB，计算可以创建多少个
	count := max / maxDatLimit
	if count == 0 {
		count = 1
	}

	g = &Group{}
	g.datArr = make([]Storager, count)

	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	for i := range g.datArr {
		var idx *IndexInMemory
		idx, err = newIndexInMemory(fmt.Sprintf("%s%d", dirName(dir), i))
		if err != nil {
			return
		}
		g.datArr[i] = idx
	}

	return
}

func (g *Group) Put(data []byte) (index string, err error) {

	groupIndex := atomic.LoadInt32(&g.next)
	var idx int
	for groupIndex < int32(len(g.datArr)) {
		groupIndex = atomic.LoadInt32(&g.next)
		key := g.datArr[groupIndex].GetSeq()
		idx, err = g.datArr[groupIndex].Put(key, data)
		if err != nil {
			if errors.Is(err, ErrFull) {
				// 只有一个go程可以安全修改
				atomic.CompareAndSwapInt32(&g.next, groupIndex, groupIndex+1)
				continue
			}

			if err != nil {
				return
			}
		}
		break
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
		err = fmt.Errorf("%w %s", ErrIllegalKey, err)
		return
	}

	idxStr := key[pos+1:]
	idx, err = strconv.Atoi(idxStr)
	if err != nil {
		err = fmt.Errorf("%w %s", ErrIllegalKey, err)
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

// 关闭所有索引
func (g *Group) Close() (err error) {
	for _, s := range g.datArr {
		if s == nil {
			continue
		}

		if err = s.Close(); err != nil {
			return err
		}
	}
	return
}
