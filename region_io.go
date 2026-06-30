package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func (r *Region) Save(path string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	regionData := struct {
		MaxNeurons int      `json:"maxNeurons"`
		Active     []Neuron `json:"active"`
		Crumbs     []Neuron `json:"crumbs"`
		Edges      []Edge   `json:"edges"`
	}{
		MaxNeurons: r.maxNeurons,
		Active:     r.active,
		Crumbs:     r.crumbs,
		Edges:      r.edges,
	}

	data, err := json.Marshal(regionData)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".memory-save-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	cleanup := true
	defer func() {
		if cleanup {
			os.Remove(tmp)
		}
	}()

	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func LoadRegion(path string) (*Region, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var regionData struct {
		MaxNeurons int      `json:"maxNeurons"`
		Active     []Neuron `json:"active"`
		Crumbs     []Neuron `json:"crumbs"`
		Edges      []Edge   `json:"edges"`
	}

	if err = json.Unmarshal(data, &regionData); err != nil {
		return nil, err
	}

	if regionData.MaxNeurons < 1 {
		return nil, ErrInvalidMaxNeurons
	}

	return &Region{
		maxNeurons: regionData.MaxNeurons,
		active:     regionData.Active,
		crumbs:     regionData.Crumbs,
		edges:      regionData.Edges,
	}, nil
}
