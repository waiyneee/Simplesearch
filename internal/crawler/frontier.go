package crawler

type FrontierItem struct {
	URL   string
	Depth int
}

type Frontier struct {
	queue   []FrontierItem
	visited map[string]struct{}
}

func NewFrontier() *Frontier {
	return &Frontier{
		queue:   make([]FrontierItem, 0, 128),
		visited: make(map[string]struct{}, 256),
	}
}

// Add enqueues URL if unseen. Marks seen immediately.
func (f *Frontier) Add(url string, depth int) bool {
	if _, ok := f.visited[url]; ok {
		return false
	}
	f.visited[url] = struct{}{}
	f.queue = append(f.queue, FrontierItem{
		URL:   url,
		Depth: depth,
	})
	return true
}

// Next pops FIFO.
func (f *Frontier) Next() (FrontierItem, bool) {
	if len(f.queue) == 0 {
		return FrontierItem{}, false
	}
	item := f.queue[0]
	f.queue[0] = FrontierItem{}
	f.queue = f.queue[1:]
	return item, true
}

func (f *Frontier) Len() int {
	return len(f.queue)
}

func (f *Frontier) Seen(url string) bool {
	_, ok := f.visited[url]
	return ok
}