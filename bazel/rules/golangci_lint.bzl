"""Rule for linting Go code with golangci-lint.

Provides a test rule that runs golangci-lint against Go packages.
Fails the test if any lint issues are found.

Usage:
    load("//bazel/rules:golangci_lint.bzl", "golangci_lint_test")

    golangci_lint_test(
        name = "golangci_lint",
        srcs = ["//..."],  # not actually used for file input
        config = ".golangci.yml",  # optional config file
    )

Note: golangci-lint operates on Go packages, not individual files.
This rule runs it at the workspace root against ./... by default.
"""

def _golangci_lint_test_impl(ctx):
    script = ctx.actions.declare_file(ctx.label.name + ".sh")

    config_arg = ""
    if ctx.file.config:
        config_arg = "--config \"$(pwd)/{}\"".format(ctx.file.config.short_path)

    paths = " ".join(ctx.attr.paths)

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

if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "WARNING: golangci-lint not found in PATH, skipping lint" >&2
    echo "Install: https://golangci-lint.run/welcome/install/" >&2
    exit 0
fi

echo "Running golangci-lint..."
echo "  Working directory: $(pwd)"
echo "  Paths: {paths}"
{config_echo}

if golangci-lint run {config_arg} --timeout {timeout} {paths}; then
    echo ""
    echo "PASSED: golangci-lint found no issues"
else
    echo ""
    echo "FAILED: golangci-lint found issues" >&2
    exit 1
fi
""".format(
        workspace_name = ctx.workspace_name,
        paths = paths,
        config_arg = config_arg,
        config_echo = 'echo "  Config: {}"'.format(ctx.file.config.short_path) if ctx.file.config else "",
        timeout = ctx.attr.timeout_duration,
    )

    ctx.actions.write(
        output = script,
        content = content,
        is_executable = True,
    )

    runfiles_files = []
    if ctx.file.config:
        runfiles_files.append(ctx.file.config)

    # Include Go source files in runfiles so bazel test can access them
    runfiles_files.extend(ctx.files.srcs)

    runfiles = ctx.runfiles(files = runfiles_files)

    return [DefaultInfo(
        executable = script,
        runfiles = runfiles,
    )]

golangci_lint_test = rule(
    implementation = _golangci_lint_test_impl,
    test = True,
    attrs = {
        "srcs": attr.label_list(
            allow_files = True,
            doc = "Go source files to include in runfiles (for bazel test sandboxing).",
        ),
        "config": attr.label(
            allow_single_file = [".yml", ".yaml", ".toml"],
            mandatory = False,
            doc = "Optional golangci-lint config file (.golangci.yml).",
        ),
        "paths": attr.string_list(
            default = ["./..."],
            doc = "Go package patterns to lint. Defaults to ./...",
        ),
        "timeout_duration": attr.string(
            default = "5m",
            doc = "Timeout for golangci-lint execution.",
        ),
    },
    doc = "Runs golangci-lint on Go packages as a test target.",
)
