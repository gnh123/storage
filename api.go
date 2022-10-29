package storage

type Storager interface {
	Put(key int64, size int32, data []byte) (index int, err error)
	Get(key int64) (element Data, ok bool, err error)
	Delete(key int64) error
}
