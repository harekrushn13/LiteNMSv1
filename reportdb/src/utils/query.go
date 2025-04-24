package utils

type QueryReceive struct {
	RequestID uint64 `json:"request_id"`

	Query Query `json:"query_request"`
}

type Query struct {
	CounterID uint16 `json:"counter_id"`

	ObjectIDs []uint32 `json:"object_ids"`

	From uint32 `json:"from"`

	To uint32 `json:"to"`

	Aggregation string `json:"aggregation"`

	GroupByObjects bool `json:"group_by_objects"`

	Interval int `json:"interval"`
}

type Response struct {
	RequestID uint64 `json:"request_id"`

	Error string `json:"error,omitempty"`

	Data interface{} `json:"data"`
}
