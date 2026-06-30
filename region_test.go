package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewRegion(t *testing.T) {
	_, err := NewRegion(0)
	if err != ErrInvalidMaxNeurons {
		t.Fatalf("expected ErrInvalidMaxNeurons, got %v", err)
	}

	r, err := NewRegion(1)
	if err != nil || r == nil {
		t.Fatalf("expected valid region, got %v", err)
	}
}

func TestAdd(t *testing.T) {
	r, _ := NewRegion(3)

	n := Neuron{ID: "1", Title: "foo", Theme: "bar"}
	if err := r.Add(n); err != nil {
		t.Fatal(err)
	}

	// дубликат
	if err := r.Add(n); err != ErrDuplicateID {
		t.Fatalf("expected ErrDuplicateID, got %v", err)
	}
}

func TestEviction(t *testing.T) {
	r, _ := NewRegion(2)

	r.Add(Neuron{ID: "1", Body: "body1"})
	r.Add(Neuron{ID: "2", Body: "body2"})
	// третий вытесняет первый в крошки
	r.Add(Neuron{ID: "3", Body: "body3"})

	active := r.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 active, got %d", len(active))
	}

	crumbs := r.Crumbs()
	if len(crumbs) != 1 {
		t.Fatalf("expected 1 crumb, got %d", len(crumbs))
	}
	if crumbs[0].ID != "1" {
		t.Fatalf("expected crumb ID=1, got %s", crumbs[0].ID)
	}
	if crumbs[0].Body != "" {
		t.Fatal("crumb body must be cleared")
	}
}

func TestLookup(t *testing.T) {
	r, _ := NewRegion(10)
	r.Add(Neuron{ID: "1", Title: "auth", Theme: "backend"})

	n, loc := r.Lookup("auth", "backend")
	if loc != Active || n == nil {
		t.Fatalf("expected Active, got %v", loc)
	}

	_, loc = r.Lookup("nope", "nope")
	if loc != NotFound {
		t.Fatalf("expected NotFound, got %v", loc)
	}
}

func TestLookupCrumbs(t *testing.T) {
	r, _ := NewRegion(1)
	r.Add(Neuron{ID: "1", Title: "old", Theme: "t"})
	r.Add(Neuron{ID: "2", Title: "new", Theme: "t"}) // вытесняет "1"

	_, loc := r.Lookup("old", "t")
	if loc != Crumbs {
		t.Fatalf("expected Crumbs, got %v", loc)
	}
}

func TestEvict(t *testing.T) {
	r, _ := NewRegion(10)
	r.Add(Neuron{ID: "1"})

	if err := r.Evict("1"); err != nil {
		t.Fatal(err)
	}
	if err := r.Evict("1"); err != ErrNeuronNotFound {
		t.Fatalf("expected ErrNeuronNotFound, got %v", err)
	}
}

func TestLink(t *testing.T) {
	r, _ := NewRegion(10)
	r.Add(Neuron{ID: "a"})
	r.Add(Neuron{ID: "b"})

	if err := r.Link("a", "b"); err != nil {
		t.Fatal(err)
	}
	// дубликат ребра — не ошибка
	if err := r.Link("b", "a"); err != nil {
		t.Fatal(err)
	}

	if err := r.Link("a", "a"); err != ErrSelfLink {
		t.Fatalf("expected ErrSelfLink, got %v", err)
	}

	if err := r.Link("a", "ghost"); err != ErrNeuronNotFound {
		t.Fatalf("expected ErrNeuronNotFound, got %v", err)
	}
}

func TestSaveLoad(t *testing.T) {
	r, _ := NewRegion(10)
	r.Add(Neuron{ID: "1", Title: "x", Theme: "y", Body: "content"})
	r.Add(Neuron{ID: "2", Title: "a", Theme: "b"})
	r.Link("1", "2")

	path := t.TempDir() + "/mem.json"
	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}

	r2, err := LoadRegion(path)
	if err != nil {
		t.Fatal(err)
	}

	active := r2.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 neurons, got %d", len(active))
	}

	n, loc := r2.Lookup("x", "y")
	if loc != Active || n.Body != "content" {
		t.Fatalf("loaded neuron mismatch: loc=%v body=%q", loc, n.Body)
	}

	os.Remove(path)
}

func TestLRU(t *testing.T) {
	r, _ := NewRegion(2)
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	r.Add(Neuron{ID: "a", Title: "ta", Theme: "th", LastSeen: base})
	r.Add(Neuron{ID: "b", Title: "tb", Theme: "th", LastSeen: base})

	if _, loc := r.Lookup("ta", "th"); loc != Active {
		t.Fatalf("expected Active, got %v", loc)
	}

	r.Add(Neuron{ID: "c", Title: "tc", Theme: "th", LastSeen: base.Add(time.Hour)})

	active := r.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 active, got %d", len(active))
	}
	ids := map[string]bool{active[0].ID: true, active[1].ID: true}
	if !ids["a"] || !ids["c"] {
		t.Fatalf("expected active [a,c], got %v", active)
	}

	crumbs := r.Crumbs()
	if len(crumbs) != 1 || crumbs[0].ID != "b" {
		t.Fatalf("expected crumb b (LRU), got %v", crumbs)
	}
}

func TestLookupUpdatesLastSeen(t *testing.T) {
	r, _ := NewRegion(10)
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Add(Neuron{ID: "1", Title: "x", Theme: "y", LastSeen: old})

	n, _ := r.Lookup("x", "y")
	if !n.LastSeen.After(old) {
		t.Fatalf("expected LastSeen to be refreshed, got %v", n.LastSeen)
	}
}

func TestGetByID(t *testing.T) {
	r, _ := NewRegion(10)
	r.Add(Neuron{ID: "1", Title: "x"})

	n, loc := r.GetByID("1")
	if loc != Active || n == nil || n.Title != "x" {
		t.Fatalf("unexpected: n=%v loc=%v", n, loc)
	}

	if _, loc := r.GetByID("ghost"); loc != NotFound {
		t.Fatalf("expected NotFound, got %v", loc)
	}
}

func TestEdges(t *testing.T) {
	r, _ := NewRegion(10)
	r.Add(Neuron{ID: "a"})
	r.Add(Neuron{ID: "b"})
	r.Link("a", "b")

	edges := r.Edges()
	if len(edges) != 1 || edges[0].A != "a" || edges[0].B != "b" {
		t.Fatalf("unexpected edges: %v", edges)
	}

	edges[0].A = "mutated"
	again := r.Edges()
	if again[0].A != "a" {
		t.Fatal("Edges() must return a snapshot")
	}
}

func TestLoadRegionInvalidMaxNeurons(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(path, []byte(`{"maxNeurons":0}`), 0644)

	if _, err := LoadRegion(path); err != ErrInvalidMaxNeurons {
		t.Fatalf("expected ErrInvalidMaxNeurons, got %v", err)
	}
}

func TestSaveAtomicNoTempLeft(t *testing.T) {
	r, _ := NewRegion(2)
	r.Add(Neuron{ID: "1"})

	dir := t.TempDir()
	path := filepath.Join(dir, "mem.json")
	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected only target file to remain, got %d entries", len(entries))
	}
}
