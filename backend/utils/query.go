package utils

type QueryRequest struct {
	CounterID uint16 `msgpack:"counter_id" json:"counter_id" binding:"required"`

	ObjectIDs []uint32 `msgpack:"object_ids" json:"object_ids,omitempty"`

	From uint32 `msgpack:"from" json:"from" binding:"required"`

	To uint32 `msgpack:"to" json:"to" binding:"required"`

	Aggregation string `msgpack:"aggregation" json:"aggregation,omitempty" binding:"required"`

	GroupByObjects bool `msgpack:"group_by_objects" json:"group_by_objects,omitempty"`

	Interval int `msgpack:"interval" json:"interval,omitempty"`
}

type QueryMap struct {
	RequestID uint64 `json:"request_id"`

	QueryRequest QueryRequest `json:"query_request"`

	Response chan Response
}

type QuerySend struct {
	RequestID uint64 `msgpack:"request_id" json:"request_id"`

	QueryRequest QueryRequest `msgpack:"query_request" json:"query_request"`
}

type Response struct {
	RequestID uint64 `msgpack:"request_id" json:"request_id"`

	Error string `msgpack:"error,omitempty" json:"error,omitempty"`

	Data interface{} `msgpack:"data" json:"data"`
}

type DataPoint struct {
	Timestamp uint32 `json:"timestamp"`

	Value interface{} `json:"value"`
}
