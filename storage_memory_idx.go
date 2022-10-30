package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"

	"github.com/antlabs/deepcopy"
	"google.golang.org/protobuf/proto"
)

// 索引组成
// 4个字节的payload长度
// 8个字节的key
// 8个字节的offset
// 4个字节的size
// 4个字节的crc32
// 8个字节的过期时间

var (
	_            Storager = (*IndexInMemory)(nil)
	defaultTable          = crc32.MakeTable(0xD5828281)
	maxDatLimit           = 32 * GB
	payload               = 4
	ErrFull               = errors.New("The space is full")
)

type Index struct {
	Key     int32  `protobuf:"varint,1,opt,name=key,proto3" json:"key,omitempty"`         //返回给客户端的值
	Size    int32  `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`       //大小
	Offset  int64  `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`   //偏移量
	Timeout int64  `protobuf:"varint,4,opt,name=timeout,proto3" json:"timeout,omitempty"` //超时时间
	Crc32   uint32 `protobuf:"varint,5,opt,name=crc32,proto3" json:"crc32,omitempty"`     //crc32校验和
}

type Data struct {
	Index
	Data []byte
}

type metadata struct {
	// 下个索引值, 需要持久化到文件中
	Seq int64
	// 当前记录总字节数, 限制单个索引能管理的最大数据，需要持久化到文件中
	TotalSize int64
	//被删除文件的个数,需要持久化到文件中
	DeleteCount int
	// 文件总数, 被删除的也计算在内, 需要持久化到文件中
	FileCount int
	// false代表可写, 需要持久化到文件中
	Readonly bool
	//最后一个offset, 需要持久化到文件中
	DatOffset int64
}

// 一个内存索引管理32GB文件
type IndexInMemory struct {
	idx *os.File //索引文件
	dat *os.File //数据文件
	md  *os.File //元数据文件
	// 读写锁
	rwmu sync.RWMutex

	//最后一个索引数据的offset
	idxOffset int64

	metadata

	// sync.Map没有Len比较蛋疼，所以这里还是map+读写锁
	allIndex map[int64]Index
}

func idxName(fileName string) string {
	return fmt.Sprintf("%s.idx", fileName)
}

func datName(fileName string) string {
	return fmt.Sprintf("%s.dat", fileName)
}

func newIndexInMemory(fileName string) (idx *IndexInMemory, err error) {
	var memIndex IndexInMemory

	// 打开并加载索引文件
	if err = memIndex.loadIdx(fileName); err != nil {
		return nil, err
	}

	if err = memIndex.loadDat(fileName); err != nil {
		return nil, err
	}
	return &memIndex, nil
}

func (i *IndexInMemory) loadIdx(idxName string) (err error) {
	// 打开索引文件
	i.idx, err = os.OpenFile(idxName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	head := uint32(0)
	defer func() {
		if err == io.EOF {
			err = nil
		}
	}()

	for {

		err = binary.Read(i.idx, binary.LittleEndian, &head)
		if err != nil {
			return
		}

		buf := make([]byte, head)
		_, err = i.idx.ReadAt(buf, i.idxOffset)
		if err != nil {
			return err
		}

		var index IdxVersion0
		if err = proto.Unmarshal(buf, &index); err != nil {
			return err
		}

		var index2 Index
		if err = deepcopy.Copy(&index2, &index).Do(); err != nil {
			return err
		}
		i.idxOffset += int64(head)
		i.allIndex[int64(index2.Key)] = index2
	}
}

func (i *IndexInMemory) loadDat(datName string) (err error) {
	i.dat, err = os.OpenFile(datName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	return
}

func (i *IndexInMemory) checkFull() (err error) {

	i.rwmu.Lock()
	i.rwmu.Unlock()

	if i.Readonly {
		err = ErrFull
		return
	}

	if i.TotalSize >= int64(maxDatLimit) {
		i.Readonly = true
	}

	return ErrFull
}

func (i *IndexInMemory) GetSeq() (key int64) {
	i.rwmu.Lock()
	key = i.Seq
	i.Seq++
	i.rwmu.Unlock()
	return
}

// 保存
func (i *IndexInMemory) Put(key int64, data []byte) (index int, err error) {
	if err := i.checkFull(); err != nil {
		return 0, err
	}

	// TODO sync.Pool
	var buf bytes.Buffer

	crc := crc32.Update(0, defaultTable, data)

	idx := IdxVersion0{}
	idx.Key = key
	idx.Size = int32(len(data))
	idx.Offset = i.DatOffset
	idx.Crc32 = crc

	all, err := proto.Marshal(&idx)
	if err != nil {
		return 0, err
	}

	if err := binary.Write(&buf, binary.LittleEndian, int32(len(all))); err != nil {
		return 0, err
	}

	i.rwmu.Lock()
	// 1. 写入索引文件
	n, err := i.idx.Write(buf.Bytes())
	if err != nil {
		i.rwmu.Unlock()
		return 0, err
	}

	// 3. 更新数据文件
	if _, err := i.dat.WriteAt(data, i.DatOffset); err != nil {
		i.idx.Truncate(i.idxOffset) //修改文件指针的大小
		i.idx.Seek(int64(-n), 1)    //修改文件偏移指针
		i.rwmu.Unlock()
		return 0, err
	}
	// 2. 更新offset
	i.DatOffset += int64(len(data))
	i.idxOffset += int64(len(all)) + 4

	var idxMem Index
	deepcopy.Copy(&idxMem, &idx).Do()
	i.allIndex[i.Seq] = idxMem
	i.Seq++
	i.FileCount++
	i.rwmu.Unlock()
	return 0, nil
}

// 获取
func (i *IndexInMemory) Get(key int64) (element Data, ok bool, err error) {
	i.rwmu.RLock()
	element.Index, ok = i.allIndex[key]
	if !ok {
		i.rwmu.Unlock()
		return
	}

	// TODO sync.Pool
	element.Data = make([]byte, element.Size)
	i.idx.ReadAt(element.Data, element.Offset)
	crc := crc32.Checksum(element.Data, defaultTable)
	if crc != element.Crc32 {
		err = fmt.Errorf("The data file is bad:key(%d)\n", key)
		i.rwmu.RUnlock()
		return
	}

	i.rwmu.RUnlock()
	return
}

// 删除
func (i *IndexInMemory) Delete(key int64) error {
	i.rwmu.Lock()
	delete(i.allIndex, key)
	i.DeleteCount++
	i.rwmu.Unlock()
	return nil
}

// close
func (i *IndexInMemory) Close() (err error) {
	i.rwmu.Lock()
	defer i.rwmu.Unlock()

	if err = i.idx.Sync(); err != nil {
		return err
	}

	if err = i.dat.Sync(); err != nil {
		return err
	}
	if err = i.idx.Close(); err != nil {
		return err
	}

	return i.dat.Close()
}
