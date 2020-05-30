package cache

import "sync"

type MemoryPointerCache struct {
	cache *sync.Map
}

func NewMemoryPointerCache() *MemoryPointerCache {
	return &MemoryPointerCache{cache: &sync.Map{}}
}

func (mc *MemoryPointerCache) Delete(v interface{}) {
	// get memory address first
	u := GetMemoryPointer(v)
	if u != 0 {
		mc.cache.Delete(u)
	}
}

func (mc *MemoryPointerCache) GetOrCompute(v interface{}, compute func(v interface{}) interface{}) interface{} {
	// get memory address first
	u := GetMemoryPointer(v)
	if u != 0 {
		// check cache
		if val, ok := mc.cache.Load(u); ok {
			return val
		} else {
			// compute new value if needed
			val := compute(v)
			mc.cache.Store(u, val)
			return val
		}
	}

	return nil
}
