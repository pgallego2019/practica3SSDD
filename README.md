# Práctica 3 - El Taller del Pueblo
#### Sistemas Distribuidos - GIT - URJC 2025
## Introducción

La práctica tiene como objetivo desarrollar un taller procesa vehículos que deben atravesar cuatro fases secuenciales: espera de plaza, reparación, limpieza y revisión final.

Cada vehículo tiene una incidencia de tipo A, B o C, lo que determina su prioridad y los tiempos asociados a cada fase. La simulación incorpora además restricciones reales como número de plazas y mecánicos disponibles.

Para ello, se implementan dos simuladores que usan distintos mecanismos de concurrencia del paquete sync de Go: WaitGroup y RWMutex. El objetivo es comparar ambas implementaciones mediante tests controlados con diferentes distribuciones de vehículos.


## Reorganización del código del proyecto

En la práctica anterior el código estaba distribuido en solo tres archivos principales: _main.go_ ,
_simulacion.go_ y _simulacion_test.go_

Pero la estructura era difícil de mantener a medida que el proyecto crecía y había que añadir/eliminar funcionalidades. Para mejorar la modularidad y la claridad, reestructuré el proyecto en paquetes. Esta organización permite agrupar código con una responsabilidad concreta y facilitar la reutilización de código.
```
taller
├── go.mod
├── main.go
├── menus
├── models
├── utils
└── sim
```
Cada paquete incluye:

- **main.go** : Punto de entrada de la aplicación. Desde aquí se cargan los menús y se
coordina la ejecución general del sistema.
- **menus/** : Contiene la lógica relacionada con la interacción con el usuario. Contiene
varios submenús para cada una de las estructuras. Permiten delegar acciones a otros
paquetes.
- **models/** : Contiene las estructuras de datos principales del proyecto: Cliente, Vehículo,
Incidencia, Plaza, Mecánico y Taller. Cada archivo contiene un modelo y métodos
asociados.
- **utils/** : Contiene funciones auxiliares de uso general como impresión formateada o
control de pantalla. Permiten simplificar tareas repetitivas.
- **sim/** : Contiene la lógica relacionada con la simulación del taller y gestiona toda la concurrencia del sistema. Incluye:
    - Implementación de los dos simuladores solicitados en el enunciado (simulador _RWMutex_ y simulador _WaitGroup_ )
    - Interfaz _ISimulador_ , ambos simuladores la implementaron para poder ejecutar el sistema con exactamente la misma lógica externa, manteniendo únicamente diferencias internas en sincronización.
    - El sistema de fases con sus transiciones.
    - Métricas y resultados de los tests.
    - Estructuras auxiliares, como colas de prioridad.
    - Funciones auxiliares, como intercalado de vehículos para simular llegadas aleatorias.


## Explicación del diseño

El diseño de la simulación se basa en la separación explícita entre datos (models) y lógica concurrente (sim). Esta organización permite mantener la arquitectura clara, modular y extensible.

### Modelos

Los modelos que representan elementos del taller han sido simplificados respecto a
prácticas anteriores y ahora están todos contenidos en el paquete _models_
- **Vehículo** : Matrícula, Marca, Modelo, FechaEntrada, FechaSalida, Incidencia
- **Incidencia** : ID, Tipo, Descripción, Estado, TiempoFase
- **Plaza** : ID, Ocupada, VehiculoMat
- **Mecánico** : ID, Nombre, Activo
- **Taller** : Clientes, Vehiculos, Mecanicos, Incidencias, Plazas, NextClienteID, NextIncidenciaID, NextMecanicoID.

### Diagrama de estados

Para representar formalmente el ciclo de vida de un vehículo dentro de la simulación, se utiliza el siguiente **diagrama de estados**. Cada vehículo atraviesa un conjunto de estados bien definidos (llegada, entrada, atención, limpieza, revisión y salida), y este diagrama permite visualizar las transiciones entre ellos.

![Diagrama de estados](https://github.com/pgallego2019/practica3SSDD/blob/main/diagramas/Diagramas_P3_SSDD-Copia%20de%20diagrama%20de%20estados.drawio.png)

### Estructuras auxiliares

El paquete _sim_ define estructuras adicionales necesarias para mantener un flujo
concurrente seguro.
- **_ColaPrioritaria_** : Esta estructura implementa una cola de prioridad con tres listas internas (alta, media, baja) basadas en el tipo de incidencia del vehículo. Las listas internas permiten mantener prioridad explícitamente de manera más sencilla. 
Además, hay un _RWMutex_ que permite lecturas concurrentes (p.ej. consultar tamaño) y escrituras exclusivas (insertar o extraer). Se usa un canal de notificaciones para que los workers pudieran "dormir" mientras esperan una notificación y así evitar el consumo activo de CPU que se da con el polling. Cuando se inserta el primer vehículo, se notifica. También si al extraer un vehículo quedan más, se vuelve a notificar evitando así bloqueos cuando se insertan varios coches seguidos. Esta estructura enlaza fases y evita condiciones de carrera.

- **RecursosSim** : Esta estructura modela los recursos compartidos del taller: las colas por fase y los canales que controlan la capacidad de cada fase. Cada fase del taller tiene su cola independiente y eso permite que trabajen en paralelo sin bloquear a los demás. El diseño es thread safe, entonces no necesitamos Mutex.
- **Métricas**: Se usan dos estructuras:

    * Metricas: Esta estructura almacena información agregada sobre el rendimiento de la simulación. Es para recopilar tiempos totales por vehículo y recuentos por fase. También calculamos el tiempo total de ejecución para comparar entre simuladores. Al haber muchos workers, hay muchas escrituras concurrentes. Por eso, protegemos con un RWMutex para que pueda haber lecturas concurrentes y lecturas periódicas para mostrar métricas.

    * MetricasFase: Es una estructura complementaria que registra valores estadísticos básicos de cada fase (mínimo, máximo, promedio y número de vehículos). Aquí cada worker actualiza las métricas, por lo que tenemos riesgo de condición de carrera, entonces también la protegemos con un Mutex. No usamos RWMutex
porque la frecuencia de escritura es mucho mayor que la de lectura.

- **Workers** : Los workers ejecutan unos pasos que crean un pipeline que permite procesar vehículos simultáneamente.

    1. Espera de notificación o vehículo disponible
    2. Obtención del token
    3. Espera del tiempo de fase con variación aleatoria
    4. Registro de métricas
    5. Inserción en la cola de la siguiente fase
    6. Liberación del token

- **Fases** : Cada fase de la simulación funciona mediante una cola de entrada, un número limitado de workers (controlados por el canal), una cola de salida para la fase siguiente.

Para comprender de forma global el comportamiento de la simulación, se presenta el siguiente **diagrama de flujo** , que recoge las fases por las que pasa cada vehículo desde su llegada al taller hasta su salida. Este diagrama permite visualizar el pipeline completo (entrada, atención mecánica, limpieza y revisión), así como el orden secuencial y la transición entre fases que más tarde gestionarán los workers y las colas concurrentes.

![Diagrama de flujo](https://github.com/pgallego2019/practica3SSDD/blob/main/diagramas/Diagramas_P3_SSDD-diagrama%20de%20flujo.drawio.png)

### Diagrama de clases del paquete sim

A continuación se incluye el diagrama de clases que representa la estructura estática del paquete sim. Este diagrama muestra las entidades principales que intervienen en la simulación (simuladores, colas, recursos y métricas) así como las relaciones entre ellas. Su objetivo es proporcionar una visión global del diseño orientado a objetos que sustenta la arquitectura concurrente de la simulación.

![Diagrama de clases](https://github.com/pgallego2019/practica3SSDD/blob/main/diagramas/Diagramas_P3_SSDD-diagrama%20de%20clases.drawio.png)

### Implementación del paquete sim

El paquete sim constituye el núcleo funcional de la simulación del Taller del Pueblo. Su objetivo es reproducir, de forma controlada y concurrente, el flujo de vehículos a través de las distintas fases del taller, aplicando restricciones de recursos (plazas disponibles, mecánicos activos) y midiendo el rendimiento global del sistema mediante métricas detalladas.

Para ello, el paquete implementa los siguientes módulos:

#### Módulo vehiculos_sim

Este módulo contiene toda la lógica relacionada con la generación, caracterización y temporización de los vehículos que participan en la simulación. Su función es desacoplar la creación y preparación de los vehículos respecto de los simuladores y motores de concurrencia, facilitando su reutilización en múltiples escenarios.

1. Definición de fases
Las fases representan el pipeline fijo por el que atraviesa cada vehículo dentro del taller. El orden es determinista y está pensado para ser utilizado por los workers de los distintos simuladores: Entrada, Atención, Limpieza, Revisión. Cada fase corresponde a un período de trabajo dentro del taller, y forma parte del cálculo del tiempo total por vehículo.

2. Variación del tiempo de fase
Cada fase tiene un tiempo base (1, 3 o 5 segundos, según categoría) y se calcula una variación en el intervalo [-15%, +15%]. Se varía el tiempo de fase para que exista suficiente diversidad para observar colisiones, colas más largas o diferencias entre simuladores.

3. Generación de vehículos, hay dos tipos:
- Totalmente aleatoria: Distribuye vehículos entre categorías A/B/C y asigna ID
secuenciales. Mezcla internamente y luego intercala categorías para una
distribución más realista. Se usa en la simulación normal desde el menú
- Controlada por categorías: dada la distribución de categorías necesarias, se crean los vehículos y se intercalan las categorías. Se usa en los escenarios de test.

#### Módulo recursos_sim

Este módulo contiene la infraestructura que permite simular los recursos limitados del taller (plazas, mecánicos, puestos de limpieza y revisión) y los workers que procesan los vehículos en cada etapa. Es uno de los módulos clave del sistema porque implementa: Colas donde esperan los vehículos, canales (semáforos) que limitan la concurrencia, workers que representan a los distintos operarios del taller y transición encadenada entre fases.

* Funcionamiento del worker
La función LanzarWorker encapsula el comportamiento estándar de un operario en
cualquiera de las fases. Cada worker ejecuta un bucle infinito que sigue el mismo patrón:

1. **Espera de nuevos elementos:** Queda bloqueado en la cola de entrada hasta que esta emite una notificación.

2. **Toma un token del semáforo:** El semáforo representa el recurso limitado (plaza, mecánico, limpiador, revisor). Si no hay tokens disponibles, el worker espera.

3. **Procesa el vehículo con variación temporal:** Aplica la duración base de la incidencia **con la variación aleatoria**. Esto simula tiempos reales fluctuantes.

4. **Libera el recurso:** Devuelve el token al semáforo para que otro worker pueda utilizarlo.

5. **Registra métricas:** El módulo no imprime directamente, sino que delega en la estructura Metricas la gestión del tiempo y del registro por fase.

6. **Encola en la siguiente fase o termina:** Si existe colaOut, se pasa a la fase siguiente y si es la última fase se hace wg.Done() para indicar que el vehículo está acabado.

Este diseño unifica el comportamiento de todos los operarios y evita tener lógica duplicada en cada simulador. Las ventajas de este diseño es que se centraliza el comportamiento del worker, reduce código repetido entre simuladores y controla concurrentemente el accesos a los recursos. Además permite añadir nuevas fases fácilmente y hace que el sistema sea escalable porque bastaría con lanzar más workers usando la misma función

#### Diagrama de secuencia (worker-colas)

Este diagrama describe la comunicación temporal entre un worker del simulador y las distintas estructuras que intervienen en el procesamiento de un vehículo: colas de fase, semáforos de plazas, métricas y actualización de estados. Su finalidad es ilustrar visualmente el ciclo de trabajo de un worker y la coordinación entre los componentes concurrentes.

![Diagrama de secuencia](https://github.com/pgallego2019/practica3SSDD/blob/main/diagramas/Diagramas_P3_SSDD-diagrama%20de%20secuencia.drawio.png)

#### Módulo resultados_sim

Este módulo gestiona el almacenamiento y presentación de los resultados de cada
simulación ejecutada.

Su objetivo es: Registrar métricas globales al concluir cada simulación, comparar múltiples simuladores o escenarios y presentar una tabla final conjunta.

Para poder comparar el rendimiento entre simuladores de forma objetiva, la tabla final
contiene:
- Nombre del escenario
- Simulador utilizado: RWMutex o WaitGroup
- Tiempo total: duración completa de la simulación.
- Tiempo medio por vehículo: promedio de tiempos finales.
- Vehículos procesados por fase (para verificar consistencia).

La tabla permite detectar: qué simulador es más rápido, cuál es más estable, efectos de cuellos de botella, diferencias en rendimiento según el número de recursos...

#### Módulo logica_sim

Este módulo actúa como puente entre la interfaz del usuario (menú principal) y los distintos simuladores del taller. Aquí es donde se gestionan: la entrada interactiva de parámetros, la creación de vehículos, la elección del simulador y la ejecución completa de la simulación. 

El módulo no implementa la lógica del simulador en sí, sino que coordina los elementos ya generados por otros módulos del paquete sim.

#### Módulo simulador_waitgroup

El SimuladorWaitGroup fue diseñado para ser más eficiente y más escalable. Su
característica clave es la introducción de una ColaPrioritaria por fase y un sistema basado en notificaciones, no en polling.

Los workers esperan siempre sobre un select, entonces no consumen CPU cuando la cola está vacía. Extraen el vehículo con prioridad y el uso de ColaPrioritaria lo hace escalable y con mínima contención.

+ Ventajas: No hay polling, priorización correcta de categorías, arquitectura eficiente y escalable, orden lógico más cercano al funcionamiento de un taller real, no necesita mutex externos para las colas.
+ Desventajas: Ligeramente más complejo conceptualmente, el sistema de notificaciones y múltiples colas requiere mayor cuidado en diseño y estructuras auxiliares.

#### Módulo simulador_rwmutex

El simuladorRWMutex utiliza las siguientes primitivas del paquete sync: RWMutex para proteger el acceso a las colas, WaitGroup para el final de la fase de Revisión, time.Sleep como mecanismo de polling cuando la cola está vacía.

Cada fase está formada por una cola implementada como *[]Vehiculo, un RWMutex
asociado y un grupo de workers (limitado por la capacidad de cada fase) que hacen polling activo. Los workers siguen la siguiente secuencia:

1. Intentan extraer un vehículo de la cola usando pop
2. Si la cola está vacía, hacen time.Sleep(3ms) y vuelven a intentarlo
3. Si hay vehículo, consume un token de la fase, ejecuta el “trabajo”, registra las métricas, devuelve el token y pasa el vehículo a la cola siguiente usando push.

El RWMutex permite lecturas concurrentes sobre las colas y evita las condiciones de
carrera sin usar canales intermedios.
- Ventajas: Implementación más sencilla, uso directo de primitivas básicas, fácil de depurar.
- Desventajas: Polling permanente (consumo innecesario de CPU), más probable que haya contención por el uso intensivo de lock/unlock, las colas basadas en slices requieren operaciones con mayor coste.

#### Módulo simulador_interfaz

Ambos simuladores ejecutan exactamente la misma lógica externa (el mismo número de
vehículos, las mismas fases, los mismos recursos y las mismas métricas) pero difieren profundamente en cómo gestionan la concurrencia internamente.

Para permitir esta intercambiabilidad se diseñó la interfaz _ISimulador_ , para poder comparar ambos simuladores en igualdad de condiciones y para que el código de alto nivel del taller no tuviera que conocer los detalles de cada implementación.

Ambos simuladores la implementan, permitiendo que:
- El menú principal ejecute cualquiera de los dos sin cambiar código,
- Los tests puedan comparar de forma aislada el rendimiento
- Se reduzca duplicación de lógica
- Se facilite la extensibilidad (añadir otro simulador sería trivial).

Así, la capa superior no sabe si está ejecutando el simulador con RWMutex o el basado en WaitGroup: solo conoce que ejecutará RunSim() sobre una instancia que cumple la interfaz.

## Test realizados

El objetivo principal es comparar dos implementaciones de sincronización:

1. Simulador WaitGroup + colas prioritarias
2. Simulador basado en RWMutex

Los tests permiten:
- Validar que ambos simuladores completan correctamente el flujo de trabajo: Entrada → Atención → Limpieza → Revisión.
- Analizar los tiempos de ejecución por fase.
- Comparar el tiempo total y tiempo medio por vehículo.
- Estudiar el impacto de distintas distribuciones de categorías de vehículos.
- Detectar posibles cuellos de botella o diferencias de eficiencia entre estrategias de sincronización.

Cada ejecución del test:
- Hay 10 plazas y 2 mecánicos
- Genera 30 vehículos distribuidos en distintas categorías (A-mecánica, B-eléctrica, C-carrocería).
- Ejecuta la simulación con cada uno de los dos motores de concurrencia.
- Registra:
    + Tiempo total de la simulación.
    + Tiempo medio por vehículo.
    + Métricas por fase: mínimo, máximo y promedio.
    + Métricas por categoría.

Además, el test agrega los resultados en una tabla comparativa final, lo que permite estudiar la tendencia global.

### Métricas obtenidas y análisis

**1. Tiempo total de simulación**

Ambos simuladores obtienen tiempos similares, pero el WaitGroup tiende a ser ligeramente más estable entre escenarios, mientras que el RWMutex presenta más variabilidad, especialmente cuando se invierte la carga hacia categorías mayoritarias.

**2. Tiempo promedio por vehículo**

Cuando un tipo de vehículo es mayoritario (p. ej. 20A/5B/5C), la categoría mayoritaria se procesa mucho más rápido (≈ 6 s). Las minoritarias se estancan y muestran tiempos medios de ≈ 12 s. Esto confirma el funcionamiento correcto de la política de colas: Prioridad por orden de llegada, pero afectada por cuellos de botella en mecánicos y plazas.

**3. Métricas**

Las cuatro fases muestran duraciones muy homogéneas:
- Entrada: 0.85 – 1.14 s
- Atención: 0.86 – 1.15 s
- Limpieza: 0.85 – 1.15 s
- Revisión: 0.85 – 1.15 s
Esto valida que los tiempos simulados (aleatorios) se están registrando bien y que la sincronización no introduce retrasos inesperados en fases concretas.

### Comparación entre WaitGroup y RWMutex

**1. WaitGroup**

Ventajas: Distribución de tiempos más regular, mejor rendimiento en escenarios
balanceados, simplicidad de sincronización al esperar fases completas.

Desventajas: Puede ser menos eficiente en situaciones de alta contención si existen tareas muy desiguales.

**2. RWMutex**

Ventajas: Ligera ventaja cuando la carga está concentrada en pocas categorías
(20A/5B/5C), más flexible para permitir lectura concurrente.

Desventajas: Mayor variabilidad entre iteraciones, penalización en escenarios donde
muchas goroutines compiten por el lock.
En general, WaitGroup muestra mayor estabilidad, RWMutex mayor variabilidad.

### Escenarios

Se evaluaron tres escenarios, cada uno con 30 vehículos, pero distribuido de forma
diferente para estudiar cómo afecta la carga del sistema:

#### Escenario 1: 10A / 10B / 10C

Carga equilibrada.
Resultados:
- Los tiempos promedio por categoría son prácticamente idénticos.
- Tiempo total ≈ 18 s para ambos simuladores.
- Mínimas diferencias entre WaitGroup (11.98 s) y RWMutex (12.06 s).
Es el escenario perfecto para medir eficiencia base: ambos simuladores muestran
rendimiento similar.

#### Escenario 2: 20A / 5B / 5C

Carga concentrada en A (mecánica).
Resultados:
- Los vehículos A tardan solo ≈ 5.9 s.
- B y C tardan ≈ 11.7–11.9 s.
- El tiempo total del simulador RWMutex es el menor registrado (17.65 s).
Es el único escenario donde RWMutex es más rápido que WaitGroup.

#### Escenario 3: 5A / 5B / 20C

El escenario inverso al anterior.
Resultados:
- Los vehículos de la categoría mayoritaria (C) se procesan rápido: ≈ 5.9 s.
- Los minoritarios vuelven a duplicar su tiempo.
Aquí WaitGroup supera a RWMutex (17.89 s vs 18.38 s).

### Conclusiones

Ambas estrategias de sincronización funcionan correctamente y completan el flujo sin errores ni deadlocks.

El simulador WaitGroup es más estable y ligeramente más eficiente en escenarios
equilibrados o dominados por una categoría distinta a la inicial del taller.

El simulador RWMutex muestra mejor rendimiento cuando la mayor parte del tráfico se
concentra en una categoría específica, aunque su tiempo total presenta mayor variabilidad.

Los tiempos por fase son prácticamente idénticos entre ambos simuladores, lo que
demuestra que la diferencia de rendimiento proviene estrictamente de la estrategia de sincronización, no del procesamiento de las fases.

El test permite validar que el sistema de métricas funciona correctamente, ya que recoge mínimos, máximos, promedios y tiempos por categoría sin inconsistencias. El comportamiento del taller se ajusta a lo esperado en términos de concurrencia:
- La categoría mayoritaria monopoliza recursos y reduce sus tiempos.
- Las minoritarias sufren mayor espera acumulada.
El rendimiento total del sistema ronda los 18 segundos en cualquier escenario, lo que confirma que la carga total (30 vehículos × 4 fases × tiempos aleatorios ≈ 1 s) está siendo procesada de forma concurrente y eficiente.



