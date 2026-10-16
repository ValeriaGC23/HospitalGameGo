# CLAUDE.md — Hospital de los Costeños con Narcolepsia

Proyecto académico individual de Programación Orientada a Objetos (Universidad EIA).
Sistema de gestión hospitalaria en **Go** que registra pacientes con narcolepsia, sus episodios de sueño súbito, la asignación de camas y el personal que los atiende.

- **Entrega (link del repositorio):** 19/10/2026, 23:59
- **Sustentación oral individual:** 20/10/2026 (obligatoria)
- **Peso:** 15 % de la nota del curso

> Las decisiones marcadas con `[POR CONFIRMAR]` están pendientes de definición y se actualizan etapa por etapa. No asumas su valor: si un paso depende de una de ellas, pregunta.

---

## 1. Alcance del MVP (producto mínimo viable)

### Dentro del MVP (obligatorio)
1. Cinco structs: `Hospital`, `Patient`, `Doctor`, `Room`, `EpisodeRecord`, con campos no exportados y métodos de acceso.
2. Composición: struct `Person` (id, name, age) embebido en `Patient` y `Doctor`, sin campos duplicados.
3. Polimorfismo: interfaz `Attender` (`ID()`, `Name()`, `IsAvailable()`, `Attend(p, location)`) implementada por `Doctor` y `Orderly`; el hospital guarda el personal en `[]Attender` y despacha sin conocer el tipo concreto.
4. Estados como tipos con constantes (`iota`) y método `String()`: `PatientState`, `NarcolepsyLevel`, `RoomState`.
5. Las cuatro consultas, todas devolviendo datos (nunca imprimiendo):
   - `(h *Hospital) PatientsInHallway() []*Patient`
   - `(d *Doctor) MyEpisodes() []EpisodeRecord` (el doctor no recibe el hospital)
   - `(h *Hospital) AssignRoom(p *Patient) (*Room, error)`
   - `(h *Hospital) SevereReport() map[*Patient]int`
6. `main.go` ejecuta el escenario de la sección 5 de este archivo e imprime el resultado de las cuatro consultas.
7. Mínimo tres pruebas en `hospital/hospital_test.go`.
8. `README.md` y `AI_USAGE.md` completos y veraces.

### Bonus (decidido: se implementan ambos, pero solo después de cerrar y etiquetar el núcleo)
- **Simulación concurrente (+5 %):** cada paciente como goroutine que se duerme en intervalos aleatorios, estado del hospital protegido con `sync.Mutex`, limpio bajo `go run -race .`. Una versión con data races no suma.
- **Interfaz gráfica (+10 %):** Fyne, Wails o API REST + frontend HTML/JS (tecnología `[POR CONFIRMAR]` en la etapa 11). Se califica la separación lógica/vista: el paquete `hospital` **no cambia** al agregar la interfaz y no contiene código de visualización.

Reglas del bonus:
- El núcleo se cierra, se prueba y se etiqueta en Git (`git tag mvp-core`) antes de empezar cualquier bonus. Cada bonus se desarrolla en su propia rama.
- La sincronización (`sync.Mutex`) pertenece al paquete `hospital` y se agrega en la etapa 10, **antes** de la GUI, para que la GUI no obligue a modificar `hospital` después.
- `go run .` debe seguir ejecutando el escenario obligatorio de la sección 5 sin cambios. La simulación concurrente y la GUI se invocan por separado (mecanismo `[POR CONFIRMAR]`: flag, subcomando o directorio `cmd/`).
- Si un bonus compromete la estabilidad del núcleo, se descarta; el núcleo tiene prioridad sobre cualquier bonus.

No agregues funcionalidades fuera de lo listado sin que el usuario lo pida.

---

## 2. Reglas técnicas (no negociables)

- Lenguaje: **Go 1.21 o superior**. Solo librería estándar. Terceros únicamente para la GUI opcional.
- Sin clases ni herencia: usar structs, métodos con receptor, composición (embedding) e interfaces.
- **Encapsulación por paquete:** todo campo que no deba tocarse desde fuera va en minúscula (no exportado) y se accede mediante métodos. Prohibido un struct con todos los campos exportados y sin métodos.
- **Errores:** todo método que pueda fallar devuelve `error` como último valor. Prohibido `panic` para situaciones previsibles (por ejemplo, "no hay cama libre"). Usar errores centinela (`var ErrNoRoomAvailable = errors.New(...)`) y envolver con `fmt.Errorf("...: %w", err)` cuando se agregue contexto.
- **Estados:** nunca strings sueltos; siempre tipos propios con constantes.
- **Separación de responsabilidades:** la lógica de negocio vive en el paquete `hospital`. `main.go` solo construye el escenario e imprime. El paquete `hospital` no contiene `fmt.Print*` ni código de presentación.
- Las consultas devuelven datos; imprimir ocurre solo en `main`.
- `gofmt` aplicado y `go vet ./...` sin advertencias antes de cada commit.
- Identificadores consistentes en todo el proyecto (no mezclar inglés y español).

### Entorno

- Go 1.21 o superior ya instalado (versión exacta: se verifica con `go version` en la etapa 0).
- Módulo: `JuegoHospitalGo` (`go mod init JuegoHospitalGo`). Los imports del paquete del modelo son `JuegoHospitalGo/hospital`. El nombre de la carpeta raíz no está atado al nombre del módulo.
- Editor: VS Code con la extensión oficial de Go (`golang.go`), formato al guardar con `gofmt` y `go vet` habilitado.

### Estructura de directorios

```
costenos-narcolepsia/
  go.mod
  main.go              // escenario + impresión únicamente
  README.md
  AI_USAGE.md
  CLAUDE.md
  hospital/
    hospital.go
    patient.go
    doctor.go
    room.go
    episode.go
    hospital_test.go
```

Se pueden agregar archivos dentro de `hospital/` (por ejemplo `person.go`, `orderly.go`, `attender.go`, `types.go`) si mejoran la cohesión. No se agregan paquetes nuevos sin consultar.

---

## 3. Convenciones de código

- **Idioma de identificadores:** inglés (`Patient`, `AssignRoom`, `ErrNoRoomAvailable`), consistente con las firmas del enunciado.
- **Idioma de comentarios:** inglés (GoDoc, comentarios de línea y mensajes de error). Un solo idioma en todo el código fuente. Documentación del repositorio (`README.md`, `AI_USAGE.md`): también en inglés.
- **Estilo de comentarios (técnico y profesional, no didáctico):**
  - Todo identificador exportado lleva comentario GoDoc (la documentación estándar de Go) que **empieza con el nombre del identificador**.
  - Cada archivo y cada paquete tienen un comentario de cabecera que describe su responsabilidad.
  - Comentar el *porqué* (invariantes, decisiones de diseño, precondiciones, efectos secundarios, errores devueltos), no el *qué* obvio de la línea.
  - Documentar explícitamente los errores que puede devolver cada método y bajo qué condición.
  - No usar tono explicativo hacia el lector ni frases del tipo "aquí hacemos...".
- Receptores por puntero (`*T`) en métodos que modifican estado o en structs con estado mutable; mantener el mismo tipo de receptor en todos los métodos de un mismo tipo.
- Nombres de receptor cortos y consistentes (`h` para Hospital, `p` para Patient, `d` para Doctor, `r` para Room).
- Funciones cortas, una responsabilidad por función. Evitar variables globales mutables.
- Los métodos que devuelven colecciones internas devuelven copias (slice o mapa nuevo) para no filtrar estado mutable.

---

## 4. Flujo de trabajo con Claude Code

1. **Una etapa a la vez.** Al terminar cada etapa: ejecutar `gofmt -l .`, `go vet ./...` y `go test ./...`, y reportar el resultado.
2. **Preguntar antes de decidir** cualquier detalle de diseño no definido en este archivo (firmas, nombres, estructura interna, decisiones con más de una opción razonable). No inventar valores.
3. **Mantener el código defendible.** El usuario debe poder explicar cada línea en la sustentación. Preferir soluciones simples e idiomáticas sobre soluciones ingeniosas. Si algo no es obvio, explicar la decisión en un comentario y resumirla al usuario.
4. **No avanzar a la siguiente etapa** sin confirmación del usuario.
5. **No agregar dependencias externas** ni cambiar la estructura de carpetas sin autorización.
6. **Registrar para `AI_USAGE.md`:** herramientas usadas, prompts decisivos y cada caso en que el código generado estuvo mal o no fue idiomático, con su corrección. Anotarlo en cuanto ocurra.
7. **Commits frecuentes y pequeños** (el repositorio debe tener historial completo; un único commit el día de entrega será cuestionado en la defensa).
   - Convención de commits: **Conventional Commits en inglés**, formato `<type>(<scope>): <description>` en modo imperativo y minúscula, con tipos `feat`, `fix`, `test`, `docs`, `refactor`, `chore`, `style`. Ejemplo: `feat(room): add Occupy method with capacity validation`.
   - Un commit por unidad lógica; no mezclar cambios no relacionados. Antes de cada commit: `gofmt -l .`, `go vet ./...`, `go test ./...`.

---

## 5. Escenario de demostración (orden obligatorio en `main.go`)

1. Crear el hospital; contratar **2 doctores y 1 camillero** (`Orderly`); registrar **3 habitaciones de capacidad 1**; admitir **5 pacientes** (al menos 2 con nivel `Severe`).
2. **Cuatro pacientes** sufren un ataque de sueño en ubicaciones distintas. Tres obtienen habitación; el cuarto falla, permanece en `AsleepInHallway` y el error se imprime como mensaje legible.
3. Un paciente despierta y libera su habitación; el paciente que quedó en el pasillo se asigna a esa habitación.
4. Imprimir el resultado de las cuatro consultas, con encabezados claramente identificables (`== Patients asleep in hallway ==`, etc.).

El tercer paso reutiliza el flujo normal de asignación; no se hardcodea el resultado.

---

## 6. Pruebas (`hospital/hospital_test.go`)

Mínimo tres, idealmente más:
- Asignación de habitación cuando hay una libre: habitación pasa a `Occupied`, paciente pasa a `AsleepInBed`.
- Asignación cuando no hay habitaciones libres: devuelve error, el paciente sigue en `AsleepInHallway`, no se crea ninguna habitación.
- Al menos una de las cuatro consultas (preferible las cuatro).

Usar pruebas dirigidas por tabla (*table-driven tests*) cuando haya varios casos. Las pruebas viven en el paquete `hospital` y validan comportamiento observable a través de la API pública, o con `package hospital` interno solo si es indispensable.

---

## 7. Etapas del proyecto

Marcar con `[x]` al completar y con confirmación del usuario.

- [ ] **Etapa 0** — Entorno: verificar Go ≥ 1.21, `go mod init`, repositorio Git, estructura de carpetas, primer commit.
- [ ] **Etapa 1** — Tipos con constantes (`PatientState`, `NarcolepsyLevel`, `RoomState`) y `String()`.
- [ ] **Etapa 2** — `Person` y `Patient`.
- [ ] **Etapa 3** — `Room`.
- [ ] **Etapa 4** — `EpisodeRecord`.
- [ ] **Etapa 5** — Interfaz `Attender`, `Doctor` y `Orderly`.
- [ ] **Etapa 6** — `Hospital` y las cuatro consultas.
- [ ] **Etapa 7** — Pruebas.
- [ ] **Etapa 8** — `main.go` con el escenario.
- [ ] **Etapa 9** — `gofmt`, `go vet`, `README.md`, `AI_USAGE.md`, verificación desde clon limpio.
- [ ] **Etapa 10** — Bonus de concurrencia: `sync.Mutex` en `hospital`, simulación con goroutines, verificación con `go run -race .` (tras `git tag mvp-core`).
- [ ] **Etapa 11** — Bonus de interfaz gráfica: tecnología a elegir; el paquete `hospital` no se modifica.

### Cronograma tentativo
- Viernes 16/10: etapas 0 a 5
- Sábado 17/10: etapas 6 a 8
- Domingo 18/10: etapa 9 y bonus (10 y 11, en ese orden)
- Lunes 19/10: revisión final y entrega antes de las 23:59

Si el cronograma se retrasa, se recorta primero la etapa 11, luego la 10. El núcleo (etapas 0 a 9) no se negocia.

---

## 8. Decisiones de diseño abiertas (resolver en la etapa indicada)

| Etapa | Decisión | Estado |
|---|---|---|
| 0 | Nombre del módulo en `go.mod` | Resuelto: `JuegoHospitalGo` |
| 2 | `Person` embebido por valor o por puntero | Resuelto: **por valor** (`Person`, no `*Person`): sin riesgo de `nil`, cada entidad posee su copia |
| 2 | Formato de `id` (por ejemplo `P-001`): lo asigna el hospital o se pasa al constructor | Resuelto: **se pasa al constructor** (`NewPatient("P-001", ...)`); el `Hospital` valida unicidad al admitir/contratar y devuelve error si se repite |
| 4 | Tipo del campo "doctor que atendió" en `EpisodeRecord` cuando quien atiende es un `Orderly` (`Attender` vs `*Doctor`) | Resuelto: **`Attender`** (interfaz), de modo que el registro admite cualquier tipo de personal |
| 5 | Cómo registra el hospital al `Orderly` (`HireDoctor` solo acepta doctores): método adicional o contratación genérica | Resuelto: **`HireAttender(a Attender) error`** genérico; `HireDoctor(d *Doctor) error` se conserva por exigencia del enunciado y delega en él |
| 5 | Criterio de "disponible" para despachar un `Attender` | Resuelto: la interfaz `Attender` se amplía con **`IsAvailable() bool`** (extiende el mínimo del enunciado); el hospital recorre `[]Attender` y despacha al primero disponible sin conocer el tipo concreto |
| 6 | Quién orquesta el flujo de un ataque de sueño (paciente, hospital o personal) | Resuelto: **el `Hospital`**, mediante un método de alto nivel (`HandleSleepAttack(p *Patient, location string) (EpisodeRecord, error)`) que encadena `SufferSleepAttack` → despacho de un `Attender` → `AssignRoom` → `RegisterEpisode`. Justificación: el `Hospital` es la raíz del agregado y el único que conoce habitaciones, personal e historial; inyectar un asignador de camas en cada `Attender` agregaría una abstracción sin beneficio en este alcance y es más difícil de defender |
| 6 | Quién asigna el `id` del episodio | Resuelto: **el `Hospital` en `RegisterEpisode`**. `Attend` devuelve el registro sin `id` ni habitación; `RegisterEpisode` asigna el `id` (`E-001`, ...) con un contador interno del `Hospital` (sin estado global), completa la habitación asignada, lo guarda en el historial y lo añade al personal que atendió |
| 6 | Criterio de "episodios del día" para `SevereReport` | Resuelto: **fecha calendario actual con reloj inyectable**. El `Hospital` tiene un campo no exportado `now func() time.Time` (por defecto `time.Now`) y cuenta los episodios cuya fecha coincide con el día de `now()`; las pruebas inyectan un reloj fijo para resultados deterministas |
| 10 | Cómo se invoca la simulación concurrente sin alterar `go run .` (flag, subcomando o `cmd/`) | `[POR CONFIRMAR]` |
| 11 | Tecnología de la GUI (Fyne, Wails o REST + HTML/JS) | `[POR CONFIRMAR]` |

---

## 9. Definición de terminado (checklist de entrega)

- [ ] El proyecto corre desde un clon limpio con `go run .`
- [ ] `go test ./...` pasa
- [ ] `gofmt` aplicado y `go vet ./...` sin salida
- [ ] Los cinco structs existen con campos no exportados y métodos que los acceden
- [ ] `Person` embebido en `Patient` y `Doctor`, sin campos duplicados
- [ ] `[]Attender` con al menos dos tipos concretos distintos, usado polimórficamente
- [ ] Estados como tipos con constantes, sin strings sueltos
- [ ] Las cuatro consultas devuelven datos y se invocan desde `main`
- [ ] El caso "sin habitación disponible" devuelve error y no rompe el programa
- [ ] `README.md` (integrante, `go run .`, `go test ./...`, decisiones de diseño)
- [ ] `AI_USAGE.md` honesto: herramientas, 2-3 prompts decisivos, mínimo 2 casos de código incorrecto o no idiomático y su corrección, aprendizajes
- [ ] El autor puede explicar y modificar cualquier parte del código

---

## 10. Preparación para la sustentación

Preguntas y modificaciones en vivo esperadas (el código y los comentarios deben permitir responderlas):
- Agregar un tipo `Nurse` que cumpla `Attender` y despacharlo.
- Efecto de embeber `Person` por valor frente a por puntero.
- Cambiar un receptor de puntero a valor y qué deja de funcionar.
- Por qué un campo va en minúscula y qué impide el acceso desde fuera del paquete.
- Qué ocurre si dos goroutines asignan la misma habitación a la vez.
- Justificar decisiones: mapa frente a slice, tipo de retorno, separación de paquetes.

Al cerrar cada etapa, Claude Code debe dejar un resumen breve de las decisiones tomadas y su justificación.
