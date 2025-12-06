package sim

import (
	"fmt"
	"time"
)

// Variable global para guardar resultados
var resultados []ResultadoSimulacion

// Estructura global para guardar resultados
type ResultadoSimulacion struct {
	NombreEscenario  string
	TiempoTotal      time.Duration
	TiempoMedioVeh   time.Duration
	VehiculosPorFase map[Fase]int
}

func registrarResultado(nombre string, metricas *Metricas) {
	totalVehiculos := len(metricas.TiemposPorVehiculo)
	tiempoTotal := time.Duration(0)
	for _, t := range metricas.TiemposPorVehiculo {
		tiempoTotal += t
	}
	resultados = append(resultados, ResultadoSimulacion{
		NombreEscenario:  nombre,
		TiempoTotal:      metricas.Fin.Sub(metricas.Inicio),
		TiempoMedioVeh:   tiempoTotal / time.Duration(totalVehiculos),
		VehiculosPorFase: metricas.VehiculosPorFase,
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
