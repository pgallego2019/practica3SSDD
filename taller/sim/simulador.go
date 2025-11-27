package sim

import (
	"sync"
	"taller/models"
	"time"
)

type Simulador struct {
	Taller      *models.Taller
	MtxPlazas   sync.Mutex
	MtxMecanico sync.Mutex
	MtxLimpieza sync.Mutex
	MtxRevision sync.Mutex
	WG          sync.WaitGroup
	Start       time.Time
}

// Constructor del simulador
func NewSimulador(t *models.Taller) *Simulador {
	return &Simulador{
		Taller: t,
		Start:  time.Now(),
	}
}
