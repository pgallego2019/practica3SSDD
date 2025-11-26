package models

type Cliente struct {
	ID        int
	Nombre    string
	Telefono  int
	Email     string
	Vehiculos []*Vehiculo
}
