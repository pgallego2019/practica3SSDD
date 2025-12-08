package utils

import (
	"fmt"
	"taller/models"
)

func PrintCliente(c *models.Cliente) {
	fmt.Printf("Cliente %d - %s (%s)\n", c.ID, c.Nombre, c.Email)
	if len(c.Vehiculos) == 0 {
		fmt.Println("  Sin vehículos registrados")
		return
	}
	fmt.Println("  Vehículos:")
	for _, v := range c.Vehiculos {
		fmt.Printf("   - %s (%s %s)\n", v.Matricula, v.Marca, v.Modelo)
	}
}

func estadoToString(est int) string {
	switch est {
	case 0:
		return "Abierta"
	case 1:
		return "En proceso"
	case 2:
		return "Cerrada"
	default:
		return "Desconocido"
	}
}

func PrintVehiculo(v *models.Vehiculo) {
	fmt.Printf("Vehículo %s: %s %s\n", v.Matricula, v.Marca, v.Modelo)

	if v.Incidencia == nil {
		fmt.Println("  Sin incidencia registrada")
		return
	}

	inc := v.Incidencia
	fmt.Println("  Incidencia:")
	fmt.Printf("   - [%s] %s\n",
		estadoToString(inc.Estado),
		inc.Tipo,
	)
}

func PrintIncidencia(i *models.Incidencia) {
	if i == nil {
		fmt.Println("Incidencia no encontrada.")
		return
	}

	fmt.Printf("ID: %d\n", i.ID)
	fmt.Printf("Tipo: %s\n", i.Tipo)
	fmt.Printf("Descripción: %s\n", i.Descripcion)

	estadoStr := ""
	switch i.Estado {
	case 0:
		estadoStr = "Abierta"
	case 1:
		estadoStr = "En proceso"
	case 2:
		estadoStr = "Cerrada"
	default:
		estadoStr = "Desconocido"
	}
	fmt.Printf("Estado: %s\n", estadoStr)
}

func PrintMecanico(m *models.Mecanico) {
	if m == nil {
		fmt.Println("Mecánico no encontrado.")
		return
	}

	fmt.Printf("ID: %d\n", m.ID)
	fmt.Printf("Nombre: %s\n", m.Nombre)
	fmt.Printf("Activo: %t\n", m.Activo)
}

func PrintPlaza(p *models.Plaza) {
	if p == nil {
		fmt.Println("Plaza no encontrada.")
		return
	}

	fmt.Printf("ID: %d\n", p.ID)
	fmt.Printf("Ocupada: %t\n", p.Ocupada)
	if p.Ocupada {
		fmt.Printf("Vehículo matricula: %s\n", p.VehiculoMat)
	}
}

func PrintTaller(t *models.Taller) {
	if t == nil {
		fmt.Println("Taller no encontrado.")
		return
	}

	fmt.Println("=== Taller ===")

	// ---- CLIENTES ----
	fmt.Printf("Clientes (%d):\n", len(t.Clientes))
	if len(t.Clientes) > 0 {
		for i, c := range t.Clientes {
			fmt.Printf("  Cliente %d:\n", i+1)
			PrintCliente(c)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguno)")
	}

	// ---- VEHÍCULOS ----
	fmt.Printf("\nVehículos (%d):\n", len(t.Vehiculos))
	if len(t.Vehiculos) > 0 {
		for i, v := range t.Vehiculos {
			fmt.Printf("  Vehículo %d:\n", i+1)
			PrintVehiculo(v)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguno)")
	}

	// ---- MECÁNICOS ----
	fmt.Printf("\nMecánicos (%d):\n", len(t.Mecanicos))
	if len(t.Mecanicos) > 0 {
		for i, m := range t.Mecanicos {
			fmt.Printf("  Mecánico %d:\n", i+1)
			PrintMecanico(m)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguno)")
	}

	// ---- INCIDENCIAS ----
	fmt.Printf("\nIncidencias (%d):\n", len(t.Incidencias))
	if len(t.Incidencias) > 0 {
		for i, inc := range t.Incidencias {
			fmt.Printf("  Incidencia %d:\n", i+1)
			PrintIncidencia(inc)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguna)")
	}

	// ---- PLAZAS ----
	fmt.Printf("\nPlazas (%d):\n", len(t.Plazas))
	if len(t.Plazas) > 0 {
		for i, p := range t.Plazas {
			fmt.Printf("  Plaza %d:\n", i+1)
			PrintPlaza(p)
			fmt.Println()
		}
	} else {
		fmt.Println("  (ninguna)")
	}

	fmt.Printf("\nPróximo ID cliente: %d\n", t.NextClienteID)
	fmt.Printf("Próximo ID incidencia: %d\n", t.NextIncidenciaID)
	fmt.Printf("Próximo ID mecánico: %d\n", t.NextMecanicoID)
}
