package cursorslice

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type CursorSlice struct {
	mutex   sync.RWMutex
	items   []interface{}
	readers *readerEntry
}

type readerEntry struct {
	next, prev *readerEntry
	position   uint64
}

func NewCursorSlice() *CursorSlice {
	return &CursorSlice{}
}

func (cs *CursorSlice) Append(items ...interface{}) {
	_ = sync.Map{}
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.items = append(cs.items, items...)
}

func (cs *CursorSlice) Range(f func(key int, value interface{}) bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	var re readerEntry
	for {
		var reader = (*unsafe.Pointer)(unsafe.Pointer(&cs.readers))
		old := atomic.LoadPointer(reader)
		re = readerEntry{
			next: (*readerEntry)(old),
			prev: nil,
		}
		_, _ = reader, re
		swapped := atomic.CompareAndSwapPointer(reader, old, unsafe.Pointer(&re))
		if swapped {
			break
		}
	}

	mustLoad := true
	var nextPos uint64
	var exit bool
	for i := range cs.items {
		hasNext := re.next != nil
		if hasNext {
			for {
				if mustLoad {
					nextPos = (atomic.LoadUint64(&re.next.position))
				}
				if nextPos == uint64(len(cs.items)) {
					re.next = nil
				}
				if nextPos > uint64(i) {
					mustLoad = false
					break
				}
				mustLoad = true
				runtime.Gosched()
			}
		}

		if !exit {
			exit = !f(i, cs.items[i])
		}

		atomic.StoreUint64(&re.position, uint64(i))
	}
	atomic.StoreUint64(&re.position, uint64(len(cs.items)))
	{
		var reader = (*unsafe.Pointer)(unsafe.Pointer(&cs.readers))
		old := atomic.LoadPointer(reader)

		_ = atomic.CompareAndSwapPointer(reader, old, unsafe.Pointer(nil))
	}
}
