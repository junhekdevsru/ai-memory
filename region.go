package memory

import (
	"sync"
	"time"
)

type Region struct {
	mu         sync.RWMutex
	maxNeurons int
	active     []Neuron
	crumbs     []Neuron
	edges      []Edge
}

func NewRegion(maxNeurons int) (*Region, error) {
	if maxNeurons < 1 {
		return nil, ErrInvalidMaxNeurons
	}

	return &Region{
		maxNeurons: maxNeurons,
	}, nil
}

func (r *Region) findActive(id string) int {
	for i, n := range r.active {
		if n.ID == id {
			return i
		}
	}
	return -1
}

func (r *Region) findCrumbs(id string) int {
	for i, n := range r.crumbs {
		if n.ID == id {
			return i
		}
	}
	return -1
}

func (r *Region) findAny(id string) (int, Location) {
	if i := r.findActive(id); i >= 0 {
		return i, Active
	}
	if i := r.findCrumbs(id); i >= 0 {
		return i, Crumbs
	}
	return -1, NotFound
}

func (r *Region) Add(n Neuron) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, loc := r.findAny(n.ID); loc != NotFound {
		return ErrDuplicateID
	}

	now := time.Now()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	if n.LastSeen.IsZero() {
		n.LastSeen = now
	}

	if len(r.active) >= r.maxNeurons {
		r.evictToCrumbs(r.lruIndex())
	}

	r.active = append(r.active, n)

	return nil
}

func (r *Region) evictToCrumbs(index int) {
	n := r.active[index]

	n.Body = ""

	r.crumbs = append(r.crumbs, n)

	r.active = append(r.active[:index], r.active[index+1:]...)
}

func (r *Region) lruIndex() int {
	oldest := 0
	for i := 1; i < len(r.active); i++ {
		if r.active[i].LastSeen.Before(r.active[oldest].LastSeen) {
			oldest = i
		}
	}
	return oldest
}

func (r *Region) Lookup(title, theme string) (*Neuron, Location) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.active {
		if r.active[i].Title == title && r.active[i].Theme == theme {
			r.active[i].LastSeen = time.Now()
			result := r.active[i]
			return &result, Active
		}
	}

	for i, n := range r.crumbs {
		if n.Title == title && n.Theme == theme {
			result := r.crumbs[i]
			return &result, Crumbs
		}
	}

	return nil, NotFound
}

func (r *Region) GetByID(id string) (*Neuron, Location) {
	r.mu.Lock()
	defer r.mu.Unlock()

	i, loc := r.findAny(id)
	if loc == NotFound {
		return nil, NotFound
	}

	if loc == Active {
		r.active[i].LastSeen = time.Now()
		result := r.active[i]
		return &result, Active
	}

	result := r.crumbs[i]
	return &result, Crumbs
}

func (r *Region) Link(a, b string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if a == b {
		return ErrSelfLink
	}

	_, locA := r.findAny(a)
	if locA == NotFound {
		return ErrNeuronNotFound
	}
	_, locB := r.findAny(b)
	if locB == NotFound {
		return ErrNeuronNotFound
	}

	for _, e := range r.edges {
		if (e.A == a && e.B == b) || (e.A == b && e.B == a) {
			return nil
		}
	}

	r.edges = append(r.edges, Edge{A: a, B: b})

	return nil
}

func (r *Region) Evict(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.findActive(id)
	if i < 0 {
		return ErrNeuronNotFound
	}

	r.evictToCrumbs(i)
	return nil
}

func (r *Region) Active() []Neuron {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Neuron, len(r.active))
	copy(result, r.active)

	return result
}

func (r *Region) Crumbs() []Neuron {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Neuron, len(r.crumbs))
	copy(result, r.crumbs)

	return result
}

func (r *Region) Edges() []Edge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Edge, len(r.edges))
	copy(result, r.edges)

	return result
}
