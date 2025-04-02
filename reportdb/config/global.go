package config

const (
	BaseDir = "./src/storage/database"

	Partitions uint8 = 3

	WriterCount uint8 = 3

	CounterCount uint16 = 3

	ObjectCount uint32 = 5

	FileGrowth int64 = 64 // byte
)

type DataType uint8

const (
	TypeUint64 DataType = iota + 1

	TypeFloat64

	TypeString
)

var CounterTypeMapping = map[uint16]DataType{

	1: TypeUint64,

	2: TypeFloat64,

	3: TypeString,
}
