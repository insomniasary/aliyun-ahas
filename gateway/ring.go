package gateway

import "sync/atomic"

type ring struct {
	values []interface{}
	index  uint32
}

func newRing() *ring {
	return &ring{
		values: make([]interface{}, 0),
	}
}

func (r *ring) add(value interface{}) {
	if value == nil {
		return
	}
	r.values = append(r.values, value)
}

func (r *ring) next() interface{} {
	atomic.AddUint32(&r.index, 1)
	if atomic.LoadUint32(&r.index) > 10000 {
		atomic.StoreUint32(&r.index, 0)
	}

	index := atomic.LoadUint32(&r.index) % uint32(len(r.values))
	return r.values[index]
}
