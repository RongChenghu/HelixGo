package telegram

const defaultDedupWindow = 2000

// Deduper 用于维护最近 N 个 update_id，防止重复执行 handler。
// 结构：一个定长 ring + 一个 map 做 O(1) 查重。
type Deduper struct {
	cap  int     // 窗口大小
	ring []int64 // 环形数组，存最近的 id
	idx  int     // 当前写入位置
	seen map[int64]struct{}
}

// NewDeduper 创建一个指定容量的去重器；容量 <=0 时使用默认值。
func NewDeduper(window int) *Deduper {
	if window <= 0 {
		window = defaultDedupWindow
	}
	return &Deduper{
		cap:  window,
		ring: make([]int64, 0, window),
		seen: make(map[int64]struct{}, window),
	}
}

// Seen 在 update_id 进入处理前调用：
//   - 如果之前见过：返回 true（调用方应跳过本次处理）
//   - 如果没见过：记录进窗口并返回 false
func (d *Deduper) Seen(id int64) bool {
	if d.cap <= 0 {
		return false
	}
	if _, ok := d.seen[id]; ok {
		return true
	}

	// 先把新 id 放入 ring / seen，再处理淘汰逻辑。
	if len(d.ring) < d.cap {
		d.ring = append(d.ring, id)
	} else {
		// 环形写入：淘汰最老的一个 id。
		evict := d.ring[d.idx]
		delete(d.seen, evict)
		d.ring[d.idx] = id
	}
	d.seen[id] = struct{}{}

	d.idx++
	if d.idx >= d.cap {
		d.idx = 0
	}
	return false
}
