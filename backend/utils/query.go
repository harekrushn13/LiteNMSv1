package utils

type QueryRequest struct {
	CounterID uint16 `json:"counter_id" binding:"required"`

	ObjectIDs []uint32 `json:"object_ids,omitempty" binding:"required"`

	From uint32 `json:"from" binding:"required"`

	To uint32 `json:"to" binding:"required"`

	Aggregation string `json:"aggregation,omitempty" binding:"required"`

	GroupByObjects bool `json:"group_by_objects,omitempty"`

	Interval int `json:"interval,omitempty"`
}

type QueryMap struct {
	RequestID uint64 `json:"request_id"`

	QueryRequest QueryRequest `json:"query_request"`

	Response chan Response
}

type QuerySend struct {
	RequestID uint64 `json:"request_id"`

	QueryRequest QueryRequest `json:"query_request"`
}

type Response struct {
	RequestID uint64 `json:"request_id"`

	Data interface{} `json:"data"`

	Error string `json:"error,omitempty"`
}

type DataPoint struct {
	Timestamp uint32 `json:"timestamp"`

	Value interface{} `json:"value"`
}
