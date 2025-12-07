package sim

import "taller/models"

type ISimulador interface {
	SetVerbose(v bool)

	RunSim(
		vehiculos []*models.Vehiculo,
		Sims int,
		Nvehiculos int,
		NumPlazas int,
		NumMecanicos int,
		maxEsperas map[Fase]int,
		metricas *Metricas,
		tiempoPorVehiculo *TiempoVehiculo,
		aux map[Fase]*MetricasFase,
	)
}
