---
description: Procesa archivos nuevos en raw/ y los compila en el wiki. Úsalo cuando agregues documentos a raw/.
argument-hint: [archivo opcional]
---

# wiki-ingest

Procesa archivos nuevos en `raw/` y compila el conocimiento en `wiki/`.

## Cuándo usar

Cuando se agregue uno o más archivos a `raw/` y se quiera incorporar su contenido al wiki.

Activadores:
- "procesa los archivos nuevos"
- "ingesta raw/"
- "hay un doc nuevo en raw/"
- `/wiki-ingest`

---

## Protocolo de ejecución

### Paso 0 — Cargar el schema
Leer el schema del dominio (`CLAUDE.md` o `AGENTS.md`, el que exista en el proyecto) completo antes de cualquier otra acción.
Las reglas del dominio en el schema tienen prioridad sobre cualquier intuición.

### Paso 1 — Identificar fuentes nuevas
Leer `wiki/log.md` para obtener la lista de archivos ya procesados.
Listar archivos en `raw/` que **no** aparezcan en el log.
Si no hay archivos nuevos → responder "No hay fuentes nuevas en raw/ para procesar." y detener.

### Paso 2 — Analizar cada fuente
Por cada archivo nuevo en `raw/`:

1. Leer el archivo completo
2. Identificar y listar:
   - **Entidades** mencionadas (sistemas, roles, personas, grupos)
   - **Procesos** descritos (pasos, flujos, procedimientos)
   - **Referencias** definidas (términos, configuraciones, tablas)
   - **Políticas** o reglas establecidas
   - **Contradicciones** con lo que ya sé del wiki (si leí index.md)

3. Mostrar el análisis al usuario antes de escribir nada:
   ```
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

### Paso 6 — Registrar en log.md
Agregar al **final** de `wiki/log.md` (nunca modificar entradas anteriores):

```markdown
## YYYY-MM-DD HH:MM — ingest

**Fuente:** `raw/nombre-archivo.ext`
**Páginas creadas:** [[pag-1]], [[pag-2]]
**Páginas actualizadas:** [[pag-3]]
**Contradicciones:** ninguna | [descripción si las hay]
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
