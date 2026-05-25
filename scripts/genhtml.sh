#!/usr/bin/env bash
# genhtml.sh — Run bazel coverage, generate an HTML report with genhtml, and open it.
set -euo pipefail

REPORT_DIR="/tmp/coverage-report"

echo "Running bazel coverage..."
bazel coverage --combined_report=lcov //...

LCOV_FILE="$(bazel info output_path)/_coverage/_coverage_report.dat"

if [[ ! -f "$LCOV_FILE" ]]; then
  echo "ERROR: No combined coverage report found at $LCOV_FILE" >&2
  exit 1
fi

echo "Generating HTML report..."
rm -rf "$REPORT_DIR"
genhtml "$LCOV_FILE" --output-directory "$REPORT_DIR" --quiet

echo "Coverage report generated at $REPORT_DIR/index.html"

# Open in browser (macOS: open, Linux: xdg-open)
if command -v open &>/dev/null; then
  open "$REPORT_DIR/index.html"
elif command -v xdg-open &>/dev/null; then
  xdg-open "$REPORT_DIR/index.html"
else
  echo "Open $REPORT_DIR/index.html in your browser."
fi
