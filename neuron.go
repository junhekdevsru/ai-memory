package memory

import "time"

type Neuron struct {
	ID string							// ID
	Title string				  // Короткое название 
	Theme string          // Категория нейрона 			  
	TaskName string       // Имя задачи
	Description string    // Одно-два предложения что сделали
	Body string           // Фулл контекст
	CreatedAt time.Time 
	LastSeen time.Time
}

type Edge struct {
	A string
	B string
}

type Location int

const (
	NotFound Location = iota // 0 - не нашли нигде
	Active                   // 1 - нашли среди живых нейронов
	Crumbs 									 // 2 - нашли среди крошек
)
