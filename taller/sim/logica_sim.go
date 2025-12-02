package sim

import (
	"fmt"
	"math/rand"
	"sync"
	"taller/models"
	"time"
)

// Para representar las 4 fases del taller
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

const variacionMax = 0.4 // 40%? TODO ver si es mucho y usarlo

// Para mantener un orden de prioridad en cada fase
/*
type ColaPrioritaria struct {
	altas  []*models.Vehiculo
	medias []*models.Vehiculo
	bajas  []*models.Vehiculo
}

func NewColaPrioritaria() *ColaPrioritaria {
	return &ColaPrioritaria{}
}

func (c *ColaPrioritaria) Push(v *models.Vehiculo) {
	switch v.Incidencia.Tipo {
	case models.Mecanica:
		c.altas = append(c.altas, v)
	case models.Electrica:
		c.medias = append(c.medias, v)
	case models.Carroceria:
		c.bajas = append(c.bajas, v)
	default:
		c.medias = append(c.medias, v)
	}
}

func (c *ColaPrioritaria) PopFront() *models.Vehiculo {
	if len(c.altas) > 0 {
		x := c.altas[0]
		c.altas = c.altas[1:]
		return x
	}
	if len(c.medias) > 0 {
		x := c.medias[0]
		c.medias = c.medias[1:]
		return x
	}
	if len(c.bajas) > 0 {
		x := c.bajas[0]
		c.bajas = c.bajas[1:]
		return x
	}
	return nil
}

func (c *ColaPrioritaria) Len() int {
	return len(c.altas) + len(c.medias) + len(c.bajas)
}

func (c *ColaPrioritaria) FrontEquals(v *models.Vehiculo) bool {
	if len(c.altas) > 0 {
		return c.altas[0] == v
	}
	if len(c.medias) > 0 {
		return c.medias[0] == v
	}
	if len(c.bajas) > 0 {
		return c.bajas[0] == v
	}
	return false
}
*/

// Genera un vehículo con un tipo de incidencia aleatorio
func newSimVehiculo(id int) *models.Vehiculo {
	var esp models.Especialidad
	var tiempo int

	switch rand.Intn(3) {
	case 0:
		esp = models.Mecanica
		tiempo = 5
	case 1:
		esp = models.Electrica
		tiempo = 3
	case 2:
		esp = models.Carroceria
		tiempo = 1
	}

	return &models.Vehiculo{
		Matricula:    fmt.Sprintf("MAT-%03d", id),
		Marca:        "MarcaX",
		Modelo:       "ModeloY",
		FechaEntrada: time.Now().Format("2006-01-02 15:04:05"),
		Incidencia: &models.Incidencia{
			ID:          id,
			Tipo:        esp,
			Descripcion: "",
			Estado:      0,
			TiempoFase:  tiempo,
		},
	}
}

// Genera N vehículos, los separa por categoría y hace shuffle internamente tanto en cada categoría como en las llegadas.
func generarVehiculosAleatorios(N int) []*models.Vehiculo {
	src := rand.NewSource(time.Now().UnixNano()) // semilla local, para que no esté deprecated
	r := rand.New(src)

	vehiculos := make([]*models.Vehiculo, 0, N)
	for i := 1; i <= N; i++ {
		veh := newSimVehiculo(i)
		vehiculos = append(vehiculos, veh)
	}

	// separa por categoría
	var catA, catB, catC []*models.Vehiculo
	for _, v := range vehiculos {
		switch v.Incidencia.Tipo {
		case models.Mecanica:
			catA = append(catA, v)
		case models.Electrica:
			catB = append(catB, v)
		case models.Carroceria:
			catC = append(catC, v)
		}
	}

	// mezclar internamente cada categoría
	r.Shuffle(len(catA), func(i, j int) { catA[i], catA[j] = catA[j], catA[i] })
	r.Shuffle(len(catB), func(i, j int) { catB[i], catB[j] = catB[j], catB[i] })
	r.Shuffle(len(catC), func(i, j int) { catC[i], catC[j] = catC[j], catC[i] })

	// intercalado aleatorio entre categorías
	result := make([]*models.Vehiculo, 0, N)
	iA, iB, iC := 0, 0, 0
	for iA < len(catA) || iB < len(catB) || iC < len(catC) {
		options := []int{}
		if iA < len(catA) {
			options = append(options, 0)
		}
		if iB < len(catB) {
			options = append(options, 1)
		}
		if iC < len(catC) {
			options = append(options, 2)
		}
		ch := options[r.Intn(len(options))]
		switch ch {
		case 0:
			result = append(result, catA[iA])
			iA++
		case 1:
			result = append(result, catB[iB])
			iB++
		case 2:
			result = append(result, catC[iC])
			iC++
		}
	}

	return result
}

// Muestra el estado del vehículo en cada fase
func (s *Simulador) imprimirVehiculo(v *models.Vehiculo, fase Fase, estado string) {
	elapsed := time.Since(s.Start).Truncate(time.Millisecond)
	fmt.Printf("Tiempo %v | Vehiculo %s(ID:%d) | Incidencia %s | Fase %s | Estado %s\n",
		elapsed, v.Matricula, v.Incidencia.ID, v.Incidencia.Tipo, fase.String(), estado)
}

// varía el tiempo de fase según una variación máxima
func variacionTiempoFase(tiempoBase int) time.Duration {
	r := (rand.Float64()*2 - 1) * variacionMax // rango [-variacionMax, +variacionMax]

	variacion := float64(tiempoBase) * r
	tiempoFinal := float64(tiempoBase) + variacion

	// evitar tiempos negativos pq puede salir cero en rand
	// si pasa eso, ¿deberia devolver cero o dejarlo sin variar y devolver tiempoBase?
	if tiempoFinal < 0 {
		tiempoFinal = 0
	}

	return time.Duration(tiempoFinal * float64(time.Second))
}

// Simulacion con semáforos porque tengo límites con NumPlazas y NumMecanicos.
// Si uso mtx no limito eso, solo aseguro exclusion mutua
func (s *Simulador) RunSim(Sims int, Nvehiculos int, NumPlazas int, NumMecanicos int, maxEsperas map[Fase]int) {

	for sim := 1; sim <= Sims; sim++ {
		fmt.Printf("\n=== SIMULACIÓN %d ===\n", sim)

		s.Start = time.Now()
		vehiculos := generarVehiculosAleatorios(Nvehiculos)
		s.WG = sync.WaitGroup{}

		//  Semáforos añadidos
		// NumPlazas es la capacidad total del taller. limpieza y revision tienen la misma capacidad
		semPlazas := make(chan struct{}, NumPlazas)
		semMec := make(chan struct{}, NumMecanicos)
		semLimp := make(chan struct{}, NumPlazas)
		semRev := make(chan struct{}, NumPlazas)

		// Tomo NumPlazas como plazas totales del taller

		for i := 0; i < NumPlazas; i++ {
			semPlazas <- struct{}{}
		}
		for i := 0; i < NumMecanicos; i++ {
			semMec <- struct{}{}
		}
		for i := 0; i < NumPlazas; i++ {
			semLimp <- struct{}{}
		}
		for i := 0; i < NumPlazas; i++ {
			semRev <- struct{}{}
		}

		//  Procesamiento vehículos
		for _, v := range vehiculos {
			s.WG.Add(1)

			go func(v *models.Vehiculo) {
				defer s.WG.Done() //si no limpio me da fatal error: all goroutines are asleep - deadlock! cuando se acaba la simulación

				// --------- Fase 1: PLAZA ----------
				<-semPlazas // ocupa plaza
				s.imprimirVehiculo(v, FaseEntrada, "Entra plaza")
				//time.Sleep(time.Duration(v.Incidencia.TiempoFase) * time.Second) //aqui es donde tengo que usar la variacionMax?
				time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
				s.imprimirVehiculo(v, FaseEntrada, "Sale plaza")
				semPlazas <- struct{}{} // libera plaza

				// --------- Fase 2: MECÁNICO ----------
				<-semMec // ocupa mecánico
				s.imprimirVehiculo(v, FaseAtencion, "Atendido por mecánico")
				time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
				s.imprimirVehiculo(v, FaseAtencion, "Finaliza mecánico")
				semMec <- struct{}{} // libera mecánico

				// --------- Fase 3: LIMPIEZA ----------
				<-semLimp // ocupa limpieza
				s.imprimirVehiculo(v, FaseLimpieza, "Limpiando")
				time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
				s.imprimirVehiculo(v, FaseLimpieza, "Limpieza finalizada")
				semLimp <- struct{}{} // libera limpieza

				// --------- Fase 4: REVISIÓN ----------
				<-semRev // ocupa revisión
				s.imprimirVehiculo(v, FaseRevision, "Revisión")
				time.Sleep(variacionTiempoFase(v.Incidencia.TiempoFase))
				s.imprimirVehiculo(v, FaseRevision, "Vehículo entregado")
				semRev <- struct{}{} // libera revisión

			}(v)
		}

		s.WG.Wait()
		fmt.Printf("=== FIN SIMULACIÓN %d ===\n", sim)
	}
}

// Inicia una simulacion con parámetros de prueba
func SimularTaller(t *models.Taller) {
	N := 4         // número de vehículos por simulación
	NumPlazas := 1 // quito models.MAX_PLAZAS para probar con menos plazas
	NumMecanicos := 1
	Sims := 1 // Para poder ejecutar varias simulaciones seguidas

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
