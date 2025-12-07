package sim

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"taller/models"
)

// Inicia una simulacion con parámetros de prueba
func SimularTaller(t *models.Taller) {
	reader := bufio.NewReader(os.Stdin)
	var simulador ISimulador

	// Valores por defecto
	N := 10
	NumPlazas := 2
	NumMecanicos := 1
	Sims := 1
	opcion := 1

	fmt.Printf("Introduce el número de vehículos (por defecto %d): ", N)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		fmt.Sscan(input, &N)
	}

	fmt.Printf("Introduce el número de plazas disponibles (por defecto %d): ", NumPlazas)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		fmt.Sscan(input, &NumPlazas)
	}

	fmt.Printf("Introduce el número de mecánicos (por defecto %d): ", NumMecanicos)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		fmt.Sscan(input, &NumMecanicos)
	}

	fmt.Printf("Introduce cuántas simulaciones quieres ejecutar (por defecto %d): ", Sims)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		fmt.Sscan(input, &Sims)
	}

	fmt.Println("Elige el tipo de simulador:")
	fmt.Println("  1 - Simulador WaitGroup con colas prioritarias")
	fmt.Println("  2 - Simulador con RWMutex")
	fmt.Print("Opción (1 por defecto): ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		fmt.Sscan(input, &opcion)
	}

	// Validación
	if N <= 0 || NumPlazas <= 0 || NumMecanicos <= 0 || Sims <= 0 {
		fmt.Println("Error: todos los valores deben ser mayores que cero.")
		return
	}

	// Generación de N vehículos aleatorios
	vehiculos := GenerarVehiculosAleatorios(N)

	// Elegir simulador
	switch opcion {
	case 2:
		simulador = NewSimuladorRWMutex(t)
	default:
		simulador = NewSimuladorWaitGroup(t)
	}

	// Activar la opción de imprimir los mensajes
	simulador.SetVerbose(true)

	maxEsperas := map[Fase]int{
		FaseEntrada:  0,
		FaseAtencion: 0,
		FaseLimpieza: 0,
		FaseRevision: 0,
	}

	fmt.Println("=== INICIANDO TEST DEL TALLER ===")
	simulador.RunSim(vehiculos, Sims, N, NumPlazas, NumMecanicos, maxEsperas, nil, nil, nil)
	fmt.Println("=== TEST DEL TALLER FINALIZADO ===")
}
