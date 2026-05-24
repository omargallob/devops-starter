"""Rule for linting GitHub Actions workflow files with actionlint.

Provides a test rule that runs actionlint against the workflow directory.
Uses the workspace-root pattern (tags = ["local"]) so that actionlint can
validate cross-workflow references (needs, reusable workflows, etc.).

Usage:
    load("//bazel/rules:actionlint.bzl", "actionlint_test")

    actionlint_test(
        name = "actionlint",
        srcs = glob([".github/workflows/*.yml"]),
        tags = ["local"],
    )
"""

def _actionlint_test_impl(ctx):
    script = ctx.actions.declare_file(ctx.label.name + ".sh")

    content = """#!/usr/bin/env bash
set -euo pipefail

# Resolve workspace directory for cross-workflow validation.
if [ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]; then
    cd "$BUILD_WORKSPACE_DIRECTORY"
else
    WORKSPACE_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || true)
    if [ -n "$WORKSPACE_ROOT" ] && [ -d "$WORKSPACE_ROOT/.github/workflows" ]; then
        cd "$WORKSPACE_ROOT"
    else
        echo "ERROR: cannot determine workspace root" >&2
        exit 1
    fi
fi

if ! command -v actionlint >/dev/null 2>&1; then
    echo "WARNING: actionlint not found in PATH, skipping lint" >&2
    echo "Install: https://github.com/rhysd/actionlint#install" >&2
    exit 0
fi

echo "Running actionlint..."
echo "  Working directory: $(pwd)"

if actionlint; then
    echo ""
    echo "PASSED: actionlint found no issues"
else
    echo ""
    echo "FAILED: actionlint found issues" >&2
    exit 1
fi
"""

    ctx.actions.write(
        output = script,
        content = content,
        is_executable = True,
    )

    runfiles = ctx.runfiles(files = ctx.files.srcs)

    return [DefaultInfo(
        executable = script,
        runfiles = runfiles,
    )]

actionlint_test = rule(
    implementation = _actionlint_test_impl,
    test = True,
    attrs = {
        "srcs": attr.label_list(
            allow_files = [".yaml", ".yml"],
            doc = "Workflow files (included in runfiles for dependency tracking).",
        ),
    },
    doc = "Runs actionlint on GitHub Actions workflows as a test target.",
)
