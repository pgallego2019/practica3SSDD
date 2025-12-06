package sim

import (
	"sync"
	"time"
)

type RecursosSim struct {
	ColaEntrada, ColaMecanico, ColaLimpieza, ColaRevision *ColaPrioritaria
	SemPlazas, SemLimp, SemRev, SemMec                    chan struct{}
	NumPlazas, NumMecanicos                               int
}

func inicializarRecursos(numPlazas, numMecanicos int) *RecursosSim {
	rs := &RecursosSim{
		ColaEntrada:  NewColaPrioritaria(),
		ColaMecanico: NewColaPrioritaria(),
		ColaLimpieza: NewColaPrioritaria(),
		ColaRevision: NewColaPrioritaria(),
		SemPlazas:    make(chan struct{}, numPlazas),
		SemLimp:      make(chan struct{}, numPlazas),
		SemRev:       make(chan struct{}, numPlazas),
		SemMec:       make(chan struct{}, numMecanicos),
		NumPlazas:    numPlazas,
		NumMecanicos: numMecanicos,
	}

	for i := 0; i < numPlazas; i++ {
		rs.SemPlazas <- struct{}{}
		rs.SemLimp <- struct{}{}
		rs.SemRev <- struct{}{}
	}
	for i := 0; i < numMecanicos; i++ {
		rs.SemMec <- struct{}{}
	}
	return rs
}

func LanzarWorker(colaIn, colaOut *ColaPrioritaria, sem chan struct{}, fase Fase, metricas *Metricas, wg *sync.WaitGroup) {
	go func() {
		for {
			<-colaIn.notify
			v := colaIn.PopFront()
			if v == nil {
				continue
			}
			<-sem
			start := time.Now()
			time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
			sem <- struct{}{}
			metricas.RegistrarVehiculo(v, fase, time.Since(start))
			if colaOut != nil {
				colaOut.Push(v)
			} else if wg != nil {
				wg.Done()
			}
		}
	}()
}
