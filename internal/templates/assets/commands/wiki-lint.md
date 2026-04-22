---
description: Audita el wiki completo y genera un reporte de salud con errores, advertencias e info.
---

# wiki-lint

Audita el wiki completo y genera un reporte de salud con severidades.

## Cuándo usar

- Periódicamente (sugerido: cada vez que se hayan procesado 5+ fuentes nuevas)
- Cuando se sospeche inconsistencia
- Antes de compartir el wiki con nuevas personas
- Después de actualizar el schema del dominio

Activadores:
- "audita el wiki"
- "revisa la consistencia"
- "¿hay problemas en el wiki?"
- `/wiki-lint`

---

## Protocolo de ejecución

### Paso 0 — Cargar el schema
Leer el schema del dominio (`CLAUDE.md` o `AGENTS.md`, el que exista en el proyecto) completo. Las reglas del dominio definen qué está mal y qué no.

### Paso 1 — Inventario
Listar todos los archivos en `wiki/` (excepto `index.md`, `log.md`, y reportes de lint anteriores).
Leer `wiki/index.md` para comparar contra archivos reales.

### Paso 2 — Ejecutar los 11 checks

**Checks de Error 🔴** (bloquean confiabilidad del wiki):

1. **Frontmatter incompleto** — páginas sin algún campo obligatorio (tipo, titulo, dominio, status, confianza, fuentes, actualizado)
2. **Slugs inválidos** — nombres de archivo que no siguen las reglas de nomenclatura del schema
3. **Wikilinks rotos** — `[[referencias]]` a páginas que no existen en `wiki/`
4. **Deprecados sin sucesor** — páginas con `status: deprecado` sin campo `ver_sucesor`
5. **Tipo inválido** — páginas con un `tipo` que no está definido en el schema

**Checks de Advertencia 🟡** (degradan calidad del wiki):

6. **Borradores viejos** — páginas con `status: borrador` creadas hace más de 30 días sin revisión
7. **Confianza baja sin nota** — páginas con `confianza: baja` sin el bloque de advertencia visible
8. **Páginas huérfanas** — páginas que no son referenciadas por ninguna otra página ni por `index.md`
9. **Conceptos sin página** — términos que aparecen como `[[wikilink]]` en múltiples páginas pero no tienen página propia

**Checks de Info 🔵** (oportunidades de mejora):

10. **Páginas largas** — páginas con más de 500 palabras que podrían dividirse
11. **Fuentes sin procesar** — archivos en `raw/` que no tienen entrada en `wiki/log.md`

### Paso 3 — Generar reporte
Guardar en `wiki/lint-YYYY-MM-DD.md`:

```markdown
---
tipo: reporte
titulo: Lint Report YYYY-MM-DD
dominio: [wiki-slug]
status: vigente
confianza: alta
fuentes: []
actualizado: YYYY-MM-DD
---

# Lint Report — YYYY-MM-DD

**Resumen:** X errores 🔴 · Y advertencias 🟡 · Z info 🔵

---

## 🔴 Errores (X)

### Frontmatter incompleto
- `wiki/nombre-pagina.md` — falta campo: `confianza`
- `wiki/otra-pagina.md` — falta campo: `fuentes`

### Wikilinks rotos
- `wiki/crear-usuario.md` → [[rol-supervisor]] (no existe)

---

## 🟡 Advertencias (Y)

### Borradores viejos (+30 días)
- `wiki/politica-acceso.md` — borrador desde 2026-03-01

### Páginas huérfanas
- `wiki/configuracion-smtp.md` — no referenciada por ninguna página

---

## 🔵 Info (Z)

### Páginas largas
- `wiki/sistema-renab.md` — 823 palabras, considerar dividir

### Fuentes sin procesar
- `raw/manual-v3.pdf` — sin entrada en log.md

---

## Acciones recomendadas

1. [acción concreta para resolver el error más crítico]
2. [acción concreta para el segundo]
3. ...
```

### Paso 4 — Actualizar index.md
Agregar el reporte generado al índice.

### Paso 5 — Registrar en log.md
```markdown
## YYYY-MM-DD HH:MM — lint

**Resultado:** X errores, Y advertencias, Z info
**Reporte:** [[lint-YYYY-MM-DD]]
**Acción requerida:** sí / no

---
```

---

## Qué NO hacer

- ❌ Nunca modificar páginas durante el lint — solo reportar
- ❌ Nunca borrar páginas huérfanas automáticamente — solo reportar
- ❌ Nunca corregir automáticamente errores sin confirmación del usuario
