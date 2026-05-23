"""Macro for building Go binaries for multiple platforms.

Uses Bazel's Starlark transition mechanism to cross-compile a go_binary
for each specified platform, then collects all outputs into a filegroup.

Usage:
    load("//bazel:go_release.bzl", "go_release")

    go_release(
        name = "release",
        binary = ":devops-starter",
        platforms = [
            "//platforms:linux_amd64",
            "//platforms:linux_arm64",
            "//platforms:darwin_amd64",
            "//platforms:darwin_arm64",
        ],
    )

This generates:
    :release_linux_amd64   - Binary built for linux/amd64
    :release_linux_arm64   - Binary built for linux/arm64
    :release_darwin_amd64  - Binary built for darwin/amd64
    :release_darwin_arm64  - Binary built for darwin/arm64
    :release               - Filegroup containing all of the above
"""

def _platform_transition_impl(settings, attr):
    return {"//command_line_option:platforms": attr.platform}

_platform_transition = transition(
    implementation = _platform_transition_impl,
    inputs = [],
    outputs = ["//command_line_option:platforms"],
)

def _go_release_binary_impl(ctx):
    # The binary attribute has been transitioned to the target platform.
    # Extract the output file and symlink it with the desired output name.
    src_files = ctx.attr.binary[0][DefaultInfo].files.to_list()
    if not src_files:
        fail("No files produced by binary target")

    src = src_files[0]
    output = ctx.actions.declare_file(ctx.attr.output_name)
    ctx.actions.symlink(output = output, target_file = src)

    return [DefaultInfo(
        files = depset([output]),
        runfiles = ctx.runfiles(files = [output]),
    )]

_go_release_binary = rule(
    implementation = _go_release_binary_impl,
    attrs = {
        "binary": attr.label(
            mandatory = True,
            cfg = _platform_transition,
        ),
        "platform": attr.string(mandatory = True),
        "output_name": attr.string(mandatory = True),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
)

def go_release(name, binary, platforms, prefix = None, visibility = None):
    """Build a Go binary for multiple platforms.

    Creates a transitioned binary target for each platform and a filegroup
    collecting all outputs.

    Args:
        name: Base name for generated targets.
        binary: Label of the go_binary to cross-compile.
        platforms: List of platform labels (e.g., ["//platforms:linux_amd64"]).
        prefix: Output filename prefix. Defaults to the binary target name.
        visibility: Visibility for all generated targets.
    """
    if not prefix:
        # Extract binary name from label: "//cmd/devops-starter" → "devops-starter"
        prefix = binary.split(":")[-1] if ":" in binary else binary.split("/")[-1]

    targets = []

    for platform in platforms:
        # Extract short name: "//platforms:linux_amd64" → "linux_amd64"
        short = platform.split(":")[-1]
        target_name = "{}_{}".format(name, short)

        # Output name: "devops-starter-linux-amd64" (dashes, not underscores)
        output_name = "{}-{}".format(prefix, short.replace("_", "-"))

        _go_release_binary(
            name = target_name,
            binary = binary,
            platform = platform,
            output_name = output_name,
            visibility = visibility,
        )
        targets.append(":{}".format(target_name))

    native.filegroup(
        name = name,
        srcs = targets,
        visibility = visibility,
    )
