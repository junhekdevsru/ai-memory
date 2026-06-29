# ai-memory

A Go package for storing structured memory entries (neurons) with an active memory cap, links between entries, and JSON persistence.

## Concept

Memory is split into two layers:

- **Active** — full entries living in memory. Count is capped at `maxNeurons`.
- **Crumbs** — evicted entries with the `Body` field cleared. Metadata is preserved, heavy content is not.

When the active layer fills up, the oldest neuron is automatically pushed down to crumbs.

## Installation

```bash
go get github.com/junhekdevsru/ai-memory
```

## Quick start

```go
import "github.com/junhekdevsru/ai-memory"

// Create a region with a cap of 100 active neurons
r, err := memory.NewRegion(100)

// Add a neuron
err = r.Add(memory.Neuron{
    ID:          "task-42",
    Title:       "Auth refactor",
    Theme:       "backend",
    TaskName:    "AUTH-42",
    Description: "Moved middleware into a separate package",
    Body:        "Full task context...",
})

// Look up by title and theme
neuron, loc := r.Lookup("Auth refactor", "backend")
// loc == memory.Active or memory.Crumbs or memory.NotFound

// Link two neurons
err = r.Link("task-42", "task-99")

// Evict manually
err = r.Evict("task-42")

// Save to disk / load from disk
err = r.Save("memory.json")
r, err = memory.LoadRegion("memory.json")
```

## API

### Types

```go
type Neuron struct {
    ID          string
    Title       string
    Theme       string
    TaskName    string
    Description string
    Body        string
    CreatedAt   time.Time
    LastSeen    time.Time
}

type Edge struct {
    A, B string // neuron IDs
}

type Location int // NotFound | Active | Crumbs
```

### Region

| Method | Description |
|---|---|
| `NewRegion(maxNeurons int) (*Region, error)` | Create a region. `maxNeurons` must be ≥ 1 |
| `Add(n Neuron) error` | Add a neuron. Evicts the first one when full |
| `Lookup(title, theme string) (*Neuron, Location)` | Find a neuron by title and theme |
| `Link(a, b string) error` | Create a link between two neurons |
| `Evict(id string) error` | Manually push a neuron to crumbs |
| `Active() []Neuron` | Snapshot of the active layer |
| `Crumbs() []Neuron` | Snapshot of crumbs |
| `Save(path string) error` | Persist to a JSON file |
| `LoadRegion(path string) (*Region, error)` | Load from a JSON file |

### Errors

| Error | When |
|---|---|
| `ErrInvalidMaxNeurons` | `maxNeurons` < 1 |
| `ErrNeuronNotFound` | No neuron with the given ID |
| `ErrDuplicateID` | A neuron with this ID already exists |
| `ErrSelfLink` | Attempt to link a neuron to itself |

## Package layout

| File | Contents |
|---|---|
| `neuron.go` | Data types: `Neuron`, `Edge`, `Location` |
| `region.go` | `Region` entity and all logic |
| `region_io.go` | `Save` / `LoadRegion` — persistence |
| `errors.go` | Sentinel errors |

## Thread safety

All `Region` methods are guarded by a `sync.RWMutex`. Safe to use from multiple goroutines.
