package storage

type Storage struct {
	*Group
}

func Open(name string, max Size) (s Storage, err error) {
	s.Group, err = loadOrNewGroup(name, max)
	return
}
