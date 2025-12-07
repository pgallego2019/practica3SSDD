package sim

import (
	"fmt"
	"math/rand"
	"sync"
	"taller/models"
	"time"
)

const variacionMax = 0.15 // Mejor entre 10-20%

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
		models.Mecanica:   {},
		models.Electrica:  {},
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

// Función base para intercalar vehículos de distintas categorías
func intercalarCategorias(categorias [][]*models.Vehiculo) []*models.Vehiculo {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	indices := make([]int, len(categorias))
	result := []*models.Vehiculo{}
	total := 0
	for _, cat := range categorias {
		total += len(cat)
	}

	for len(result) < total {
		options := []int{}
		for i, cat := range categorias {
			if indices[i] < len(cat) {
				options = append(options, i)
			}
		}

		ch := options[r.Intn(len(options))]
		result = append(result, categorias[ch][indices[ch]])
		indices[ch]++
	}

	return result
}

// buildVehiculos crea vehículos de una categoría específica
func buildVehiculos(tipo models.Especialidad, prefijo string, cantidad int, tiempoFase int, startID int) ([]*models.Vehiculo, int) {
	vehiculos := make([]*models.Vehiculo, cantidad)
	for i := 0; i < cantidad; i++ {
		vehiculos[i] = &models.Vehiculo{
			Matricula:    fmt.Sprintf("%s-%03d", prefijo, startID),
			Marca:        "MarcaX",
			Modelo:       "ModeloY",
			FechaEntrada: time.Now().Format("2006-01-02 15:04:05"),
			Incidencia: &models.Incidencia{
				ID:         startID,
				Tipo:       tipo,
				TiempoFase: tiempoFase,
			},
		}
		startID++
	}
	return vehiculos, startID
}

func GenerarVehiculosAleatorios(N int) []*models.Vehiculo {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	catA, catB, catC := []*models.Vehiculo{}, []*models.Vehiculo{}, []*models.Vehiculo{}
	startID := 1

	for i := 0; i < N; i++ {
		switch r.Intn(3) {
		case 0:
			v, newID := buildVehiculos(models.Mecanica, "A", 1, 5, startID)
			catA = append(catA, v...)
			startID = newID
		case 1:
			v, newID := buildVehiculos(models.Electrica, "B", 1, 3, startID)
			catB = append(catB, v...)
			startID = newID
		case 2:
			v, newID := buildVehiculos(models.Carroceria, "C", 1, 1, startID)
			catC = append(catC, v...)
			startID = newID
		}
	}

	// Mezclar internamente cada categoría
	r.Shuffle(len(catA), func(i, j int) { catA[i], catA[j] = catA[j], catA[i] })
	r.Shuffle(len(catB), func(i, j int) { catB[i], catB[j] = catB[j], catB[i] })
	r.Shuffle(len(catC), func(i, j int) { catC[i], catC[j] = catC[j], catC[i] })

	// Intercalar usando la función base
	return intercalarCategorias([][]*models.Vehiculo{catA, catB, catC})
}

func GenerarVehiculosPorCategorias(numA, numB, numC int) []*models.Vehiculo {
	startID := 1
	catA, _ := buildVehiculos(models.Mecanica, "A", numA, 1, startID)
	catB, _ := buildVehiculos(models.Electrica, "B", numB, 1, startID)
	catC, _ := buildVehiculos(models.Carroceria, "C", numC, 1, startID)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Mezclar internamente cada categoría
	r.Shuffle(len(catA), func(i, j int) { catA[i], catA[j] = catA[j], catA[i] })
	r.Shuffle(len(catB), func(i, j int) { catB[i], catB[j] = catB[j], catB[i] })
	r.Shuffle(len(catC), func(i, j int) { catC[i], catC[j] = catC[j], catC[i] })

	// Intercalar usando la función base
	return intercalarCategorias([][]*models.Vehiculo{catA, catB, catC})
}

// Cuenta cuántos vehículos hay de cada tipo
func ContarCategorias(vehiculos []*models.Vehiculo) (mecanica, electrica, carroceria int) {
	for _, v := range vehiculos {
		switch v.Incidencia.Tipo {
		case models.Mecanica:
			mecanica++
		case models.Electrica:
			electrica++
		case models.Carroceria:
			carroceria++
		}
	}
	return
}

func ImprimirResumenCategorias(vehiculos []*models.Vehiculo) {
	m, e, c := ContarCategorias(vehiculos)
	total := len(vehiculos)

	fmt.Printf("%d vehículos generados: %d mecánica, %d eléctrica, %d carrocería\n",
		total, m, e, c)
}
