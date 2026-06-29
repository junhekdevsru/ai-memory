package memory

import (
	"encoding/json"
	"os"
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

	return os.WriteFile(path, data, 0644)
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

	return &Region{
		maxNeurons: regionData.MaxNeurons,
		active:     regionData.Active,
		crumbs:     regionData.Crumbs,
		edges:      regionData.Edges,
	}, nil
}
