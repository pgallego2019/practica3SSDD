package menus

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"taller/models"
	"taller/utils"
)

func MenuIncidencias(t *models.Taller) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- INCIDENCIAS ---")
		fmt.Println("1. Crear incidencia")
		fmt.Println("2. Mostrar todas las incidencias")
		fmt.Println("3. Modificar incidencia")
		fmt.Println("4. Eliminar incidencia")
		fmt.Println("5. Cambiar estado de incidencia")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			var mat, tipo, pri, desc string
			var mecID int
			fmt.Print("Matrícula: ")
			fmt.Scanln(&mat)
			fmt.Print("Tipo: ")
			fmt.Scanln(&tipo)
			fmt.Print("Prioridad: ")
			fmt.Scanln(&pri)
			fmt.Print("Descripción: ")
			desc, _ = reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			fmt.Print("ID Mecánico: ")
			fmt.Scanln(&mecID)
			mec := t.GetMecanico(mecID)
			if mec == nil {
				fmt.Println("Mecánico no encontrado.")
				break
			}
			inc, err := t.NewIncidencia(mat, tipo, pri, desc)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Incidencia creada (ID: %d)\n", inc.ID)
			}
		case 2:
			if len(t.Incidencias) == 0 {
				fmt.Println("No hay incidencias registradas.")
				break
			}
			for _, inc := range t.Incidencias {
				utils.PrintIncidencia(inc)
				fmt.Println("-----------------------------")
			}
		case 3:
			var id, estado int
			var tipo, pri, desc string
			fmt.Print("ID incidencia: ")
			fmt.Scanln(&id)
			fmt.Print("Tipo: ")
			fmt.Scanln(&tipo)
			fmt.Print("Prioridad: ")
			fmt.Scanln(&pri)
			fmt.Print("Descripción: ")
			desc, _ = reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			fmt.Print("Estado (0 Abierta, 1 En proceso, 2 Cerrada): ")
			fmt.Scanln(&estado)
			if err := t.UpdateIncidencia(id, tipo, pri, desc, estado); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Incidencia actualizada.")
			}
		case 4:
			var id int
			fmt.Print("ID incidencia: ")
			fmt.Scanln(&id)
			t.DeleteIncidencia(id)
			fmt.Println("Incidencia eliminada.")
		case 5:
			var id, estado int
			fmt.Print("ID incidencia: ")
			fmt.Scanln(&id)
			fmt.Print("Nuevo estado (0 Abierta, 1 En proceso, 2 Cerrada): ")
			fmt.Scanln(&estado)
			inc := t.GetIncidencia(id)
			if inc == nil {
				fmt.Println("Incidencia no encontrada.")
				break
			}
			inc.Estado = estado
			fmt.Println("Estado actualizado.")
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}
