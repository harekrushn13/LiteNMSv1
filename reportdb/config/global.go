package config

type GlobalConfig struct {
	BaseDir string

	WriterCount uint8

	ReaderCount uint8

	CounterCount uint16

	ObjectCount uint32
}

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

func NewGlobalConfig() *GlobalConfig {

	return &GlobalConfig{
		BaseDir: "./src/storage/database/",

		WriterCount: 3,

		ReaderCount: 3,

		CounterCount: 3,

		ObjectCount: 5,
	}
}
