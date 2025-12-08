package menus

import (
	"fmt"
	"taller/models"
	"taller/utils"
)

func MenuMecanicos(t *models.Taller) {
	for {
		fmt.Println("\n--- MECÁNICOS ---")
		fmt.Println("1. Crear mecánico")
		fmt.Println("2. Mostrar todos los mecánicos")
		fmt.Println("3. Modificar mecánico")
		fmt.Println("4. Eliminar mecánico")
		fmt.Println("5. Listar mecánicos activos")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var nombre, esp string
			var exp int
			fmt.Print("Nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Especialidad (mecanica / electrica / carroceria): ")
			fmt.Scanln(&esp)
			fmt.Print("Años de experiencia: ")
			fmt.Scanln(&exp)
			m := t.NewMecanico(nombre, esp)
			fmt.Printf("Mecánico creado (ID: %d)\n", m.ID)
		case 2:
			if len(t.Mecanicos) == 0 {
				fmt.Println("No hay mecánicos registrados.")
				break
			}
			for _, m := range t.Mecanicos {
				utils.PrintMecanico(m)
				fmt.Println("-----------------------------")
			}
		case 3:
			var id, exp int
			var nombre, esp string
			var activo bool
			fmt.Print("ID mecánico: ")
			fmt.Scanln(&id)
			fmt.Print("Nuevo nombre: ")
			fmt.Scanln(&nombre)
			fmt.Print("Nueva especialidad (mecanica / electrica / carroceria): ")
			fmt.Scanln(&esp)
			fmt.Print("Años experiencia: ")
			fmt.Scanln(&exp)
			fmt.Print("Activo (1 sí / 0 no): ")
			var act int
			fmt.Scanln(&act)
			activo = act == 1
			if err := t.UpdateMecanico(id, nombre, activo); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Mecánico actualizado.")
			}
		case 4:
			var id int
			fmt.Print("ID mecánico: ")
			fmt.Scanln(&id)
			if err := t.DeleteMecanico(id); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Mecánico eliminado.")
			}
		case 5:
			t.ShowMecanicosActivos()
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}
