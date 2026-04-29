package templates

import (
	"fmt"
	"strings"
)

// pageTypeDescriptions maps known page type names to their markdown description block.
// Unknown types fall back to a generic description.
var pageTypeDescriptions = map[string]string{
	"proceso": "### `proceso`\n" +
		"Describe cómo hacer algo: pasos secuenciales, precondiciones, actor responsable, resultado esperado.\n" +
		"Slug: `verbo-objeto.md` (ej: `crear-usuario.md`)\n",
	"referencia": "### `referencia`\n" +
		"Define qué es algo: términos, listas, tablas, configuraciones.\n" +
		"Slug: `sustantivo.md` (ej: `roles.md`, `permisos.md`)\n",
	"entidad": "### `entidad`\n" +
		"Describe un sistema, componente, actor o grupo específico del dominio.\n" +
		"Slug: `nombre-entidad.md` (ej: `sistema-renab.md`)\n",
	"politica": "### `politica`\n" +
		"Establece reglas, restricciones o lineamientos que deben cumplirse.\n" +
		"Slug: `politica-tema.md` (ej: `politica-acceso.md`)\n",
	"regulacion": "### `regulacion`\n" +
		"Documenta normativa legal o regulatoria aplicable al dominio. Debe citar artículo o fuente legal.\n" +
		"Slug: `regulacion-tema.md` (ej: `regulacion-datos-personales.md`)\n",
	"reporte": "### `reporte`\n" +
		"Resultado generado automáticamente por operaciones del wiki (lint, síntesis).\n" +
		"Slug: `lint-YYYY-MM-DD.md` o `reporte-tema.md`\n",
}

// SchemaData holds all values needed to render schema.md.template.
type SchemaData struct {
	WikiName         string
	WikiSlug         string
	Language         string
	CreatedDate      string
	PrimaryEntities  []string
	PageTypes        []string
	Conventions      []string // raw rule strings
	CommandsDir      string   // e.g. ".claude/skills" — kept for schema replacer
	CommandsTree     string   // pre-rendered directory tree for the commands section
	InstructionsFile string   // "CLAUDE.md" or "AGENTS.md"
}

// RenderSchema renders schema.md.template with the provided data and returns
// the resulting markdown content.
func RenderSchema(data SchemaData) (string, error) {
	raw, err := ReadFile("schema.md.template")
	if err != nil {
		return "", fmt.Errorf("read schema template: %w", err)
	}

	content := string(raw)
	r := strings.NewReplacer(
		"{{WIKI_NAME}}", data.WikiName,
		"{{WIKI_SLUG}}", data.WikiSlug,
		"{{WIKI_ROOT}}", data.WikiSlug,
		"{{LANGUAGE}}", data.Language,
		"{{CREATED_DATE}}", data.CreatedDate,
		"{{ENTITIES_LIST}}", buildYAMLList(data.PrimaryEntities),
		"{{PAGE_TYPES_LIST}}", buildYAMLList(data.PageTypes),
		"{{PAGE_TYPES_DETAIL}}", buildPageTypesDetail(data.PageTypes),
		"{{DOMAIN_CONVENTIONS}}", buildConventions(data.Conventions),
		"{{COMMANDS_DIR}}", data.CommandsDir,
		"{{COMMANDS_TREE}}", data.CommandsTree,
		"{{INSTRUCTIONS_FILE}}", data.InstructionsFile,
	)
	return r.Replace(content), nil
}

// RenderIndex returns the initial wiki/index.md content.
func RenderIndex(wikiName string) string {
	return "# Índice — " + wikiName + "\n\n" +
		"> Catálogo central del wiki. La IA lo lee primero en cada operación.\n" +
		"> No editar manualmente — se actualiza automáticamente con cada `/wiki-ingest`.\n\n" +
		"| Página | Descripción | Tipo | Status | Actualizado |\n" +
		"|--------|-------------|------|--------|-------------|\n\n" +
		"<!-- Las páginas se agregan aquí automáticamente durante el ingest -->\n"
}

// RenderLog returns the initial wiki/log.md content.
func RenderLog(wikiName, wikiSlug, createdDate string, entities, pageTypes []string) string {
	entitiesStr := strings.Join(entities, ", ")
	if entitiesStr == "" {
		entitiesStr = "(ninguna)"
	}
	pageTypesStr := strings.Join(pageTypes, ", ")

	return "# Log de Operaciones — " + wikiName + "\n\n" +
		"> Historial append-only. Nunca modificar entradas anteriores.\n" +
		"> Se actualiza automáticamente con cada `/wiki-ingest` y `/wiki-lint`.\n\n" +
		"---\n\n" +
		"## " + createdDate + " — setup\n\n" +
		"**Evento:** Wiki inicializado con llm-wiki\n" +
		"**Dominio:** " + wikiName + " (" + wikiSlug + ")\n" +
		"**Entidades primarias:** " + entitiesStr + "\n" +
		"**Tipos de página:** " + pageTypesStr + "\n\n" +
		"---\n"
}

// RenderSourcesRegistry returns the initial wiki/sources.json content.
// This file is the mutable operational state used by /wiki-ingest to detect
// whether a known source changed and should be reprocessed.
func RenderSourcesRegistry() string {
	return "{\n" +
		"  \"schema_version\": 1,\n" +
		"  \"sources\": {}\n" +
		"}\n"
}

// buildYAMLList converts a string slice to a YAML block list ("  - item\n").
func buildYAMLList(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString("  - ")
		sb.WriteString(item)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// buildPageTypesDetail produces the markdown detail block for all page types.
func buildPageTypesDetail(pageTypes []string) string {
	var sb strings.Builder
	for _, t := range pageTypes {
		if desc, ok := pageTypeDescriptions[t]; ok {
			sb.WriteString(desc)
		} else {
			sb.WriteString(fmt.Sprintf("### `%s`\nTipo personalizado para este dominio.\nSlug: `%s-tema.md`\n", t, t))
		}
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

// buildConventions produces the markdown conventions block.
func buildConventions(conventions []string) string {
	if len(conventions) == 0 {
		return "> Sin convenciones específicas definidas al momento del setup.\n" +
			"> Agregar aquí las reglas particulares de este dominio a medida que emerjan."
	}
	var sb strings.Builder
	for _, c := range conventions {
		sb.WriteString("- ")
		sb.WriteString(c)
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}
