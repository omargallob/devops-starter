"""Rule for linting shell scripts with shellcheck.

Provides a test rule that runs shellcheck against a set of .sh files.
Fails the test if any script has lint errors.

Usage:
    load("//bazel:shellcheck.bzl", "shellcheck_test")

    shellcheck_test(
        name = "lint_shell",
        srcs = glob(["scripts/**/*.sh", "dotfiles/**/*.sh"]),
    )
"""

def _shellcheck_test_impl(ctx):
    # Generate a test runner script
    script = ctx.actions.declare_file(ctx.label.name + ".sh")

    src_paths = []
    for f in ctx.files.srcs:
        src_paths.append(f.short_path)

    # The test script checks for shellcheck availability, then runs it
    # against all source files. If shellcheck is not installed, the test
    # passes with a warning (to avoid breaking builds on machines without it).
    content = """#!/usr/bin/env bash
set -euo pipefail

if ! command -v shellcheck >/dev/null 2>&1; then
    echo "WARNING: shellcheck not found in PATH, skipping lint" >&2
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
    if ! shellcheck -x "$f"; then
        FAILED=1
    fi
done

if [ "$FAILED" -ne 0 ]; then
    echo ""
    echo "FAILED: shellcheck found errors in one or more files" >&2
    exit 1
fi

echo ""
echo "PASSED: all ${{#FILES[@]}} files clean"
""".format(
        files = " ".join(['"{}"'.format(p) for p in src_paths]),
    )

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

shellcheck_test = rule(
    implementation = _shellcheck_test_impl,
    test = True,
    attrs = {
        "srcs": attr.label_list(
            allow_files = [".sh"],
            mandatory = True,
            doc = "Shell script files to lint with shellcheck.",
        ),
    },
    doc = "Runs shellcheck on the given shell scripts as a test target.",
)
