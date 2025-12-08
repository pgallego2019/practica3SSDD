package models

type Especialidad string

const (
	Mecanica   Especialidad = "mecanica"
	Electrica  Especialidad = "electrica"
	Carroceria Especialidad = "carroceria"
)

type Incidencia struct {
	ID          int
	Tipo        Especialidad
	Descripcion string
	Estado      int // 0 abierta, 1 en proceso, 2 cerrada
	TiempoFase  int // 5 mecáninca, 3 eléctrica, 1 carrocería
}
