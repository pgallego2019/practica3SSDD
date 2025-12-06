package sim

import (
	"sync"
	"taller/models"
	"time"
)

// NO tener WaitGroups dentro de structs que tienen un ciclo de vida mayor que la ejecución puntual del trabajo
type Simulador struct {
	Taller *models.Taller
	Start  time.Time
	Done   chan struct{} // Para saber cuando cerrar
}

// Constructor del simulador
func NewSimulador(t *models.Taller) *Simulador {
	return &Simulador{
		Taller: t,
		Start:  time.Now(),
	}
}

// Uso un worker por fase para que runsim solo tenga que manejar el flujo

// TODO hacer la impresion cuando se cambia de fase desde la cola prioritaria?
// cuando se entra/sale se llama a imprimirVehiculo desde el worker
// Así no haría sleep en los workers cuando la cola está vacía, sino que esperaría a que haya vehículos

func (s *Simulador) workerEntrada(
	colaIn *ColaPrioritaria,
	colaOut *ColaPrioritaria,
	semPlazas chan struct{},
) {
	for {
		select {
		case <-s.Done:
			return // fin del worker
		case <-colaIn.notify:
			// hay vehículo, seguimos más abajo
		}

		v := colaIn.PopFront()
		if v == nil {
			continue
		}

		<-semPlazas // ocupa plaza

		s.imprimirVehiculo(v, FaseEntrada, "Entra plaza")
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		s.imprimirVehiculo(v, FaseEntrada, "Sale plaza")

		semPlazas <- struct{}{} // libera plaza

		colaOut.Push(v) // pasa a mecánico
	}
}

func (s *Simulador) workerMecanico(
	colaIn *ColaPrioritaria,
	colaOut *ColaPrioritaria,
	semMec chan struct{},
) {
	for {
		select {
		case <-s.Done:
			return
		case <-colaIn.notify:
		}

		v := colaIn.PopFront()
		if v == nil {
			continue
		}

		<-semMec // ocupa mecánico
		s.imprimirVehiculo(v, FaseAtencion, "Atendido por mecánico")
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		s.imprimirVehiculo(v, FaseAtencion, "Finaliza mecánico")
		semMec <- struct{}{} // libera mecánico

		colaOut.Push(v) // pasa a limpieza
	}
}

func (s *Simulador) workerLimpieza(
	colaIn *ColaPrioritaria,
	colaOut *ColaPrioritaria,
	semLimp chan struct{},
) {
	for {
		select {
		case <-s.Done:
			return
		case <-colaIn.notify:
		}

		v := colaIn.PopFront()
		if v == nil {
			continue
		}

		<-semLimp
		s.imprimirVehiculo(v, FaseLimpieza, "Limpiando")
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		s.imprimirVehiculo(v, FaseLimpieza, "Limpieza finalizada")
		semLimp <- struct{}{}

		colaOut.Push(v)
	}
}

func (s *Simulador) workerRevision(
	colaIn *ColaPrioritaria,
	semRev chan struct{},
	wg *sync.WaitGroup,
) {
	for {
		select {
		case <-s.Done:
			return
		case <-colaIn.notify:
		}

		v := colaIn.PopFront()
		if v == nil {
			continue
		}

		<-semRev
		s.imprimirVehiculo(v, FaseRevision, "Revisión")
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		s.imprimirVehiculo(v, FaseRevision, "Vehículo entregado")
		semRev <- struct{}{}

		wg.Done() // vehículo terminado
	}
}
