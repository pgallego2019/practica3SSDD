package menus

import (
	"fmt"
	"taller/models"
	"taller/utils"
)

func MenuVehiculos(t *models.Taller) {
	for {
		fmt.Println("\n--- VEHÍCULOS ---")
		fmt.Println("1. Crear vehículo")
		fmt.Println("2. Mostrar todos los vehículos")
		fmt.Println("3. Modificar vehículo")
		fmt.Println("4. Eliminar vehículo")
		fmt.Println("5. Listar incidencias de un vehículo")
		fmt.Println("6. Asignar vehículo a plaza")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var mat, marca, modelo, fechaE string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			fmt.Print("Marca: ")
			fmt.Scanln(&marca)
			fmt.Print("Modelo: ")
			fmt.Scanln(&modelo)
			fmt.Print("Fecha de entrada: ")
			fmt.Scanln(&fechaE)
			t.NewVehiculo(mat, marca, modelo, fechaE, "", nil)
			fmt.Println("Vehículo creado.")
		case 2:
			if len(t.Vehiculos) == 0 {
				fmt.Println("No hay vehículos registrados.")
				break
			}
			for _, v := range t.Vehiculos {
				utils.PrintVehiculo(v)
				fmt.Println("-----------------------------")
			}
		case 3:
			var mat, marca, modelo, fe, fs string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			fmt.Print("Marca: ")
			fmt.Scanln(&marca)
			fmt.Print("Modelo: ")
			fmt.Scanln(&modelo)
			fmt.Print("Fecha entrada: ")
			fmt.Scanln(&fe)
			fmt.Print("Fecha salida: ")
			fmt.Scanln(&fs)
			fmt.Print("Fecha salida: ")
			fmt.Scanln(&fs)
			if err := t.UpdateVehiculo(mat, marca, modelo, fe, fs); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Vehículo actualizado.")
			}
		case 4:
			var mat string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			t.DeleteVehiculo(mat)
			fmt.Println("Vehículo eliminado.")
		case 5:
			var mat string
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			t.ShowIncidenciasVehiculo(mat)
		case 6:
			var mat string
			var mecID, clienteID int
			fmt.Print("Matrícula del vehículo: ")
			fmt.Scanln(&mat)
			fmt.Print("ID del cliente: ")
			fmt.Scanln(&clienteID)
			fmt.Print("ID del mecánico: ")
			fmt.Scanln(&mecID)

			v := t.GetVehiculo(mat)
			if v == nil {
				fmt.Println("Vehículo no encontrado.")
				break
			}

			err := t.AdmitirCliente(clienteID, v, mecID)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Vehículo asignado correctamente al cliente.")
			}
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}
