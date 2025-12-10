package sim

import (
	"fmt"
	"sync"
	"taller/models"
	"time"
)

type SimuladorWaitGroup struct {
	Taller  *models.Taller
	Start   time.Time
	Done    chan struct{}
	Verbose bool
}

func (s *SimuladorWaitGroup) SetVerbose(v bool) {
	s.Verbose = v
}

func NewSimuladorWaitGroup(t *models.Taller) *SimuladorWaitGroup {
	return &SimuladorWaitGroup{
		Taller: t,
		Start:  time.Now(),
	}
}

func (s *SimuladorWaitGroup) workerEntrada(
	colaIn *ColaPrioritaria,
	colaOut *ColaPrioritaria,
	semPlazas chan struct{},
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
) {
	for {
		select {
		case <-s.Done:
			return
		case <-colaIn.notify:
			// hay vehículo, seguimos
		}

		v := colaIn.PopFront()
		if v == nil {
			continue
		}

		<-semPlazas

		s.imprimirVehiculo(v, FaseEntrada, "Entra plaza")
		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		metricas.RegistrarVehiculo(v, FaseEntrada, time.Since(start))
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, time.Since(start))
		s.imprimirVehiculo(v, FaseEntrada, "Sale plaza")
		aux[FaseEntrada].Registrar(time.Since(start))

		semPlazas <- struct{}{}

		colaOut.Push(v)
	}
}

func (s *SimuladorWaitGroup) workerMecanico(
	colaIn *ColaPrioritaria,
	colaOut *ColaPrioritaria,
	semMec chan struct{},
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
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

		<-semMec
		s.imprimirVehiculo(v, FaseAtencion, "Atendido por mecánico")
		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		metricas.RegistrarVehiculo(v, FaseAtencion, time.Since(start))
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, time.Since(start))
		s.imprimirVehiculo(v, FaseAtencion, "Finaliza mecánico")
		aux[FaseAtencion].Registrar(time.Since(start))

		semMec <- struct{}{}

		colaOut.Push(v)
	}
}

func (s *SimuladorWaitGroup) workerLimpieza(
	colaIn *ColaPrioritaria,
	colaOut *ColaPrioritaria,
	semLimp chan struct{},
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
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
		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		metricas.RegistrarVehiculo(v, FaseLimpieza, time.Since(start))
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, time.Since(start))
		s.imprimirVehiculo(v, FaseLimpieza, "Limpieza finalizada")
		aux[FaseLimpieza].Registrar(time.Since(start))

		semLimp <- struct{}{}

		colaOut.Push(v)
	}
}

func (s *SimuladorWaitGroup) workerRevision(
	colaIn *ColaPrioritaria,
	semRev chan struct{},
	wg *sync.WaitGroup,
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
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
		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		metricas.RegistrarVehiculo(v, FaseRevision, time.Since(start))
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, time.Since(start))
		s.imprimirVehiculo(v, FaseRevision, "Vehículo entregado")
		aux[FaseRevision].Registrar(time.Since(start))

		semRev <- struct{}{}

		wg.Done() // vehículo terminado
	}
}

func (s *SimuladorWaitGroup) imprimirVehiculo(v *models.Vehiculo, fase Fase, estado string) {
	if !s.Verbose {
		return
	}
	elapsed := time.Since(s.Start).Truncate(time.Millisecond)
	fmt.Printf("Tiempo %v | Vehiculo %s | Incidencia %s | Fase %s | Estado %s\n",
		elapsed, v.Matricula, v.Incidencia.Tipo, fase.String(), estado)
}

func (s *SimuladorWaitGroup) RunSim(
	vehiculos []*models.Vehiculo,
	Sims int,
	Nvehiculos int,
	NumPlazas int,
	NumMecanicos int,
	maxEsperas map[Fase]int,
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,

) {
	for sim := 1; sim <= Sims; sim++ {
		fmt.Printf("\n=== SIMULACIÓN WaitGroup %d (%d plazas y %d mecánicos) ===\n", sim, NumPlazas, NumMecanicos)

		// Defensive init: si llamaron con nil, inicializamos aquí para evitar panic desde workers
		if metricas == nil {
			metricas = NuevaMetricas()
		}
		if tiempoPorVehiculo == nil {
			tiempoPorVehiculo = NuevaTiempoVehiculo()
		}
		if aux == nil {
			aux = inicializarMetricasAux()
		}

		metricas.Simulador = "WaitGroup"

		s.Start = time.Now()
		s.Done = make(chan struct{})
		ImprimirResumenCategorias(vehiculos)
		metricas.Inicio = time.Now()
		var wgFinal sync.WaitGroup
		wgFinal.Add(len(vehiculos))

		// Colas por prioridad para cada fase
		colaEntrada := NewColaPrioritaria()
		colaMecanico := NewColaPrioritaria()
		colaLimpieza := NewColaPrioritaria()
		colaRevision := NewColaPrioritaria()

		// Canales con capacidad
		semPlazas := make(chan struct{}, NumPlazas)
		semLimp := make(chan struct{}, NumPlazas)
		semRev := make(chan struct{}, NumPlazas)
		semMec := make(chan struct{}, NumMecanicos)
		for i := 0; i < NumPlazas; i++ {
			semPlazas <- struct{}{}
			semLimp <- struct{}{}
			semRev <- struct{}{}
		}
		for i := 0; i < NumMecanicos; i++ {
			semMec <- struct{}{}
		}

		// Lanzar workers
		for i := 0; i < NumPlazas; i++ {
			go s.workerEntrada(colaEntrada, colaMecanico, semPlazas, metricas, tiempoPorVehiculo, aux)
		}
		for i := 0; i < NumMecanicos; i++ {
			go s.workerMecanico(colaMecanico, colaLimpieza, semMec, metricas, tiempoPorVehiculo, aux)
		}
		for i := 0; i < NumPlazas; i++ {
			go s.workerLimpieza(colaLimpieza, colaRevision, semLimp, metricas, tiempoPorVehiculo, aux)
		}
		for i := 0; i < NumPlazas; i++ {
			go s.workerRevision(colaRevision, semRev, &wgFinal, metricas, tiempoPorVehiculo, aux)
		}

		// Encolar vehículos en la cola de entrada
		for _, v := range vehiculos {
			colaEntrada.Push(v)
		}

		// Esperar a que todos los vehículos terminen la última fase
		wgFinal.Wait()
		close(s.Done)
		fmt.Printf("=== FIN SIMULACIÓN WaitGroup %d ===\n", sim)

		metricas.Fin = time.Now()
	}
}
