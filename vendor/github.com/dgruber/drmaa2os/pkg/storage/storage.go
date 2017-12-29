package storage

//go:generate stringer -type=KeyType
type KeyType int

const (
	JobSessionType KeyType = iota
	ReservationSessionType
)

func (k KeyType) String() string {
	switch k {
	case JobSessionType:
		return "JobSessionType"
	default:
		return "ReservationSessionType"
	}
}

type Storer interface {
	Init() error
	Put(t KeyType, key, value string) error
	Get(t KeyType, key string) (string, error)
	List(t KeyType) ([]string, error)
	Exists(t KeyType, key string) bool
	Delete(t KeyType, key string) error
	Exit() error
}
