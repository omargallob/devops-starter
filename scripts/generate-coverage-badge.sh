#!/usr/bin/env bash
# scripts/generate-coverage-badge.sh
#
# Generates a flat-style coverage badge SVG from a coverage percentage.
#
# Usage: ./scripts/generate-coverage-badge.sh <percentage> <output-file>
#   e.g.: ./scripts/generate-coverage-badge.sh 82.5 badges/coverage.svg

set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <percentage> <output-file>" >&2
  exit 1
fi

PCT="$1"
OUTPUT="$2"

# Determine badge color based on coverage threshold.
if (( $(echo "$PCT >= 80" | bc -l) )); then
  COLOR="#4c1"       # bright green
elif (( $(echo "$PCT >= 60" | bc -l) )); then
  COLOR="#dfb317"    # yellow
elif (( $(echo "$PCT >= 40" | bc -l) )); then
  COLOR="#fe7d37"    # orange
else
  COLOR="#e05d44"    # red
fi

# Format the label text.
LABEL="coverage"
MESSAGE="${PCT}%"

# Calculate widths: approximate 6.5px per character at font-size 11.
LABEL_WIDTH=$(echo "${#LABEL} * 6.5 + 10" | bc)
MSG_WIDTH=$(echo "${#MESSAGE} * 6.5 + 10" | bc)
# bc may return decimals — truncate to int.
LABEL_WIDTH=${LABEL_WIDTH%.*}
MSG_WIDTH=${MSG_WIDTH%.*}
TOTAL_WIDTH=$((LABEL_WIDTH + MSG_WIDTH))

# Center positions (scaled x10 for the transform="scale(.1)").
LABEL_X=$(( LABEL_WIDTH * 10 / 2 ))
MSG_X=$(( (LABEL_WIDTH + MSG_WIDTH / 2) * 10 ))

# Label and message textLength (scaled x10).
LABEL_TL=$(( (LABEL_WIDTH - 10) * 10 ))
MSG_TL=$(( (MSG_WIDTH - 10) * 10 ))

cat > "$OUTPUT" <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="${TOTAL_WIDTH}" height="20" role="img" aria-label="${LABEL}: ${MESSAGE}">
  <title>${LABEL}: ${MESSAGE}</title>
  <linearGradient id="s" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r"><rect width="${TOTAL_WIDTH}" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)">
    <rect width="${LABEL_WIDTH}" height="20" fill="#555"/>
    <rect x="${LABEL_WIDTH}" width="${MSG_WIDTH}" height="20" fill="${COLOR}"/>
    <rect width="${TOTAL_WIDTH}" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
    <text aria-hidden="true" x="${LABEL_X}" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="${LABEL_TL}">${LABEL}</text>
    <text x="${LABEL_X}" y="140" transform="scale(.1)" fill="#fff" textLength="${LABEL_TL}">${LABEL}</text>
    <text aria-hidden="true" x="${MSG_X}" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="${MSG_TL}">${MESSAGE}</text>
    <text x="${MSG_X}" y="140" transform="scale(.1)" fill="#fff" textLength="${MSG_TL}">${MESSAGE}</text>
  </g>
</svg>
SVG

echo "Badge written to ${OUTPUT} (${MESSAGE}, color: ${COLOR})"
