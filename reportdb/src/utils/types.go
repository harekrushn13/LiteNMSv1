package utils

type Events struct {
	ObjectId uint32 `msgpack:"objectId" json:"objectId"`

	CounterId uint16 `msgpack:"counterId" json:"counterId"`

	Timestamp uint32 `msgpack:"timestamp" json:"timestamp"`

	Value interface{} `msgpack:"value" json:"value"`
}

type DataPoint struct {
	Timestamp uint32 `json:"timestamp"`

	Value interface{} `json:"value"`
}

type QueryReceive struct {
	RequestID uint64 `msgpack:"request_id" json:"request_id"`

	Query Query `msgpack:"query_request" json:"query_request"`
}

type Query struct {
	CounterID uint16 `msgpack:"counter_id" json:"counter_id"`

	ObjectIDs []uint32 `msgpack:"object_ids" json:"object_ids"`

	From uint32 `msgpack:"from" json:"from"`

	To uint32 `msgpack:"to" json:"to"`

	Aggregation string `msgpack:"aggregation" json:"aggregation"`

	GroupByObjects bool `msgpack:"group_by_objects" json:"group_by_objects"`

	Interval int `msgpack:"interval" json:"interval"`
}

type Response struct {
	RequestID uint64 `msgpack:"request_id" json:"request_id"`

	Error string `msgpack:"error,omitempty" json:"error,omitempty"`

	Data interface{} `msgpack:"data" json:"data"`
}
