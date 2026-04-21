#!/bin/bash
# =============================================================================
# llm-wiki-template — setup.sh
# Genera un wiki nuevo a partir del template, configurado para tu dominio.
# Uso: ./setup.sh
# =============================================================================

set -e

# Colores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo ""
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║       LLM Wiki — Setup               ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

# ─────────────────────────────────────────
# 1. Recolectar datos del dominio
# ─────────────────────────────────────────

echo -e "${YELLOW}→ Nombre del wiki${NC} (ej: MIDES RENAB, Banco XYZ):"
read -r WIKI_NAME

echo -e "${YELLOW}→ Slug del wiki${NC} (kebab-case, sin espacios, ej: mides-renab, banco-xyz):"
read -r WIKI_SLUG

echo -e "${YELLOW}→ Idioma principal${NC} (ej: es, en) [default: es]:"
read -r LANGUAGE
LANGUAGE=${LANGUAGE:-es}

echo -e "${YELLOW}→ Directorio destino${NC} (ej: ../mides-renab-wiki) [default: ./${WIKI_SLUG}-wiki]:"
read -r WIKI_DIR
WIKI_DIR=${WIKI_DIR:-"./${WIKI_SLUG}-wiki"}

echo ""
echo -e "${YELLOW}→ Entidades primarias del dominio${NC}"
echo "  Son los 'sustantivos' principales: objetos, actores, sistemas que existen en tu dominio."
echo "  Escribe una por línea. Línea vacía para terminar."
echo "  Ej: usuario, beneficiario, rol, permiso, sistema"
echo ""

ENTITIES=()
while true; do
    read -r -p "  Entidad: " entity
    [[ -z "$entity" ]] && break
    ENTITIES+=("$entity")
done

echo ""
echo -e "${YELLOW}→ Tipos de página del dominio${NC}"
echo "  Opciones sugeridas: proceso, referencia, entidad, politica, regulacion, reporte"
echo "  Escribe los que aplican, uno por línea. Línea vacía para usar los defaults."
echo ""

PAGE_TYPES=()
while true; do
    read -r -p "  Tipo: " ptype
    [[ -z "$ptype" ]] && break
    PAGE_TYPES+=("$ptype")
done

# Defaults si no ingresó tipos
if [ ${#PAGE_TYPES[@]} -eq 0 ]; then
    PAGE_TYPES=("proceso" "referencia" "entidad" "politica")
    echo "  Usando defaults: ${PAGE_TYPES[*]}"
fi

echo ""
echo -e "${YELLOW}→ Convenciones específicas del dominio${NC}"
echo "  Reglas particulares de TU dominio que la IA debe seguir siempre."
echo "  Escribe una por línea. Línea vacía para terminar."
echo "  Ej: 'Los procesos siempre incluyen el actor responsable'"
echo "      'Los roles siempre listan sus permisos asociados'"
echo ""

CONVENTIONS=()
while true; do
    read -r -p "  Convención: " conv
    [[ -z "$conv" ]] && break
    CONVENTIONS+=("$conv")
done

# ─────────────────────────────────────────
# 2. Crear estructura de directorios
# ─────────────────────────────────────────

echo ""
echo -e "${BLUE}Creando wiki en: ${WIKI_DIR}${NC}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "${WIKI_DIR}/raw"
mkdir -p "${WIKI_DIR}/wiki"
mkdir -p "${WIKI_DIR}/.claude/commands"

touch "${WIKI_DIR}/raw/.gitkeep"

# ─────────────────────────────────────────
# 3. Construir bloques de contenido
# ─────────────────────────────────────────

CREATED_DATE=$(date +%Y-%m-%d)

# Entities list para YAML
ENTITIES_LIST=""
for e in "${ENTITIES[@]}"; do
    ENTITIES_LIST="${ENTITIES_LIST}  - ${e}\n"
done

# Page types list para YAML
PAGE_TYPES_LIST=""
for t in "${PAGE_TYPES[@]}"; do
    PAGE_TYPES_LIST="${PAGE_TYPES_LIST}  - ${t}\n"
done

# Page types detail — descripción de cada tipo
PAGE_TYPES_DETAIL=""
for t in "${PAGE_TYPES[@]}"; do
    case "$t" in
        proceso)
            PAGE_TYPES_DETAIL="${PAGE_TYPES_DETAIL}### \`proceso\`\nDescribe cómo hacer algo: pasos secuenciales, precondiciones, actor responsable, resultado esperado.\nSlug: \`verbo-objeto.md\` (ej: \`crear-usuario.md\`)\n\n"
            ;;
        referencia)
            PAGE_TYPES_DETAIL="${PAGE_TYPES_DETAIL}### \`referencia\`\nDefine qué es algo: términos, listas, tablas, configuraciones.\nSlug: \`sustantivo.md\` (ej: \`roles.md\`, \`permisos.md\`)\n\n"
            ;;
        entidad)
            PAGE_TYPES_DETAIL="${PAGE_TYPES_DETAIL}### \`entidad\`\nDescribe un sistema, componente, actor o grupo específico del dominio.\nSlug: \`nombre-entidad.md\` (ej: \`sistema-renab.md\`)\n\n"
            ;;
        politica)
            PAGE_TYPES_DETAIL="${PAGE_TYPES_DETAIL}### \`politica\`\nEstablece reglas, restricciones o lineamientos que deben cumplirse.\nSlug: \`politica-tema.md\` (ej: \`politica-acceso.md\`)\n\n"
            ;;
        regulacion)
            PAGE_TYPES_DETAIL="${PAGE_TYPES_DETAIL}### \`regulacion\`\nDocumenta normativa legal o regulatoria aplicable al dominio. Debe citar artículo o fuente legal.\nSlug: \`regulacion-tema.md\` (ej: \`regulacion-datos-personales.md\`)\n\n"
            ;;
        reporte)
            PAGE_TYPES_DETAIL="${PAGE_TYPES_DETAIL}### \`reporte\`\nResultado generado automáticamente por operaciones del wiki (lint, síntesis).\nSlug: \`lint-YYYY-MM-DD.md\` o \`reporte-tema.md\`\n\n"
            ;;
        *)
            PAGE_TYPES_DETAIL="${PAGE_TYPES_DETAIL}### \`${t}\`\nTipo personalizado para este dominio.\nSlug: \`${t}-tema.md\`\n\n"
            ;;
    esac
done

# Domain conventions
DOMAIN_CONVENTIONS=""
if [ ${#CONVENTIONS[@]} -eq 0 ]; then
    DOMAIN_CONVENTIONS="> Sin convenciones específicas definidas al momento del setup.\n> Agregar aquí las reglas particulares de este dominio a medida que emerjan."
else
    for c in "${CONVENTIONS[@]}"; do
        DOMAIN_CONVENTIONS="${DOMAIN_CONVENTIONS}- ${c}\n"
    done
fi

# ─────────────────────────────────────────
# 4. Generar CLAUDE.md
# ─────────────────────────────────────────

sed \
    -e "s|{{WIKI_NAME}}|${WIKI_NAME}|g" \
    -e "s|{{WIKI_SLUG}}|${WIKI_SLUG}|g" \
    -e "s|{{WIKI_ROOT}}|${WIKI_SLUG}-wiki|g" \
    -e "s|{{LANGUAGE}}|${LANGUAGE}|g" \
    -e "s|{{CREATED_DATE}}|${CREATED_DATE}|g" \
    -e "s|{{ENTITIES_LIST}}|${ENTITIES_LIST}|g" \
    -e "s|{{PAGE_TYPES_LIST}}|${PAGE_TYPES_LIST}|g" \
    "${SCRIPT_DIR}/CLAUDE.md.template" > "${WIKI_DIR}/CLAUDE.md"

# Reemplazos multilinea con Python (más confiable que sed para bloques)
python3 - <<PYEOF
import re

with open("${WIKI_DIR}/CLAUDE.md", "r") as f:
    content = f.read()

content = content.replace("{{PAGE_TYPES_DETAIL}}", """${PAGE_TYPES_DETAIL}""")
content = content.replace("{{DOMAIN_CONVENTIONS}}", """${DOMAIN_CONVENTIONS}""")

with open("${WIKI_DIR}/CLAUDE.md", "w") as f:
    f.write(content)
PYEOF

# ─────────────────────────────────────────
# 5. Generar index.md y log.md
# ─────────────────────────────────────────

cat > "${WIKI_DIR}/wiki/index.md" << EOF
# Índice — ${WIKI_NAME}

> Catálogo central del wiki. La IA lo lee primero en cada operación.
> No editar manualmente — se actualiza automáticamente con cada \`/wiki-ingest\`.

| Página | Descripción | Tipo | Status | Actualizado |
|--------|-------------|------|--------|-------------|

<!-- Las páginas se agregan aquí automáticamente durante el ingest -->
EOF

cat > "${WIKI_DIR}/wiki/log.md" << EOF
# Log de Operaciones — ${WIKI_NAME}

> Historial append-only. Nunca modificar entradas anteriores.
> Se actualiza automáticamente con cada \`/wiki-ingest\` y \`/wiki-lint\`.

---

## ${CREATED_DATE} — setup

**Evento:** Wiki inicializado con setup.sh
**Dominio:** ${WIKI_NAME} (${WIKI_SLUG})
**Entidades primarias:** $(IFS=', '; echo "${ENTITIES[*]}")
**Tipos de página:** $(IFS=', '; echo "${PAGE_TYPES[*]}")

---
EOF

# ─────────────────────────────────────────
# 6. Copiar skills
# ─────────────────────────────────────────

cp "${SCRIPT_DIR}/.claude/commands/wiki-ingest.md" "${WIKI_DIR}/.claude/commands/"
cp "${SCRIPT_DIR}/.claude/commands/wiki-query.md"  "${WIKI_DIR}/.claude/commands/"
cp "${SCRIPT_DIR}/.claude/commands/wiki-lint.md"   "${WIKI_DIR}/.claude/commands/"

# ─────────────────────────────────────────
# 7. Inicializar Git
# ─────────────────────────────────────────

cd "${WIKI_DIR}"

cat > .gitignore << EOF
.DS_Store
*.swp
*.tmp
EOF

git init -q
git add .
git commit -q -m "chore: init wiki ${WIKI_NAME}"

cd - > /dev/null

# ─────────────────────────────────────────
# 8. Resumen final
# ─────────────────────────────────────────

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   ✓ Wiki creado exitosamente                 ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  📁 Directorio: ${BLUE}${WIKI_DIR}${NC}"
echo -e "  📄 Schema:     ${BLUE}${WIKI_DIR}/CLAUDE.md${NC}"
echo -e "  📂 Fuentes:    ${BLUE}${WIKI_DIR}/raw/${NC}"
echo -e "  📂 Wiki:       ${BLUE}${WIKI_DIR}/wiki/${NC}"
echo ""
echo -e "${YELLOW}Próximos pasos:${NC}"
echo "  1. Copia tus documentos existentes en raw/"
echo "  2. Abre Claude Code en el directorio del wiki"
echo "  3. Ejecuta: /wiki-ingest"
echo "  4. Pregunta lo que necesites: /wiki-query ¿qué roles existen?"
echo ""
