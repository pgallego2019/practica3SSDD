package sim

import (
	"sync"
	"taller/models"
	"time"
)

type Simulador struct {
	Taller *models.Taller
	WG     sync.WaitGroup
	Start  time.Time
}

// Constructor del simulador
func NewSimulador(t *models.Taller) *Simulador {
	return &Simulador{
		Taller: t,
		Start:  time.Now(),
	}
}
