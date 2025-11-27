package main

import (
	"fmt"
	"taller/menus"
	"taller/models"
	"taller/sim"
	"taller/utils"
)

const MAX_PLAZAS = 10 // número máximo total de plazas en el taller

func main() {
	t := &models.Taller{}

	for {
		fmt.Println("\n===== GESTIÓN DE TALLER =====")
		fmt.Println("1. Clientes")
		fmt.Println("2. Vehículos")
		fmt.Println("3. Incidencias")
		fmt.Println("4. Mecánicos")
		fmt.Println("5. Plazas y estado del taller")
		fmt.Println("6. Limpiar pantalla")
		fmt.Println("7. Simulación del Taller del Pueblo")
		fmt.Println("0. Salir")
		fmt.Print("Seleccione una opción: ")

		var op int
		fmt.Scanln(&op)

		switch op {
		case 1:
			menus.MenuClientes(t)
		case 2:
			menus.MenuVehiculos(t)
		case 3:
			menus.MenuIncidencias(t)
		case 4:
			menus.MenuMecanicos(t)
		case 5:
			menus.MenuPlazas(t)
		case 6:
			utils.ClearScreen()
		case 7:
			sim.SimularTaller(t)
		case 0:
			fmt.Println("Saliendo del sistema...")
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}
