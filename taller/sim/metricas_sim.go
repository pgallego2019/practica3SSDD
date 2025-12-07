package sim

import (
	"fmt"
	"sync"
	"taller/models"
	"time"
)

// Estructura para almacenar métricas de la simulación
type Metricas struct {
	VehiculosPorFase   map[Fase]int
	TiemposPorVehiculo map[int]time.Duration
	mutex              sync.RWMutex //CAMBIO POR RW
	Inicio             time.Time
	Fin                time.Time
	Simulador          string
}

func NuevaMetricas() *Metricas {
	return &Metricas{
		VehiculosPorFase:   make(map[Fase]int),
		TiemposPorVehiculo: make(map[int]time.Duration),
	}
}

// Métricas por fase
type MetricasFase struct {
	Min, Max, Total time.Duration
	Contador        int
	mutex           sync.Mutex
}

// Inicializa estructuras para métricas adicionales
func inicializarMetricasAux() map[Fase]*MetricasFase {
	return map[Fase]*MetricasFase{
		FaseEntrada:  {Min: time.Hour, Max: 0},
		FaseAtencion: {Min: time.Hour, Max: 0},
		FaseLimpieza: {Min: time.Hour, Max: 0},
		FaseRevision: {Min: time.Hour, Max: 0},
	}
}

// Actualiza métricas por fase
func (m *MetricasFase) Registrar(duracion time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if duracion < m.Min {
		m.Min = duracion
	}
	if duracion > m.Max {
		m.Max = duracion
	}
	m.Total += duracion
	m.Contador++
}

// Devuelve tiempo promedio por fase
func (m *MetricasFase) Promedio() time.Duration {
	if m.Contador == 0 {
		return 0
	}
	return m.Total / time.Duration(m.Contador)
}

func (m *Metricas) RegistrarVehiculo(v *models.Vehiculo, fase Fase, duracion time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.VehiculosPorFase[fase]++
	m.TiemposPorVehiculo[v.Incidencia.ID] += duracion
}

// Muestra métricas de todas las fases
func imprimirMetricasPorFase(aux map[Fase]*MetricasFase) {
	fmt.Println("=== Métricas por fase ===")
	for fase, m := range aux {
		fmt.Printf("%s: Min %v | Promedio %v | Max %v | Vehículos %d\n",
			fase, m.Min, m.Promedio(), m.Max, m.Contador)
	}
}

// Función genérica de worker que registra métricas
func LanzarWorkerMetricas(
	colaIn, colaOut *ColaPrioritaria,
	sem chan struct{},
	fase Fase,
	metricas *Metricas,
	aux map[Fase]*MetricasFase,
	tiempoPorVehiculo *TiempoVehiculo,
	finalWg *sync.WaitGroup,
) {
	go func() {
		for {
			<-colaIn.notify
			v := colaIn.PopFront()
			if v == nil {
				continue
			}

			// Imprimir solo al inicio de la fase (intentando no imprimir demasiadas lineas)
			fmt.Printf("Tiempo %v | Coche %s | Incidencia %s | Inicio %s\n",
				time.Since(metricas.Inicio).Truncate(time.Millisecond),
				v.Matricula,
				v.Incidencia.Tipo,
				fase,
			)

			<-sem
			start := time.Now()
			time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
			sem <- struct{}{}

			duracion := time.Since(start)

			metricas.RegistrarVehiculo(v, fase, duracion)
			aux[fase].Registrar(duracion)
			tiempoPorVehiculo.Registrar(v.Incidencia.ID, duracion)

			if colaOut != nil {
				colaOut.Push(v)
			} else if finalWg != nil {
				finalWg.Done()
			}
		}
	}()
}
