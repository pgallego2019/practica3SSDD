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
