package telegram

import (
	"sync"
)

const defaultDedupWindow = 2000

// dedupSet keeps a bounded set of update_id to skip duplicates (Telegram may re-send boundary updates).
type dedupSet struct {
	mu   sync.Mutex
	m    map[int64]struct{}
	max  int
}

func newDedupSet(window int) *dedupSet {
	if window <= 0 {
		window = defaultDedupWindow
	}
	return &dedupSet{m: make(map[int64]struct{}), max: window}
}

// Seen returns true if id was already added (duplicate).
func (d *dedupSet) Seen(id int64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.m[id]
	return ok
}

// Add adds id and evicts oldest (smallest update_id) if over capacity.
func (d *dedupSet) Add(id int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.m[id]; ok {
		return
	}
	for len(d.m) >= d.max {
		var minID int64
		first := true
		for k := range d.m {
			if first || k < minID {
				minID = k
				first = false
			}
		}
		delete(d.m, minID)
	}
	d.m[id] = struct{}{}
}
