package menus

import (
	"fmt"
	"taller/models"
	"taller/utils"
)

func MenuPlazas(t *models.Taller) {
	for {
		fmt.Println("\n--- PLAZAS / ESTADO DEL TALLER ---")
		fmt.Println("1. Ver estado completo del taller")
		fmt.Println("2. Ver plazas ocupadas/libres")
		fmt.Println("0. Volver")

		var op int
		fmt.Print("Seleccione: ")
		fmt.Scanln(&op)

		switch op {
		case 1:
			utils.PrintTaller(t)
		case 2:
			if len(t.Plazas) == 0 {
				fmt.Println("No hay plazas registradas.")
				break
			}
			for i := range t.Plazas {
				utils.PrintPlaza(t.Plazas[i])
				fmt.Println("-----------------------------")
			}
		case 0:
			return
		default:
			fmt.Println("Opción inválida.")
		}
	}
}
