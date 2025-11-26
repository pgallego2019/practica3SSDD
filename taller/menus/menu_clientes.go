package menus

import (
	"fmt"
	"taller/models"
	"taller/utils"
)

func MenuClientes(t *models.Taller) {
	for {
		fmt.Println("\n--- CLIENTES ---")
		fmt.Println("1. Crear cliente")
		fmt.Println("2. Mostrar todos los clientes")
		fmt.Println("3. Modificar cliente")
		fmt.Println("4. Eliminar cliente")
		fmt.Println("5. Listar vehículos de un cliente")
		fmt.Println("0. Volver al menú principal")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var nombre, email string
			var tel int
			fmt.Print("Nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Teléfono: ")
			fmt.Scanln(&tel)
			fmt.Print("Email: ")
			fmt.Scanln(&email)
			t.NewCliente(nombre, tel, email, nil)
			fmt.Println("Cliente creado correctamente.")
		case 2:
			if len(t.Clientes) == 0 {
				fmt.Println("No hay clientes registrados.")
				break
			}
			for _, c := range t.Clientes {
				utils.PrintCliente(c)
				fmt.Println("-----------------------------")
			}
		case 3:
			var id, tel int
			var nombre, email string
			fmt.Print("ID de cliente: ")
			fmt.Scanln(&id)
			fmt.Print("Nuevo nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Nuevo teléfono: ")
			fmt.Scanln(&tel)
			fmt.Print("Nuevo email: ")
			fmt.Scanln(&email)
			if err := t.UpdateCliente(id, nombre, tel, email); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Cliente actualizado.")
			}
		case 4:
			var id int
			fmt.Print("ID de cliente: ")
			fmt.Scanln(&id)
			t.DeleteCliente(id)
			fmt.Println("Cliente eliminado.")
		case 5:
			var id int
			fmt.Print("ID de cliente: ")
			fmt.Scanln(&id)
			t.ShowVehiculosCliente(id)
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}
