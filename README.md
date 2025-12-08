# Practica 3 -  El Taller del pueblo
#### Sistemas Distribuidos - GIT - URJC 2025
## Introducción

La práctica tiene como objetivo desarrollar un sistema concurrente en Go que modele el funcionamiento del “Taller del Pueblo”. El taller debe gestionar vehículos que llegan con una incidencia a resolver y que pasan por cuatro fases secuenciales:

- Espera de plaza
- Reparación
- Limpieza
- Revisión final

Cada vehículo tiene una incidencia que pertenece a una de tres categorías (A, B o C) que determinan distintos niveles de prioridad y distintos tiempos de fase. Además, existen limitaciones de recursos como número de plazas o número de mecánicos.

El sistema se implementa mediante dos mecanismos de concurrencia distintos del paquete sync de Go: WaitGroup y RWMutex.

El objetivo es comparar ambas implementaciones mediante tests controlados con diferentes distribuciones de vehículos.

## Reorganización del código del proyecto
En la práctica anterior el código estaba distribuido en solo tres archivos principales:

- main.go
- simulacion.go
- simulacion_test.go

Pero la estructura era difícil de mantener a medida que el proyecto crecía y había que añadir/eliminar funcionalidades. Para mejorar la modularidad y la claridad, reestructuré el proyecto en paquetes. Esta organización permite agrupar código con una responsabilidad concreta y facilitar la reutilización de código.

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

- **main.go**: Punto de entrada de la aplicación. Desde aquí se cargan los menús y se coordina la ejecución general del sistema.

- **menus/**: Contiene la lógica relacionada con la interacción con el usuario. Contiene varios submenús para cada una de las estructuras. Permiten delegar acciones a otros paquetes.

- **models/**: Contiene las estructuras de datos principales del proyecto: Cliente, Vehículo, Incidencia, Plaza, Mecánico y Taller. Cada archivo contiene un modelo y métodos asociados. 

- **sim/**: Contiene la lógica relacionada con la simulación del taller y gestiona toda la concurrencia del sistema. Incluye:
	- Implementación de los dos simuladores solicitados en el enunciado (simulador RWMutex y simulador WaitGroup)
	- Interfaz _ISimulador_, ambos simuladores la implementar para poder ejecutar el sistema con exactamente la misma lógica externa, manteniendo únicamente diferencias internas en sincronización.
	- El sistema de fases con sus transiciones.
	- Métricas y los resultados de los tests.
	- Estructuras auxiliares, como colas de prioridad.
	- Funciones auxiliares, como intercalado de vehículos para simular llegadas aleatorias.

- **utils/**: Contiene funciones auxiliares de uso general como impresión formateada o control de pantalla. Permiten simplificar tareas repetitivas.

## Explicación del diseño

### Estructuras de datos
Los modelos que representan elementos del taller han sido simplificados respecto a prácticas anteriores y ahora están todos contenidos en el paquete _models_

- **Vehículo**: Matricula, Marca, Modelo, FechaEntrada, FechaSalida, Incidencia
- **Incidencia**: ID, Tipo, Descripcion, Estado, TiempoFase
- **Plaza**: ID, Ocupada, VehiculoMat
- **Mecánico**: ID, Nombre, Activo
- **Taller**: Clientes, Vehiculos, Mecanicos, Incidencias, Plazas, NextClienteID, NextIncidenciaID, NextMecanicoID.

En el paquete _sim_ definimos estructuras nuevas para la simulación concurrente, mirando archivo por archivo tenemos que:

- _**/sim/colas_prioridad.go**_: La estructura _ColaPrioritaria_ implementa una cola de prioridad con tres listas internas (alta, media, baja) basadas en el tipo de incidencia del vehículo. 

Las listas internas permiten mantener prioridad explícitamente de manera más sencilla. Además, hay un RWMutex que permite lecturas concurrentes y escrituras exclusivas, de esta forma se puede consultar el tamaño de la cola sin bloqueos.

Añadí un canal _notify_ para que los workers pudieran "dormir" mientras esperan una notificación y así evitar el consumo activo de CPU que se da con el polling. Cuando se inserta el primer vehículo, se notifica. También si al extraer un vehículo quedan más, se vuelve a notificar evitando así bloqueos cuando se insertan varios coches seguidos.

Esta estructura enlaza fases y evita condiciones de carrera.

- _**/sim/metricas_sim.go**_: La estructura _Metricas_ almacena información agregada sobre el rendimiento de la simulación. Es para recopilar tiempos totales por vehículo y recuentos por fase. También calculamos el tiempo total de ejecución para comparar entre simuladores.

Al haber muchos workers, hay muchas escrituras concurrentes. Por eso, protegemos con un RWMutex para que pueda haber lecturas concurrentes y lecturas periódicas para mostrar métricas.

Hay otra estructura complementaria _MetricasFase_ que registra valores estadísticos básicos de cada fase (mínimo, máximo, promedio y número de vehículos). Aquí cada worker actualiza las métricas, por lo que tenemos riesgo de condición de carrera, entonces también la protegemos con un Mutex. No usamos RWMutex porque aquí se escribe mucho más de lo que se lee.

- _**/sim/recursos_sim.go**_: La estructura _RecursosSim_ modela los recursos compartidos del taller: las colas por fase y los canales que controlan la capacidad de cada fase. Cada fase del taller tiene su cola independiente y eso permite que trabajen en paralelo sin bloquear a los demás. El diseño es thread safe, entonces no necesitamos Mutex.

### Diagrama de clases

### Diagrama de secuencia

### Diagrama de flujo

### Diagrama de concurrencia

## Implementación del paquete _sim_
El paquete sim constituye el núcleo funcional de la simulación del Taller del Pueblo. Su objetivo es reproducir, de forma controlada y concurrente, el flujo de vehículos a través de las distintas fases del taller (Entrada, Atención por mecánico, Limpieza y Revisión final), aplicando restricciones de recursos (plazas disponibles, mecánicos activos) y midiendo el rendimiento global del sistema mediante métricas detalladas.

Para ello, el paquete implementa:

- Generación estructurada o aleatoria de vehículos, con tiempos base por fase y variación aleatoria controlada.

- Colas con prioridad (A > B > C) para simular categorías de incidencias.

- Dos arquitecturas de simulación alternativas para comparar rendimiento:

	- Simulador con WaitGroup + Colas con notificación (sin polling).

	- Simulador con RWMutex + polling (modelo tradicional).

- Registro detallado de métricas, por fase y por vehículo.

Comparación automática de escenarios y generación de tabla final de resultados.

A continuación se describe exhaustivamente cada uno de los módulos del paquete, sus estructuras de datos, objetivos, decisiones de diseño y flujo de ejecución.

### Módulo vehiculos_sim
Este módulo encapsula todos los elementos relacionados con la creación de vehículos, la definición de categorías del taller, la asignación del tiempo de cada fase y la variación temporal aleatoria exigida por el enunciado.
	#### Definición de fases
Estas fases definen el pipeline por el que pasan los vehículos. El orden es fijo y determinista.
	#### Variación del tiempo de fase
Cada fase tiene un tiempo base (1, 3 o 5 segundos, según categoría) y se calcula una variación en el intervalo [-15%, +15%]. Se varía el tiempo de fase para que exista suficiente diversidad para observar colisiones, colas más largas o diferencias entre simuladores.
	#### Generación de vehículos
		- Totalmente aleatoria: Distribuye vehículos entre categorías A/B/C y asigna ID secuenciales. Mezcla internamente y luego intercala categorías para una distribución más realista.
		-controlada por categorías: dada la distribución de categorías necesarias, se crean los vehículos y se intercalan las categorías.

### Módulo recursos_sim

Introduce una función auxiliar para instanciar workers. Funciona así:

1. Espera notificación de elementos.

2. Toma un token del semáforo (plaza/mecánico).

3. Procesa la fase con variación temporal.

4. Registra métricas.

5. Encola en la siguiente fase o finaliza (wg.Done()).

### 5. Módulo simulador_waitgroup.go — Simulador eficiente con colas notificadas

Este simulador es la implementación recomendada porque: No usa polling, los workers están dormidos hasta que reciben notify, las colas tienen prioridad real.

5.2 Lógica de RunSim

Inicialización de métricas
Se inicializan estructuras incluso si vienen nil (defensive init).

Creación de colas prioritarias: colaEntrada → colaMecanico → colaLimpieza → colaRevision

Creación de semáforos: semPlazas (capacidad = plazas disponibles), semMec, semLimp, semRev

Lanzamiento de pools de workers: NumPlazas workers de entrada, NumMecanicos workers de atención, NumPlazas workers de limpieza, NumPlazas workers de revisión. Cada worker ejecuta el patrón de fases con cola prioritaria y notificación.

Encolado de vehículos: Cada vehículo se encola progresivamente en colaEntrada.

Finalización: wgFinal.Wait(), se establece metricas.Fin, se cierra s.Done para que los workers terminen correctamente.

Este simulador tiene el menor coste en CPU porque no usa loops con sleep.

### 6. Módulo simulador_rwmutex.go — Simulador basado en RWMutex y polling

Esta segunda implementación está construida para comparar técnicas de concurrencia.

6.2 Funcionamiento

Las colas son slices protegidos con sync.RWMutex. 
Los workers hacen: pop() protegido por mutex. Si está vacío → time.Sleep(3ms)(polling → mayor carga CPU). Procesan fase. Reencolan.

6.3 Cierre ordenado

Workers usan case s.Done para salir cuando la simulación termina.

La implementación se basa en la interfaz:
### 8. Módulo simulador_test.go — Comparación automática de ambas arquitecturas

Ejecuta los escenarios:

10 / 10 / 10

20 / 5 / 5

5 / 5 / 20

Para cada uno:

Ejecuta SimuladorWaitGroup y SimuladorRWMutex.

Registra tiempos totales, por fase y por vehículo.

Calcula promedios por categoría.

Inserta la fila en la tabla final.

Esto permite analizar empíricamente:

Cómo afecta la priorización.

Cómo escalan ambos modelos de concurrencia.

Qué diferencias aparecen en saturación de recursos.


"'logica_sim.go' contiene únicamente la lógica de interacción con el usuario y la inicialización de la simulación (parámetros, elección de simulador y generación de vehículos). No introduce nuevas estructuras de datos, por lo que no forma parte del diseño concurrente descrito en esta sección."

"ResultadoSimulacion almacena un resumen ejecutable de cada ejecución: el nombre del escenario, el tipo de simulador empleado, tiempo total, tiempo promedio por vehículo y vehículos procesados por fase."

"El paquete sim proporciona una implementación completa, modular y analíticamente comparable de dos estrategias de simulación.

La estructura del paquete:

Facilita la comprensión del flujo de un taller real.

Cumple los requisitos del enunciado: prioridad, fases, recursos limitados, variación temporal y métricas detalladas.

Permite comparar dos arquitecturas concurrentes en un entorno controlado.

La implementación con WaitGroup + colas notificadas es más eficiente y técnicamente superior; la implementación RWMutex + polling sirve como referencia para analizar el impacto de técnicas de sincronización en sistemas concurrentes."
### Simulador RWMutex
Usa:

sync.RWMutex para proteger estructuras compartidas

RLock para lectura concurrente

Lock para mutaciones

Menor riesgo de bloqueo total

Mejor paralelismo

Es más complejo pero más eficiente.

### Simulador WaitGroup
Usa:

Un WaitGroup para esperar a que todas las goroutines terminen

Control más sencillo

Menos paralelismo real

Mayor bloqueo de fases

Es menos eficiente en competencia alta.

## Test realizados
Los tests se definen en simulacion_test.go y ejecutan los tres escenarios obligatorios del enunciado:

| Escenario  | Vehículos |
| ------------- |:-------------:|
| 1      | 10A, 10B, 10C     |
| 2      | 20A, 5B, 5C     |
| 3      | 5A, 5B, 20C     |


Cada escenario se simula una vez con el simulador WaitGroup y una vez con el simulador RWMutex

Los tests realizan:

Impresión de métricas por fase

Tiempo total

Tiempo medio por vehículo

Tiempo medio por categoría

Comparativa entre simuladores

También se agrega un modo verbose opcional que permite ver: formato del mensaje

### Métricas obtenidas y análisis
A partir de tus simulaciones, las conclusiones generales esperables son:

RWMutex

Mayor paralelismo real

Fewer artificial stalls

Métricas más coherentes entre fases

Consume más CPU pero menor tiempo total de simulación

WaitGroup

Simulación correcta pero más lenta

Más “esperas secuenciales” provocadas por dependencias no optimizadas

Menos competición por mutexes

Simulación más determinista pero menos eficiente

En especial, RWMutex gestiona mucho mejor:

Las colas de prioridad

La liberación parcial de recursos

La simultaneidad entre vehículos
## Conclusiones
La práctica permitió profundizar en:

Diseño de sistemas concurrentes en Go

Diferencias entre sincronización por WaitGroup y RWMutex

Modelado de sistemas reales mediante concurrencia

Técnicas de medición y análisis de rendimiento

Organización de proyectos Go en paquetes

Implementación de colas de prioridad y fases secuenciales

Integración de tests automatizados

Conclusión técnica

RWMutex ofrece un rendimiento superior gracias a su alto grado de paralelismo y bajo nivel de bloqueo, siendo el mejor simulador para cargas grandes o alta competencia de recursos.
WaitGroup, aunque más simple, no se adapta igual de bien al control fino de recursos y genera más congestión.

Conclusión personal

La práctica permitió simular un sistema realista con concurrencia avanzada y comprender más profundamente cómo Go permite resolver problemas complejos mediante goroutines, sincronización y estructuras compartidas.
