package gateway

import "sync/atomic"

var reqIdFactor uint64

func generateId() uint64 {
	return atomic.AddUint64(&reqIdFactor, 1)
}
