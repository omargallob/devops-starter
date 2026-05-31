"""Rule for validating GoReleaser configuration.

Provides a test rule that runs goreleaser check against a config file.
Fails the test if the configuration is invalid.

Usage:
    load("//bazel/rules:goreleaser.bzl", "goreleaser_check_test")

    goreleaser_check_test(
        name = "goreleaser_check",
        config = ".goreleaser.yaml",
    )

Note: Requires goreleaser to be installed and available on PATH.
If not found, the test is skipped with a warning.
"""

def _goreleaser_check_test_impl(ctx):
    script = ctx.actions.declare_file(ctx.label.name + ".sh")

    config_path = ctx.file.config.short_path

    content = """#!/usr/bin/env bash
set -euo pipefail

# Resolve workspace directory.
# BUILD_WORKSPACE_DIRECTORY is set by "bazel run" but not "bazel test".
if [ -n "${{BUILD_WORKSPACE_DIRECTORY:-}}" ]; then
    cd "$BUILD_WORKSPACE_DIRECTORY"
else
    # For local tests, find the workspace root via git.
    WORKSPACE_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || true)
    if [ -n "$WORKSPACE_ROOT" ] && [ -f "$WORKSPACE_ROOT/go.mod" ]; then
        cd "$WORKSPACE_ROOT"
    else
        echo "ERROR: cannot determine workspace root" >&2
        exit 1
    fi
fi

if ! command -v goreleaser >/dev/null 2>&1; then
    echo "WARNING: goreleaser not found in PATH, skipping check" >&2
    echo "Install: https://goreleaser.com/install/" >&2
    exit 0
fi

echo "Running goreleaser check..."
echo "  Working directory: $(pwd)"
echo "  Config: {config_path}"

if goreleaser check --config "{config_path}" --quiet; then
    echo ""
    echo "PASSED: goreleaser config is valid"
else
    echo ""
    echo "FAILED: goreleaser config is invalid" >&2
    exit 1
fi
""".format(
        config_path = config_path,
    )

    ctx.actions.write(
        output = script,
        content = content,
        is_executable = True,
    )

    runfiles_files = [ctx.file.config]
    runfiles = ctx.runfiles(files = runfiles_files)

    return [DefaultInfo(
        executable = script,
        runfiles = runfiles,
    )]

goreleaser_check_test = rule(
    implementation = _goreleaser_check_test_impl,
    test = True,
    attrs = {
        "config": attr.label(
            allow_single_file = [".yml", ".yaml"],
            mandatory = True,
            doc = "GoReleaser config file (.goreleaser.yaml).",
        ),
    },
    doc = "Runs goreleaser check to validate the configuration as a test target.",
)
