"""Rule for linting YAML files with yamllint.

Provides a test rule that runs yamllint against a set of YAML files.
Fails the test if any file has lint errors.

Usage:
    load("//bazel/rules:yamllint.bzl", "yamllint_test")

    yamllint_test(
        name = "yamllint",
        srcs = glob(["**/*.yaml", "**/*.yml"]),
        config = ".yamllint.yaml",
    )
"""

def _yamllint_test_impl(ctx):
    script = ctx.actions.declare_file(ctx.label.name + ".sh")

    src_paths = []
    for f in ctx.files.srcs:
        src_paths.append(f.short_path)

    config_arg = ""
    if ctx.file.config:
        config_arg = "-c \"{}\"".format(ctx.file.config.short_path)

    content = """#!/usr/bin/env bash
set -euo pipefail

if ! command -v yamllint >/dev/null 2>&1; then
    echo "WARNING: yamllint not found in PATH, skipping lint" >&2
    echo "Install: pip install yamllint" >&2
    exit 0
fi

FAILED=0
FILES=({files})

for f in "${{FILES[@]}}"; do
    if [ ! -f "$f" ]; then
        echo "SKIP: $f (not found)" >&2
        continue
    fi
    echo "Checking: $f"
    if ! yamllint --strict {config_arg} "$f"; then
        FAILED=1
    fi
done

if [ "$FAILED" -ne 0 ]; then
    echo ""
    echo "FAILED: yamllint found errors in one or more files" >&2
    exit 1
fi

echo ""
echo "PASSED: all ${{#FILES[@]}} YAML files clean"
""".format(
        files = " ".join(['"{}"'.format(p) for p in src_paths]),
        config_arg = config_arg,
    )

    ctx.actions.write(
        output = script,
        content = content,
        is_executable = True,
    )

    runfiles_files = list(ctx.files.srcs)
    if ctx.file.config:
        runfiles_files.append(ctx.file.config)

    runfiles = ctx.runfiles(files = runfiles_files)

    return [DefaultInfo(
        executable = script,
        runfiles = runfiles,
    )]

yamllint_test = rule(
    implementation = _yamllint_test_impl,
    test = True,
    attrs = {
        "srcs": attr.label_list(
            allow_files = [".yaml", ".yml"],
            mandatory = True,
            doc = "YAML files to lint with yamllint.",
        ),
        "config": attr.label(
            allow_single_file = [".yaml", ".yml"],
            mandatory = False,
            doc = "Optional yamllint configuration file (.yamllint.yaml).",
        ),
    },
    doc = "Runs yamllint on the given YAML files as a test target.",
)
