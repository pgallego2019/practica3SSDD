package sim

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"taller/models"
	"time"
)

// Para representar las 4 fases por las que pasan los vehículos
type Fase int

const (
	FaseEntrada Fase = iota + 1
	FaseAtencion
	FaseLimpieza
	FaseRevision
)

func (f Fase) String() string {
	switch f {
	case FaseEntrada:
		return "Entrada"
	case FaseAtencion:
		return "Atencion"
	case FaseLimpieza:
		return "Limpieza"
	case FaseRevision:
		return "Revision"
	default:
		return "Desconocida"
	}
}

const variacionMax = 0.15 // Mejor entre 10-20%

// Muestra el estado del vehículo en cada fase
func (s *Simulador) imprimirVehiculo(v *models.Vehiculo, fase Fase, estado string) {
	elapsed := time.Since(s.Start).Truncate(time.Millisecond)
	fmt.Printf("Tiempo %v | Vehiculo %s | Incidencia %s | Fase %s | Estado %s\n",
		elapsed, v.Matricula, v.Incidencia.Tipo, fase.String(), estado)
}

// varía el tiempo de fase según una variación máxima
func variacionTiempoFase(tiempoBase int) time.Duration {
	r := (rand.Float64()*2 - 1) * variacionMax

	variacion := float64(tiempoBase) * r
	tiempoFinal := float64(tiempoBase) + variacion

	if tiempoFinal < 0 {
		tiempoFinal = float64(tiempoBase)
	}

	return time.Duration(tiempoFinal * float64(time.Second))
}

// Ahora que tengo colas de prioridad tengo que asegurar que los vehiculos se atienden por orden de prioridad
// Los mecanicos deben atender primero los de mecanica, luego los electricos y por ultimo los de carroceria
func (s *Simulador) RunSim(Sims int, Nvehiculos int, NumPlazas int, NumMecanicos int, maxEsperas map[Fase]int) {

	for sim := 1; sim <= Sims; sim++ {
		fmt.Printf("\n=== SIMULACIÓN %d ===\n", sim)

		s.Start = time.Now()
		s.Done = make(chan struct{})
		vehiculos := GenerarVehiculosAleatorios(Nvehiculos)
		ImprimirResumenCategorias(vehiculos)
		// NO meter en Simulador el WG
		//s.WG = sync.WaitGroup{}
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
			go s.workerEntrada(colaEntrada, colaMecanico, semPlazas)
		}
		for i := 0; i < NumMecanicos; i++ {
			go s.workerMecanico(colaMecanico, colaLimpieza, semMec)
		}
		for i := 0; i < NumPlazas; i++ {
			go s.workerLimpieza(colaLimpieza, colaRevision, semLimp)
		}
		for i := 0; i < NumPlazas; i++ {
			go s.workerRevision(colaRevision, semRev, &wgFinal)
		}

		// Encolar vehículos en la cola de entrada
		for _, v := range vehiculos {
			colaEntrada.Push(v)
		}

		// Esperar a que todos los vehículos terminen la última fase
		wgFinal.Wait()
		close(s.Done)
		fmt.Printf("=== FIN SIMULACIÓN %d ===\n", sim)
	}
}

// Inicia una simulacion con parámetros de prueba
func SimularTaller(t *models.Taller) {
	reader := bufio.NewReader(os.Stdin)

	// Valores por defecto
	N := 10
	NumPlazas := 2
	NumMecanicos := 1
	Sims := 1

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

	// Validación
	if N <= 0 || NumPlazas <= 0 || NumMecanicos <= 0 || Sims <= 0 {
		fmt.Println("Error: todos los valores deben ser mayores que cero.")
		return
	}

	maxEsperas := map[Fase]int{
		FaseEntrada:  0,
		FaseAtencion: 0,
		FaseLimpieza: 0,
		FaseRevision: 0,
	}

	simulador := NewSimulador(t)

	fmt.Println("=== INICIANDO TEST DEL TALLER ===")
	simulador.RunSim(Sims, N, NumPlazas, NumMecanicos, maxEsperas)
	fmt.Println("=== TEST DEL TALLER FINALIZADO ===")
}
