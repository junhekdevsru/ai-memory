package memory

import "errors"

// Попытка создать регион с лимитом 1 или меньше
// Лимит нейронов должен быть хотя бы 1
var ErrInvalidMaxNeurons = errors.New("memory: max neurons must be at least 1")

var ErrNeuronNotFound = errors.New("memory: neuron not found")

var ErrDuplicateID = errors.New("memory: neuron with this ID already exists")

var ErrSelfLink = errors.New("memory: cannot link neuron to itself")
