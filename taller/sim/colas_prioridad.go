package sim

import (
	"sync"
	"taller/models"
)

// Para mantener un orden de prioridad en cada fase uso una cola con 3 listas internas
// Pero necesita un rwmutex para ser seguro en concurrencia

//TODO añadir un canal de mensajes cuando se cambia de fase?
/*Push() envía notify
worker hace <-notify
worker hace PopFront()
worker SI llama a imprimirVehiculo (no runsim)*/

type ColaPrioritaria struct {
	altas  []*models.Vehiculo
	medias []*models.Vehiculo
	bajas  []*models.Vehiculo
	mtx    sync.RWMutex
	notify chan struct{} // canal para notificar a workers que hay un vehículo
}

func NewColaPrioritaria() *ColaPrioritaria {
	return &ColaPrioritaria{
		notify: make(chan struct{}, 1),
		/*Cada worker solo despierta cuando recibe un notify. (buffer 1)
		si push mete 5 vehículos cuando la cola está vacía, SOLO envía 1 notify.
		Los otros 4 vehículos nunca despieren a los workers -> atascos

		Por eso, cuando un worker hace PopFront, si aún hay vehículos en la cola,
		vuelve a enviar otro notify y despierta a otro worker.
		Push solo despierta cuando llega el primer vehículo a una cola vacía.

		NO tengo espera activa (sleep) en los workers, solo esperan en el canal
		NO hay condición de carrera porque el canal es seguro en concurrencia (mtx)
		Se imprimen mensajes desde los workers cuando entran/salen de fase
		*/
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
