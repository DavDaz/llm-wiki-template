# Guía conceptual — LLM Wiki

> Para entender qué está pasando realmente, no solo cómo ejecutarlo.

---

## Qué pregunta el wizard y por qué

### Nombre del wiki

Puramente cosmético. Aparece en el título del schema, en `wiki/index.md` y en `wiki/log.md`. No afecta ninguna lógica. Sirve para que cuando el CLI abra el proyecto sepa con qué dominio está trabajando.

---

### Slug del wiki

Tiene dos usos concretos:

1. **Nombre del directorio** que se crea (`banco-xyz/`)
2. **Campo `dominio:`** en el frontmatter de cada página generada

```yaml
---
tipo: proceso
titulo: Crear Usuario
dominio: banco-xyz   ← viene del slug
---
```

Es el identificador técnico del dominio. Útil si en el futuro tenés múltiples wikis y querés saber a cuál pertenece cada página.

---

### Idioma

La IA genera todas las páginas de `wiki/` en este idioma. Si ponés `es`, escribe en español. Si ponés `en`, en inglés. Solo afecta el contenido generado, no los archivos del sistema (`index.md`, `log.md`, el schema).

---

### Entidades primarias

Este es el campo más importante de entender bien.

**Las entidades NO son palabras clave por documento.** Son los conceptos centrales de tu dominio completo — independientemente de qué documentos vayas a cargar.

Cuando ejecutás `/wiki-ingest` con un documento nuevo, la IA usa las entidades como anclas para decidir qué merece su propia página en el wiki. Si `beneficiario` es una entidad del dominio y aparece mencionado en un manual de 80 páginas, la IA crea o actualiza `beneficiario.md` en el wiki.

**Ejemplo:** si tu dominio es un sistema de gestión de personal, las entidades son:

```
usuario, rol, permiso, departamento, cargo, legajo
```

Eso es todo el dominio. No importa si el primer documento que cargás habla solo de usuarios — las demás entidades ya están declaradas para cuando lleguen sus documentos.

**¿Qué pasa si surge una entidad nueva?**

No necesitás re-crear el wiki. Abrís `CLAUDE.md` o `AGENTS.md` y la agregás en la sección `entidades_primarias`. Después corrés `/wiki-lint` para detectar si hay páginas existentes que la mencionan sin tener su propia entrada.

---

### Tipos de página

Define la taxonomía del wiki. Cada página generada por la IA tiene exactamente uno de estos tipos en su frontmatter — lo que determina su estructura interna y cómo se nombra el archivo.

| Tipo | Cuándo usarlo | Formato del slug |
|------|---------------|-----------------|
| `proceso` | Pasos para hacer algo | `verbo-objeto.md` → `crear-usuario.md` |
| `referencia` | Listas, tablas, definiciones | `sustantivo.md` → `roles.md` |
| `entidad` | Descripción de un sistema o actor | `nombre-entidad.md` → `sistema-renab.md` |
| `politica` | Reglas o restricciones que deben cumplirse | `politica-tema.md` → `politica-acceso.md` |
| `regulacion` | Normativa legal con cita de fuente | `regulacion-tema.md` |
| `reporte` | Generado automáticamente por `/wiki-lint` | `lint-YYYY-MM-DD.md` |

El wizard pre-carga los tipos defaults (`proceso, referencia, entidad, politica`). Editá la lista para tu dominio — si no usás regulaciones legales, sacalo; si necesitás un tipo custom, agregalo.

---

### Convenciones específicas

Son las reglas de negocio que la IA aplica en **todas** las operaciones (ingest, query y lint). No son cosas que están escritas en tus documentos — son cosas que vos sabés que siempre deben cumplirse en tu dominio.

```
Todo proceso debe indicar el rol responsable de ejecutarlo
Los expedientes siempre tienen un número único de 8 dígitos
Los roles siempre listan sus permisos asociados explícitamente
```

Si no tenés convenciones claras al inicio, dejalo vacío. Emergen naturalmente a medida que usás el wiki y encontrás inconsistencias. Cuando las identifiques, las agregás directamente en `CLAUDE.md` / `AGENTS.md`.

---

## Cómo documentar manualmente en raw/

`raw/` es inmutable — la IA nunca modifica ni borra archivos ahí. Podés poner cualquier formato (PDF, Word convertido a MD, notas de reunión, exportaciones de Confluence). No hay estructura obligatoria.

Pero cuanto más claro y organizado sea el documento, mejor extrae la IA. Esta estructura funciona bien para documentos escritos manualmente:

```markdown
# [Título descriptivo del tema]

## Contexto
Breve descripción de qué cubre este documento y por qué existe.

## [Sección temática 1]
Contenido. Nombrá las entidades explícitamente — no "el sistema valida",
sino "el sistema RENAB valida". Cuanto más explícito, menos inferencia necesita la IA.

## [Sección temática 2]
...

## Notas / Pendientes
Cosas inciertas o no confirmadas. La IA las trata con `confianza: baja`
automáticamente, así no contaminan el wiki con información dudosa.
```

**Las dos reglas que más impactan la calidad del ingest:**

1. **Usá `##` para separar temas distintos.** La IA segmenta por secciones para decidir qué va a qué página. Sin headers, puede crear páginas mezcladas.

2. **Nombrá las entidades explícitamente.** Escribí los nombres reales — sistemas, roles, campos — no pronombres ni genéricos.

### Ejemplo de proceso bien documentado para raw/

```markdown
## Proceso: anulación de entrada contable

**Actor responsable:** Contador supervisor
**Precondición:** La entrada debe estar en estado "confirmada"

1. Ingresar al módulo de contabilidad
2. Buscar la entrada por número de expediente (8 dígitos)
3. Seleccionar "Anular" y completar el motivo obligatorio
4. El sistema genera un asiento de contrapartida automáticamente
5. El contador jefe debe aprobar dentro de las 24hs

**Resultado:** La entrada queda en estado "anulada".
El asiento de contrapartida queda en estado "borrador" hasta aprobación.
```

Eso le da a la IA todo lo necesario para generar `anular-entrada.md` con frontmatter correcto y `confianza: alta`.

---

## Cuándo evolucionar el schema

El schema del dominio (`CLAUDE.md` o `AGENTS.md`) no es estático — está diseñado para crecer con el dominio. Editarlo es normal y esperado.

| Situación | Qué hacer |
|-----------|-----------|
| Aparece un concepto nuevo que siempre va a existir | Agregar a `entidades_primarias` en el schema |
| Empezás a documentar un módulo nuevo con categorías propias | Agregar a `tipos_de_pagina` en el schema |
| Detectás que la IA siempre olvida algo importante | Agregar a `Convenciones específicas` |
| Habilitás o deshabilitás un tool backend | `llm-wiki add-tool opencode` / `llm-wiki remove-tool pi` + `llm-wiki migrate` |

Después de cualquier cambio en el schema, corré `/wiki-lint` para detectar páginas existentes que ya no cumplen las nuevas reglas.

---

## Agente local con Ollama

Si usás OpenCode o Pi con Ollama, podés crear un modelo especializado (`gpt-oss-wiki-agent`) que tiene el protocolo del wiki integrado en su system prompt. La diferencia con un modelo genérico:

| Modelo genérico | gpt-oss-wiki-agent |
|---|---|
| Tenés que explicarle el sistema cada sesión | Ya sabe todo — abrís y operás |
| Puede responder de memoria (sin leer el wiki) | Está instruido para NUNCA responder sin leer primero |
| Puede "ayudar" con cosas ajenas al wiki | Rechaza explícitamente cualquier cosa fuera de scope |
| El routing al comando correcto depende del usuario | Hace el routing automáticamente según el mensaje |

**Cómo funciona la combinación:**

```
Sistema prompt del Modelfile   →  sabe CÓMO operar el wiki (protocolos)
AGENTS.md / CLAUDE.md          →  sabe QUÉ es este wiki específico (dominio)
.pi/prompts/ o .opencode/commands/  →  las instrucciones detalladas de cada comando
```

Las tres capas se complementan. El Modelfile pone el marco general ("sos un agente de wiki, no hacés otra cosa"). El schema del dominio pone el contexto específico ("este wiki es sobre MIDES RENAB, las entidades son usuario, rol, permiso"). Los comandos ponen el protocolo paso a paso de cada operación.

**Crear el modelo:**

```bash
# Desde la raíz del repositorio llm-wiki
ollama create gpt-oss-wiki-agent -f ollama/wiki-agent.modelfile

# Verificar capabilities
ollama show gpt-oss-wiki-agent
```

**Nota de temperatura:** el Modelfile usa `temperature 0.2` — deliberadamente bajo. Para operaciones de wiki querés precisión y reproducibilidad, no creatividad. El modelo siempre va a seguir el protocolo en lugar de improvisar.

---

## El ciclo completo

```
llm-wiki init                 → configurás el dominio una vez
raw/ ← tus documentos        → tirás fuentes en cualquier momento
/wiki-ingest                  → la IA extrae y organiza el conocimiento
/wiki-query                   → preguntás en lenguaje natural
/wiki-lint                    → auditás consistencia periódicamente
CLAUDE.md / AGENTS.md         → evoluciona cuando el dominio cambia
llm-wiki manage               → gestionás tools sin tocar archivos
```
