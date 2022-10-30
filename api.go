package storage

type Storager interface {
	Put(key int64, data []byte) (err error)
	Get(key int64) (element Data, ok bool, err error)
	GetSeq() int64
	Delete(key int64) error
	Close() error
}
