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

	Verbose bool          //para no imprimir en test y solo en simulacion
	Done    chan struct{} // para parar workers
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

// workerEntrada: espera tokens en semPlazas, procesa y pasa a colaMec
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

		// ocupa plaza
		<-semPlazas

		s.imprimirVehiculo(v, FaseEntrada, "Entra plaza")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseEntrada, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseEntrada].Registrar(duracion)

		s.imprimirVehiculo(v, FaseEntrada, "Sale plaza")

		// libera plaza
		semPlazas <- struct{}{}

		push(&s.mtxMec, &s.colaMec, v)
	}
}

// workerMecanico: usa semMec
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

		<-semMec // ocupa mecánico

		s.imprimirVehiculo(v, FaseAtencion, "Atendido por mecánico")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseAtencion, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseAtencion].Registrar(duracion)

		s.imprimirVehiculo(v, FaseAtencion, "Finaliza mecánico")

		semMec <- struct{}{} // libera mecánico

		push(&s.mtxLimp, &s.colaLimp, v)
	}
}

// workerLimpieza: usa semLimp
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

		<-semLimp // ocupa plaza de limpieza

		s.imprimirVehiculo(v, FaseLimpieza, "Limpiando")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseLimpieza, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseLimpieza].Registrar(duracion)

		s.imprimirVehiculo(v, FaseLimpieza, "Limpieza finalizada")

		semLimp <- struct{}{} // libera plaza limpieza

		push(&s.mtxRev, &s.colaRev, v)
	}
}

// workerRevision: usa semRev y es el que marca finalización (wgFinal.Done)
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

		<-semRev // ocupa plaza de revisión

		s.imprimirVehiculo(v, FaseRevision, "Revisión")

		start := time.Now()
		time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
		duracion := time.Since(start)

		metricas.RegistrarVehiculo(v, FaseRevision, duracion)
		tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)
		aux[FaseRevision].Registrar(duracion)

		s.imprimirVehiculo(v, FaseRevision, "Vehículo entregado")

		semRev <- struct{}{} // libera plaza revisión

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
		fmt.Printf("\n=== SIMULACIÓN RWMutex %d ===\n", sim)

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

		// Poblar cola de entrada (protegida)
		s.mtxEntrada.Lock()
		s.colaEntrada = append(s.colaEntrada, vehiculos...)
		s.mtxEntrada.Unlock()

		// semáforos compartidos (canales con tokens)
		semPlazas := make(chan struct{}, NumPlazas)
		semLimp := make(chan struct{}, NumPlazas)
		semRev := make(chan struct{}, NumPlazas)
		semMec := make(chan struct{}, NumMecanicos)

		// inicializar tokens
		for i := 0; i < NumPlazas; i++ {
			semPlazas <- struct{}{}
			semLimp <- struct{}{}
			semRev <- struct{}{}
		}
		for i := 0; i < NumMecanicos; i++ {
			semMec <- struct{}{}
		}

		// wgFinal: contaremos el total de vehículos que terminan la última fase
		var wgFinal sync.WaitGroup
		wgFinal.Add(len(vehiculos))

		// Lanzar pools de workers: Entrada y Limpieza usan NumPlazas, Mecanicos usan NumMecanicos, Revision usa NumPlazas (o NumPlazas)
		// Entrada
		for i := 0; i < NumPlazas; i++ {
			go s.workerEntrada(semPlazas, metricas, tiempoPorVehiculo, aux)
		}
		// Mecanicos
		for i := 0; i < NumMecanicos; i++ {
			go s.workerMecanico(semMec, metricas, tiempoPorVehiculo, aux)
		}
		// Limpieza
		for i := 0; i < NumPlazas; i++ {
			go s.workerLimpieza(semLimp, metricas, tiempoPorVehiculo, aux)
		}
		// Revision
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
