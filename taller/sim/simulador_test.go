package sim

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestComparativaEscenarios(t *testing.T) {
	EjecutarEscenario(t, 10, 10, 10)
	EjecutarEscenario(t, 20, 5, 5)
	EjecutarEscenario(t, 5, 5, 20)
	imprimirTablaResultados()
}

// Función auxiliar para ejecutar un escenario con métricas
func EjecutarEscenario(t *testing.T, numA, numB, numC int) {

	//Preparar entorno
	nplazas := 10
	nmecanicos := 2

	vehiculos := GenerarVehiculosPorCategorias(numA, numB, numC)
	rs := inicializarRecursos(nplazas, nmecanicos)

	fmt.Printf("\n--- Simulación con %d plazas y %d mecánicos (%dA / %dB / %dC) (var.máx. %.2f) ---\n", nplazas, nmecanicos, numA, numB, numC, variacionMax)

	metricas := NuevaMetricas()
	metricas.Inicio = time.Now()
	//	Metricas auxiliares
	aux := inicializarMetricasAux()
	tiempoPorVehiculo := NuevaTiempoVehiculo()

	// WaitGroup
	var wgFinal sync.WaitGroup
	wgFinal.Add(len(vehiculos))

	// Lanzar workers
	for i := 0; i < rs.NumPlazas; i++ {
		LanzarWorkerMetricas(rs.ColaEntrada, rs.ColaMecanico, rs.SemPlazas, FaseEntrada, metricas, aux, tiempoPorVehiculo, nil)
		LanzarWorkerMetricas(rs.ColaLimpieza, rs.ColaRevision, rs.SemLimp, FaseLimpieza, metricas, aux, tiempoPorVehiculo, nil)
		LanzarWorkerMetricas(rs.ColaRevision, nil, rs.SemRev, FaseRevision, metricas, aux, tiempoPorVehiculo, &wgFinal)
	}

	for i := 0; i < rs.NumMecanicos; i++ {
		LanzarWorkerMetricas(rs.ColaMecanico, rs.ColaLimpieza, rs.SemMec, FaseAtencion, metricas, aux, tiempoPorVehiculo, nil)
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
