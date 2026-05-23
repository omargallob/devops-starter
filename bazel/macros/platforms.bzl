"""Macro for declaring cross-compilation platforms.

Reduces boilerplate when defining multiple platform() targets by
accepting a simple list of (name, os, cpu) tuples.

Usage:
    load("//bazel:platforms.bzl", "declare_platforms")

    declare_platforms(
        platforms = [
            ("linux_amd64", "linux", "x86_64"),
            ("linux_arm64", "linux", "aarch64"),
            ("darwin_amd64", "darwin", "x86_64"),
            ("darwin_arm64", "darwin", "aarch64"),
        ],
    )
"""

# Mapping from common names to @platforms constraint values
_OS_CONSTRAINT = {
    "linux": "@platforms//os:linux",
    "darwin": "@platforms//os:macos",
    "macos": "@platforms//os:macos",
    "windows": "@platforms//os:windows",
}

_CPU_CONSTRAINT = {
    "x86_64": "@platforms//cpu:x86_64",
    "amd64": "@platforms//cpu:x86_64",
    "aarch64": "@platforms//cpu:aarch64",
    "arm64": "@platforms//cpu:aarch64",
}

def declare_platforms(platforms, visibility = None):
    """Declare platform targets from a list of tuples.

    Args:
        platforms: List of (name, os, cpu) tuples. os and cpu are looked up
            in internal mappings to resolve the correct constraint values.
        visibility: Optional visibility for all generated platforms.
    """
    for (name, os, cpu) in platforms:
        os_constraint = _OS_CONSTRAINT.get(os)
        if not os_constraint:
            fail("Unknown OS '{}' for platform '{}'. Known: {}".format(
                os,
                name,
                ", ".join(_OS_CONSTRAINT.keys()),
            ))

        cpu_constraint = _CPU_CONSTRAINT.get(cpu)
        if not cpu_constraint:
            fail("Unknown CPU '{}' for platform '{}'. Known: {}".format(
                cpu,
                name,
                ", ".join(_CPU_CONSTRAINT.keys()),
            ))

        kwargs = {
            "name": name,
            "constraint_values": [os_constraint, cpu_constraint],
        }
        if visibility:
            kwargs["visibility"] = visibility

        native.platform(**kwargs)
