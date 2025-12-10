package sim

import (
	"fmt"
	"sync"
	"taller/models"
	"time"
)

type SimuladorRWMutex struct {
	Taller *models.Taller
	Start  time.Time

	colaEntrada []*models.Vehiculo
	colaMec     []*models.Vehiculo
	colaLimp    []*models.Vehiculo
	colaRev     []*models.Vehiculo

	mtxEntrada sync.RWMutex
	mtxMec     sync.RWMutex
	mtxLimp    sync.RWMutex
	mtxRev     sync.RWMutex

	Verbose bool // Para no imprimir en test y solo en simulacion
	Done    chan struct{}
}

func (s *SimuladorRWMutex) SetVerbose(v bool) {
	s.Verbose = v
}

func NewSimuladorRWMutex(t *models.Taller) *SimuladorRWMutex {
	return &SimuladorRWMutex{
		Taller: t,
	}
}

func push(m *sync.RWMutex, cola *[]*models.Vehiculo, v *models.Vehiculo) {
	m.Lock()
	*cola = append(*cola, v)
	m.Unlock()
}

func pop(m *sync.RWMutex, cola *[]*models.Vehiculo) *models.Vehiculo {
	m.Lock()
	defer m.Unlock()

	if len(*cola) == 0 {
		return nil
	}

	v := (*cola)[0]
	*cola = (*cola)[1:]
	return v
}

func (s *SimuladorRWMutex) imprimirVehiculo(v *models.Vehiculo, fase Fase, estado string) {
	if !s.Verbose {
		return
	}
	elapsed := time.Since(s.Start).Truncate(time.Millisecond)
	fmt.Printf("Tiempo %v | Vehiculo %s | Incidencia %s | Fase %s | Estado %s\n",
		elapsed, v.Matricula, v.Incidencia.Tipo, fase.String(), estado)
}

func (s *SimuladorRWMutex) workerEntrada(
	semPlazas chan struct{},
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
) {
	for {
		select {
		case <-s.Done:
			return
		default:
		}

		v := pop(&s.mtxEntrada, &s.colaEntrada)
		if v == nil {
			time.Sleep(3 * time.Millisecond)
			continue
		}

		<-semPlazas

		s.imprimirVehiculo(v, FaseEntrada, "Entra plaza")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseEntrada, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseEntrada].Registrar(duracion)

		s.imprimirVehiculo(v, FaseEntrada, "Sale plaza")

		semPlazas <- struct{}{}

		push(&s.mtxMec, &s.colaMec, v)
	}
}

func (s *SimuladorRWMutex) workerMecanico(
	semMec chan struct{},
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
) {
	for {
		select {
		case <-s.Done:
			return
		default:
		}

		v := pop(&s.mtxMec, &s.colaMec)
		if v == nil {
			time.Sleep(3 * time.Millisecond)
			continue
		}

		<-semMec

		s.imprimirVehiculo(v, FaseAtencion, "Atendido por mecánico")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseAtencion, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseAtencion].Registrar(duracion)

		s.imprimirVehiculo(v, FaseAtencion, "Finaliza mecánico")

		semMec <- struct{}{}

		push(&s.mtxLimp, &s.colaLimp, v)
	}
}

func (s *SimuladorRWMutex) workerLimpieza(
	semLimp chan struct{},
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
) {
	for {
		select {
		case <-s.Done:
			return
		default:
		}

		v := pop(&s.mtxLimp, &s.colaLimp)
		if v == nil {
			time.Sleep(3 * time.Millisecond)
			continue
		}

		<-semLimp

		s.imprimirVehiculo(v, FaseLimpieza, "Limpiando")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseLimpieza, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseLimpieza].Registrar(duracion)

		s.imprimirVehiculo(v, FaseLimpieza, "Limpieza finalizada")

		semLimp <- struct{}{}

		push(&s.mtxRev, &s.colaRev, v)
	}
}

func (s *SimuladorRWMutex) workerRevision(
	semRev chan struct{},
	wgFinal *sync.WaitGroup,
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
) {
	for {
		select {
		case <-s.Done:
			return
		default:
		}

		v := pop(&s.mtxRev, &s.colaRev)
		if v == nil {
			time.Sleep(3 * time.Millisecond)
			continue
		}

		<-semRev

		s.imprimirVehiculo(v, FaseRevision, "Revisión")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseRevision, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseRevision].Registrar(duracion)

		s.imprimirVehiculo(v, FaseRevision, "Vehículo entregado")

		semRev <- struct{}{}

		wgFinal.Done() // vehículo completamente terminado
	}
}

func (s *SimuladorRWMutex) RunSim(
	vehiculos []*models.Vehiculo,
	Sims int,
	N int,
	NumPlazas int,
	NumMecanicos int,
	maxEsperas map[Fase]int,
	metricas *Metricas,
	tiempoPorVehiculo *TiempoVehiculo,
	aux map[Fase]*MetricasFase,
) {
	for sim := 1; sim <= Sims; sim++ {
		fmt.Printf("\n=== SIMULACIÓN RWMutex %d (%d plazas y %d mecánicos) ===\n", sim, NumPlazas, NumMecanicos)

		// Defensive init (opcional, aunque idealmente quien llame debe inicializar)
		if metricas == nil {
			metricas = NuevaMetricas()
		}
		if tiempoPorVehiculo == nil {
			tiempoPorVehiculo = NuevaTiempoVehiculo()
		}
		if aux == nil {
			aux = inicializarMetricasAux()
		}

		metricas.Simulador = "RWMutex"

		s.Start = time.Now()
		metricas.Inicio = time.Now()

		if s.Done == nil {
			s.Done = make(chan struct{})
		}

		ImprimirResumenCategorias(vehiculos)

		s.mtxEntrada.Lock()
		s.colaEntrada = append(s.colaEntrada, vehiculos...)
		s.mtxEntrada.Unlock()

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

		var wgFinal sync.WaitGroup
		wgFinal.Add(len(vehiculos))

		for i := 0; i < NumPlazas; i++ {
			go s.workerEntrada(semPlazas, metricas, tiempoPorVehiculo, aux)
		}
		for i := 0; i < NumMecanicos; i++ {
			go s.workerMecanico(semMec, metricas, tiempoPorVehiculo, aux)
		}
		for i := 0; i < NumPlazas; i++ {
			go s.workerLimpieza(semLimp, metricas, tiempoPorVehiculo, aux)
		}
		for i := 0; i < NumPlazas; i++ {
			go s.workerRevision(semRev, &wgFinal, metricas, tiempoPorVehiculo, aux)
		}

		// Esperar a que todos los vehículos hayan terminado la revisión
		wgFinal.Wait()

		// marcar fin y detener workers
		metricas.Fin = time.Now()
		close(s.Done) // señal para que los workers salgan
		// volver a crear canal Done para próximas simulaciones si se reutiliza el simulador
		s.Done = nil

		fmt.Printf("=== FIN SIMULACIÓN RWMutex %d ===\n", sim)
	}
}
