package sim

import (
	"fmt"
	"testing"
	"time"
)

// Ejecuta los escenarios con ambos simuladores y compara resultados
func TestComparativaEscenarios(t *testing.T) {
	escenarios := []struct{ numA, numB, numC int }{
		{10, 10, 10},
		{20, 5, 5},
		{5, 5, 20},
	}

	for _, esc := range escenarios {
		fmt.Printf("\n=== ESCENARIO %dA/%dB/%dC ===\n", esc.numA, esc.numB, esc.numC)

		fmt.Println("\n--- Simulación WaitGroup ---")
		wgSim := NewSimuladorWaitGroup(nil)
		EjecutarEscenarioSimulador(t, wgSim, esc.numA, esc.numB, esc.numC)

		fmt.Println("\n--- Simulación RWMutex ---")
		rwSim := NewSimuladorRWMutex(nil)
		EjecutarEscenarioSimulador(t, rwSim, esc.numA, esc.numB, esc.numC)
	}

	imprimirTablaResultados()
}

// Ejecuta un escenario con un simulador genérico
func EjecutarEscenarioSimulador(t *testing.T, simulador ISimulador, numA, numB, numC int) {
	//NO imprimir los cambios de fase (muchas lineas, dificil ver bien el test)
	simulador.SetVerbose(false)

	// Generar vehículos
	vehiculos := GenerarVehiculosPorCategorias(numA, numB, numC)
	totalVehiculos := len(vehiculos) // 30

	// Inicializar métricas
	metricas := NuevaMetricas()
	tiempoPorVehiculo := NuevaTiempoVehiculo()
	aux := inicializarMetricasAux()

	// Cada simulador debe registrar sus propios tiempos dentro de RunSim
	// maxEsperas no se usa actualmente, pero lo dejamos para compatibilidad
	maxEsperas := map[Fase]int{
		FaseEntrada:  0,
		FaseAtencion: 0,
		FaseLimpieza: 0,
		FaseRevision: 0,
	}

	// Ejecutar simulación
	simulador.RunSim(vehiculos, 1, totalVehiculos, 10, 2, maxEsperas, metricas, tiempoPorVehiculo, aux)

	imprimirMetricasPorFase(aux)

	// Mostrar resumen
	fmt.Printf("Vehículos generados: %dA/%dB/%dC (total %d)\n", numA, numB, numC, totalVehiculos)
	fmt.Printf("Tiempo total simulación: %v\n", metricas.Fin.Sub(metricas.Inicio))

	// Tiempos por vehículo
	tiempoTotal := time.Duration(0)
	for _, t := range metricas.TiemposPorVehiculo {
		tiempoTotal += t
	}
	if totalVehiculos > 0 {
		fmt.Printf("Tiempo promedio por vehículo: %v\n", tiempoTotal/time.Duration(totalVehiculos))
	}

	// Tiempo por categoría
	imprimirTiempoPorCategoria(vehiculos, tiempoPorVehiculo)

	// Guardar resultados
	registrarResultado(fmt.Sprintf("%dA/%dB/%dC", numA, numB, numC), metricas)
}
