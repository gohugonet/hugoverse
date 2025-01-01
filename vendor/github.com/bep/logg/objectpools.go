package logg

import "sync"

var objectPools = &objectPoolsHolder{
	entryPool: &sync.Pool{
		New: func() any {
			return &Entry{}
		},
	},
}

type objectPoolsHolder struct {
	// This is only used for the event copy passed to HandleLog.
	entryPool *sync.Pool
}

func (h *objectPoolsHolder) GetEntry() *Entry {
	return h.entryPool.Get().(*Entry)
}

func (h *objectPoolsHolder) PutEntry(e *Entry) {
	e.reset()
	h.entryPool.Put(e)
}
