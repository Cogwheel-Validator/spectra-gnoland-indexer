package maintypes

import "time"

type RpcFlags struct {
	RequestsPerWindow int
	Timeout           time.Duration
	TimeWindow        time.Duration
}

type RunningFlags struct {
	RunningMode        string
	SkipInitialDbCheck bool
	CompressEvents     bool
	FromHeight         uint64
	ToHeight           uint64
}
