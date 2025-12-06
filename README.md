# Practica 3 -  El Taller del pueblo
#### Sistemas Distribuidos - GIT - URJC 2025
## Introducción

## Reorganización del código del proyecto
Comenzamos la práctica reorganizando el código existente. En la práctica anterior el código estaba distribuido en solo tres archivos principales (main.go, simulacion.go y simulacion_test.go), pero la estructura era difícil de mantener a medida que el proyecto crecía y había que añadir/eliminar funcionalidades.

Para mejorar la modularidad y la claridad, reestructuré el proyecto en paquetes. Esta organización permite agrupar código con una responsabilidad concreta y facilitar la reutilización de código.
```
taller
    ├── go.mod
    ├── main.go
    ├── menus
    ├── models
    ├── sim
    └── utils
```
Cada paquete incluye:
- go.mod: Archivo de configuración del módulo Go. Define el nombre del módulo y destiona las dependencias externas del proyecto.
- main.go: Punto de entrada de la aplicación. Desde aquí se cargan los menús y se coordina la ejecución general del sistema.
- menus/: Contiene la lógica relacionada con la interacción con el usuario. Contiene varios submenús para cada una de las estructuras. Permiten delegar acciones a otros paquetes.
- models/: Contiene las estructuras de datos principales del proyecto: Cliente, Vehículo, Incidencia, Plaza, Mecánico y Taller. Cada archivo contiene un modelo y métodos asociados. 
- sim/: Contiene la lógica relacionada con la simulación del taller. Incluye el modelo de simulador general, la lógica que gestiona las fases y las transiciones del funcionamiento y estructuras auxiliares (como colas de prioridad).
- utils/: Contiene funciones auxiliares de uso general como impresión formateada o control de pantalla. Permiten simplificar tareas repetitivas.
## Explicación del diseño

### Estructuras de datos

### Diagrama de clases

### Diagrama de secuencia

### Diagrama de flujo

### Diagrama de concurrencia

## Implementación del paquete _sim_

## Test realizados
	
### Métricas obtenidas y análisis

## Conclusiones
