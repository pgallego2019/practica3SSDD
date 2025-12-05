package sim

import (
	"sync"
	"testing"
	"time"
	"taller/models"
	"fmt"
	"math/rand"
)

// Inicializa estructuras para métricas adicionales
func inicializarMetricasAux() map[Fase]*MetricasFase {
	return map[Fase]*MetricasFase{
		FaseEntrada:   &MetricasFase{Min: time.Hour, Max: 0},
		FaseAtencion:  &MetricasFase{Min: time.Hour, Max: 0},
		FaseLimpieza:  &MetricasFase{Min: time.Hour, Max: 0},
		FaseRevision:  &MetricasFase{Min: time.Hour, Max: 0},
	}
}

// Métricas por fase
type MetricasFase struct {
	Min, Max, Total time.Duration
	Contador        int
	mutex           sync.Mutex
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

// Muestra métricas de todas las fases
func imprimirMetricasPorFase(aux map[Fase]*MetricasFase) {
	fmt.Println("=== Métricas por fase ===")
	for fase, m := range aux {
		fmt.Printf("%s: Min %v | Promedio %v | Max %v | Vehículos %d\n",
			fase, m.Min, m.Promedio(), m.Max, m.Contador)
	}
}

// Registro de tiempo total por vehículo
type TiempoVehiculo struct {
	Tiempos map[int]time.Duration
	mutex   sync.Mutex
}

func NuevaTiempoVehiculo() *TiempoVehiculo {
	return &TiempoVehiculo{Tiempos: make(map[int]time.Duration)}
}

func (tv *TiempoVehiculo) Registrar(id int, duracion time.Duration) {
	tv.mutex.Lock()
	defer tv.mutex.Unlock()
	tv.Tiempos[id] += duracion
}

// Distribución por categoría
func imprimirTiempoPorCategoria(vehiculos []*models.Vehiculo, tv *TiempoVehiculo) {
	categorias := map[models.Especialidad][]time.Duration{
		models.Mecanica:  {},
		models.Electrica: {},
		models.Carroceria: {},
	}

	for _, v := range vehiculos {
		if t, ok := tv.Tiempos[v.Incidencia.ID]; ok {
			categorias[v.Incidencia.Tipo] = append(categorias[v.Incidencia.Tipo], t)
		}
	}

	fmt.Println("=== Promedio por categoría ===")
	for cat, tiempos := range categorias {
		total := time.Duration(0)
		for _, t := range tiempos {
			total += t
		}
		prom := time.Duration(0)
		if len(tiempos) > 0 {
			prom = total / time.Duration(len(tiempos))
		}
		fmt.Printf("%s: Promedio %v (%d vehículos)\n", cat, prom, len(tiempos))
	}
}


// Recursos de sinc
type RecursosSim struct {
	ColaEntrada, ColaMecanico, ColaLimpieza, ColaRevision *ColaPrioritaria
	SemPlazas, SemLimp, SemRev, SemMec chan struct{}
	NumPlazas, NumMecanicos int
}

func inicializarRecursos(numPlazas, numMecanicos int) *RecursosSim {
	rs := &RecursosSim{
		ColaEntrada: NewColaPrioritaria(),
		ColaMecanico: NewColaPrioritaria(),
		ColaLimpieza: NewColaPrioritaria(),
		ColaRevision: NewColaPrioritaria(),
		SemPlazas: make(chan struct{}, numPlazas),
		SemLimp: make(chan struct{}, numPlazas),
		SemRev: make(chan struct{}, numPlazas),
		SemMec: make(chan struct{}, numMecanicos),
		NumPlazas: numPlazas,
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

// Estructura global para guardar resultados
type ResultadoSimulacion struct {
	NombreEscenario    string
	TiempoTotal        time.Duration
	TiempoMedioVeh  time.Duration
	VehiculosPorFase   map[Fase]int
}

func registrarResultado(nombre string, metricas *Metricas) {
	totalVehiculos := len(metricas.TiemposPorVehiculo)
	tiempoTotal := time.Duration(0)
	for _, t := range metricas.TiemposPorVehiculo {
		tiempoTotal += t
	}
	resultados = append(resultados, ResultadoSimulacion{
		NombreEscenario:   nombre,
		TiempoTotal:       metricas.Fin.Sub(metricas.Inicio),
		TiempoMedioVeh: tiempoTotal / time.Duration(totalVehiculos),
		VehiculosPorFase:  metricas.VehiculosPorFase,
	})
}

func imprimirTablaResultados() {
	fmt.Printf("\n===== COMPARATIVA FINAL DE SIMULACIONES =====\n")
	fmt.Printf("%-15s %-15s %-20s %-10s %-10s %-10s %-10s\n",
		"Escenario", "Tiempo Total", "Tiempo Promedio Veh", "Entrada", "Atencion", "Limpieza", "Revision")
	for _, r := range resultados {
		fmt.Printf("%-15s %-15v %-20v %-10d %-10d %-10d %-10d\n",
			r.NombreEscenario,
			r.TiempoTotal,
			r.TiempoMedioVeh,
			r.VehiculosPorFase[FaseEntrada],
			r.VehiculosPorFase[FaseAtencion],
			r.VehiculosPorFase[FaseLimpieza],
			r.VehiculosPorFase[FaseRevision],
		)
	}
}

func TestComparativaEscenarios(t *testing.T) {
	runTestEscenario(t, 10, 10, 10)
	runTestEscenario(t, 20, 5, 5)
	runTestEscenario(t, 5, 5, 20)
	imprimirTablaResultados()
}


var resultados []ResultadoSimulacion

// Estructura para almacenar métricas de la simulación
type Metricas struct {
	VehiculosPorFase map[Fase]int
	TiemposPorVehiculo map[int]time.Duration
	mutex sync.Mutex
	Inicio time.Time
	Fin time.Time
}

func NuevaMetricas() *Metricas {
	return &Metricas{
		VehiculosPorFase: make(map[Fase]int),
		TiemposPorVehiculo: make(map[int]time.Duration),
	}
}

// Genera los vehículos pedidos por categoría
func GenerarVehiculosPorCategorias(numA, numB, numC int) []*models.Vehiculo {
    vehiculos := []*models.Vehiculo{}
    id := 1

    // Categoría A - Mecánica
    for i := 0; i < numA; i++ {
        vehiculos = append(vehiculos, &models.Vehiculo{
            Matricula: fmt.Sprintf("A-%03d", id),
            Marca: "MarcaX",
            Modelo: "ModeloY",
            FechaEntrada: time.Now().Format("2006-01-02 15:04:05"),
            Incidencia: &models.Incidencia{
                ID: id,
                Tipo: models.Mecanica,
                TiempoFase: 1,
            },
        })
        id++
    }

    // Categoría B - Eléctrica
    for i := 0; i < numB; i++ {
        vehiculos = append(vehiculos, &models.Vehiculo{
            Matricula: fmt.Sprintf("B-%03d", id),
            Marca: "MarcaX",
            Modelo: "ModeloY",
            FechaEntrada: time.Now().Format("2006-01-02 15:04:05"),
            Incidencia: &models.Incidencia{
                ID: id,
                Tipo: models.Electrica,
                TiempoFase: 1,
            },
        })
        id++
    }

    // Categoría C - Carrocería
    for i := 0; i < numC; i++ {
        vehiculos = append(vehiculos, &models.Vehiculo{
            Matricula: fmt.Sprintf("C-%03d", id),
            Marca: "MarcaX",
            Modelo: "ModeloY",
            FechaEntrada: time.Now().Format("2006-01-02 15:04:05"),
            Incidencia: &models.Incidencia{
                ID: id,
                Tipo: models.Carroceria,
                TiempoFase: 1,
            },
        })
        id++
    }

    return vehiculos
}

// Mezcla aleatoriamente un slice de vehículos
func MezclarVehiculos(vehiculos []*models.Vehiculo) []*models.Vehiculo {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    r.Shuffle(len(vehiculos), func(i, j int) {
        vehiculos[i], vehiculos[j] = vehiculos[j], vehiculos[i]
    })
    return vehiculos
}

func (m *Metricas) RegistrarVehiculo(v *models.Vehiculo, fase Fase, duracion time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.VehiculosPorFase[fase]++
	m.TiemposPorVehiculo[v.Incidencia.ID] += duracion
}

func lanzarWorker(colaIn, colaOut *ColaPrioritaria, sem chan struct{}, fase Fase, metricas *Metricas, wg *sync.WaitGroup) {
	go func() {
		for {
			<-colaIn.notify
			v := colaIn.PopFront()
			if v == nil { continue }
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

func prepararVehiculos(numA, numB, numC int) []*models.Vehiculo {
	vehiculos := GenerarVehiculosPorCategorias(numA, numB, numC)
	return MezclarVehiculos(vehiculos)
}

// Función auxiliar para ejecutar un escenario con métricas
func runTestEscenario(t *testing.T, numA, numB, numC int) {
	vehiculos := prepararVehiculos(numA, numB, numC)
	rs := inicializarRecursos(2, 2)
	metricas := NuevaMetricas()
	metricas.Inicio = time.Now()
	
	//aux
	aux := inicializarMetricasAux()
	tiempoPorVehiculo := NuevaTiempoVehiculo()
	//aux

	var wgFinal sync.WaitGroup
	wgFinal.Add(len(vehiculos))


	// Función genérica de worker que registra métricas
	lanzarWorkerMetricas := func(colaIn, colaOut *ColaPrioritaria, sem chan struct{}, fase Fase, finalWg *sync.WaitGroup) {
		go func() {
			for {
				<-colaIn.notify
				v := colaIn.PopFront()
				if v == nil { continue }
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

	// Lanzar workers
	for i := 0; i < rs.NumPlazas; i++ {
		lanzarWorkerMetricas(rs.ColaEntrada, rs.ColaMecanico, rs.SemPlazas, FaseEntrada, nil)
		lanzarWorkerMetricas(rs.ColaLimpieza, rs.ColaRevision, rs.SemLimp, FaseLimpieza, nil)
		lanzarWorkerMetricas(rs.ColaRevision, nil, rs.SemRev, FaseRevision, &wgFinal)
	}
	for i := 0; i < rs.NumMecanicos; i++ {
		lanzarWorkerMetricas(rs.ColaMecanico, rs.ColaLimpieza, rs.SemMec, FaseAtencion, nil)
	}

	// Encolar vehículos
	for _, v := range vehiculos {
		rs.ColaEntrada.Push(v)
	}

	// Esperar finalización
	wgFinal.Wait()
	metricas.Fin = time.Now()

	// Registrar y mostrar métricas
	fmt.Printf("\n=== MÉTRICAS ESCENARIO %dA/%dB/%dC ===\n", numA, numB, numC)
	fmt.Printf("Tiempo total simulación: %v\n", metricas.Fin.Sub(metricas.Inicio))
	for fase, num := range metricas.VehiculosPorFase {
		fmt.Printf("Vehículos atendidos fase %s: %d\n", fase, num)
	}
	totalVehiculos := len(vehiculos)
	tiempoTotal := time.Duration(0)
	for _, t := range metricas.TiemposPorVehiculo {
		tiempoTotal += t
	}
	fmt.Printf("Tiempo promedio por vehículo: %v\n", tiempoTotal/time.Duration(totalVehiculos))

	imprimirMetricasPorFase(aux)
	imprimirTiempoPorCategoria(vehiculos, tiempoPorVehiculo)

	registrarResultado(fmt.Sprintf("%dA/%dB/%dC", numA, numB, numC), metricas)
}

