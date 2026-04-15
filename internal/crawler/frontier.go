package crawler 

// import "sync"  
//no need of mu

type FrontierItem struct{
	URL string
	Depth int
}

type Frontier struct{
	queue []FrontierItem
	visited map[string]struct{}
}


// Add fn is here ::enqueues url if it has not been seen before.
// Returns true if added, false if it was already seen.
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

// Next dequeues one item in FIFO order.
// ok=false means frontier is empty.
func (f *Frontier) Next() (item FrontierItem, ok bool) {
	if len(f.queue) == 0 {
		return FrontierItem{}, false
	}

	item = f.queue[0]
	f.queue[0] = FrontierItem{} // help GC
	f.queue = f.queue[1:]
	return item, true
}

 