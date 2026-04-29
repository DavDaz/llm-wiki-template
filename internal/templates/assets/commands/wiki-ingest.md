---
name: wiki-ingest
description: Procesa fuentes nuevas en raw/ o reprocesa una fuente específica si cambió.
argument-hint: [archivo opcional | --force archivo]
---

# wiki-ingest

Compila conocimiento desde `raw/` hacia `wiki/`.

Tiene dos modos:

- `/wiki-ingest` → modo automático (solo fuentes nuevas)
- `/wiki-ingest raw/archivo.ext` → modo dirigido (solo ese archivo)

## Cuándo usar

Cuando se agregue uno o más archivos a `raw/`, o cuando se actualice un archivo ya procesado y se quiera re-ingestarlo sin reprocesar todo.

Activadores:
- "procesa los archivos nuevos"
- "ingesta raw/"
- "hay un doc nuevo en raw/"
- "reprocesá raw/manual.md"
- `/wiki-ingest`

---

## Protocolo de ejecución

### Paso 0 — Cargar el schema
Leer el schema del dominio (`CLAUDE.md` o `AGENTS.md`, el que exista en el proyecto) completo antes de cualquier otra acción.
Las reglas del dominio en el schema tienen prioridad sobre cualquier intuición.

### Paso 1 — Determinar modo

#### Modo A — automático (`/wiki-ingest`)
1. Leer `wiki/log.md` para obtener la lista de archivos ya procesados.
2. Listar archivos en `raw/` que **no** aparezcan en el log.
3. Si no hay archivos nuevos → responder "No hay fuentes nuevas en raw/ para procesar." y detener.
4. Definir `fuentes_a_procesar = [archivos nuevos detectados]`.

#### Modo B — dirigido (`/wiki-ingest raw/archivo.ext`)
1. Validar que la ruta exista y esté dentro de `raw/`.
2. Definir `fuentes_a_procesar = [archivo objetivo]` (solo uno).
3. Leer `wiki/sources.json`; si no existe, inicializar:
   ```json
   {
     "schema_version": 1,
     "sources": {}
   }
   ```
4. Calcular `fingerprint_actual` del archivo objetivo (hash/huella determinística del contenido).
5. Si el archivo ya existe en `sources.json` con el mismo fingerprint y NO se pasó `--force`:
   - responder: "El archivo objetivo no cambió desde el último ingest. No hay acciones para ejecutar."
   - detener.
6. Si cambió o no existe en `sources.json`, continuar.

### Paso 2 — Analizar cada fuente
Por cada archivo en `fuentes_a_procesar`:

1. Leer el archivo completo
2. Identificar y listar:
   - **Entidades** mencionadas (sistemas, roles, personas, grupos)
   - **Procesos** descritos (pasos, flujos, procedimientos)
   - **Referencias** definidas (términos, configuraciones, tablas)
   - **Políticas** o reglas establecidas
   - **Contradicciones** con lo que ya sé del wiki (si leí index.md)

3. Mostrar el análisis al usuario antes de escribir nada:
   ```
   Modo: [automatico|dirigido]
   Fuente: raw/nombre-archivo.ext
   Voy a crear: [lista de páginas nuevas]
   Voy a actualizar: [lista de páginas existentes]
   Contradicciones detectadas: [lista o "ninguna"]
   ¿Continuar? (s/n)
   ```

### Paso 3 — Verificar duplicados
Leer `wiki/index.md` completo.
Por cada concepto identificado en el paso 2:
- Buscar si ya existe una página con ese concepto
- Si existe con >50% overlap → planificar actualización, no creación
- Si existe con nombre similar pero diferente → planificar link cruzado

### Paso 4 — Escribir páginas
Para cada página a crear o actualizar:

**Formato de página nueva:**
```markdown
---
tipo: [proceso|referencia|entidad|politica]
titulo: [Nombre Legible]
dominio: [wiki-slug definido en el schema]
status: borrador
confianza: [alta|media|baja según reglas del schema]
fuentes: [raw/nombre-archivo.ext]
actualizado: YYYY-MM-DD
---

# [Título]

[Contenido estructurado con headings claros]

## Ver también
- [[pagina-relacionada-1]]
- [[pagina-relacionada-2]]
```

**Reglas al escribir:**
- Headings claros, párrafos cortos, bullet points para listas
- Usar `[[wikilinks]]` en lugar de duplicar contenido
- Si hay contradicción → agregar bloque visible:
  ```
  > ⚠️ Contradicción detectada: este documento dice X, pero [[pagina-existente]] dice Y.
  > Pendiente verificación humana.
  ```
- Si la confianza es baja → agregar:
  ```
  > ⚠️ Confianza baja: esta información fue inferida sin fuente directa.
  ```

### Paso 5 — Actualizar index.md
Agregar o actualizar una línea por cada página tocada:

```markdown
| [[slug-pagina]] | Descripción en una línea | tipo | status | YYYY-MM-DD |
```

El `index.md` sigue este formato:
```markdown
# Índice — [Wiki Name]

| Página | Descripción | Tipo | Status | Actualizado |
|--------|-------------|------|--------|-------------|
| [[crear-usuario]] | Pasos para crear un usuario en el sistema | proceso | vigente | 2026-04-21 |
```

### Paso 6 — Registrar estado de fuentes
Actualizar `wiki/sources.json` para cada fuente procesada:

```json
{
  "schema_version": 1,
  "sources": {
    "raw/nombre-archivo.ext": {
      "fingerprint": "[hash/huella]",
      "processed_at": "YYYY-MM-DD HH:MM",
      "pages_touched": ["wiki/pag-1.md", "wiki/pag-2.md"]
    }
  }
}
```

### Paso 7 — Registrar en log.md
Agregar al **final** de `wiki/log.md` (nunca modificar entradas anteriores):

```markdown
## YYYY-MM-DD HH:MM — ingest

**Modo:** automatico | dirigido
**Fuente:** `raw/nombre-archivo.ext`
**Páginas creadas:** [[pag-1]], [[pag-2]]
**Páginas actualizadas:** [[pag-3]]
**Contradicciones:** ninguna | [descripción si las hay]
**Fingerprint:** [hash/huella usada]
**Notas:** [observaciones relevantes]

---
```

---

## Qué NO hacer

- ❌ Nunca modificar o borrar archivos en `raw/`
- ❌ Nunca crear páginas sin frontmatter
- ❌ Nunca ignorar una contradicción — siempre anotarla
- ❌ Nunca responder desde memoria — siempre leer el wiki primero
- ❌ Nunca crear una página nueva si ya existe una con ese concepto
- ❌ Nunca reprocesar TODO `raw/` cuando el modo dirigido pide un solo archivo
