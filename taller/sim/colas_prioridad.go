package sim

import (
	"sync"
	"taller/models"
)

type ColaPrioritaria struct {
	altas  []*models.Vehiculo
	medias []*models.Vehiculo
	bajas  []*models.Vehiculo
	mtx    sync.RWMutex
	notify chan struct{}
}

func NewColaPrioritaria() *ColaPrioritaria {
	return &ColaPrioritaria{
		notify: make(chan struct{}, 1),
	}
}

func (c *ColaPrioritaria) Push(v *models.Vehiculo) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	switch v.Incidencia.Tipo {
	case models.Mecanica:
		c.altas = append(c.altas, v)
	case models.Electrica:
		c.medias = append(c.medias, v)
	case models.Carroceria:
		c.bajas = append(c.bajas, v)
	}

	// Notificar a los workers que hay un vehículo disponible
	select {
	case c.notify <- struct{}{}:
	default:
		// si el canal ya tiene notificación, no bloquear
	}
}

func (c *ColaPrioritaria) PopFront() *models.Vehiculo {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	var v *models.Vehiculo

	if len(c.altas) > 0 {
		v = c.altas[0]
		c.altas = c.altas[1:]
	} else if len(c.medias) > 0 {
		v = c.medias[0]
		c.medias = c.medias[1:]
	} else if len(c.bajas) > 0 {
		v = c.bajas[0]
		c.bajas = c.bajas[1:]
	}

	// Si aún hay vehículos, notificar otro worker
	if len(c.altas)+len(c.medias)+len(c.bajas) > 0 {
		select {
		case c.notify <- struct{}{}:
		default:
		}
	}

	return v
}

func (c *ColaPrioritaria) Len() int {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return len(c.altas) + len(c.medias) + len(c.bajas)
}
