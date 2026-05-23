"""Macro for creating OCI test images from a Go binary.

Reduces the boilerplate of: layer_from_binary + image_manifest + tarball
into a single macro call.

Usage:
    load("//bazel:test_image.bzl", "test_image")

    test_image(
        name = "ubuntu_test",
        base = "@ubuntu_base",
        binary = "//cmd/devops-starter",
        repo_tag = "devops-starter:ubuntu-test",
    )

This generates:
    :ubuntu_test_layer    - Binary packaged as an OCI layer
    :ubuntu_test_image    - OCI image with the layer on the base
    :ubuntu_test_tarball  - Loadable tarball (docker load)
"""

load("@rules_img//img:image.bzl", "image_manifest")
load("@rules_img//img:layer.bzl", "layer_from_binary")

def test_image(name, base, binary, repo_tag = None, path = "/usr/local/bin/", visibility = None):
    """Create an OCI test image containing a Go binary.

    Args:
        name: Base name for generated targets.
        base: Label of the base OCI image (e.g., "@ubuntu_base").
        binary: Label of the go_binary to package.
        repo_tag: Docker repo:tag for the tarball. Defaults to "devops-starter:{name}".
        path: Directory in the container to place the binary. Defaults to /usr/local/bin/.
        visibility: Visibility for all generated targets.
    """
    if not repo_tag:
        repo_tag = "devops-starter:{}".format(name)

    layer_name = "{}_layer".format(name)
    image_name = "{}_image".format(name)
    tarball_name = "{}_tarball".format(name)

    layer_from_binary(
        name = layer_name,
        binary = binary,
        path = path,
        visibility = ["//visibility:private"],
    )

    image_manifest(
        name = image_name,
        base = base,
        layers = [":{}".format(layer_name)],
        visibility = visibility,
    )

    native.filegroup(
        name = tarball_name,
        srcs = [":{}".format(image_name)],
        output_group = "oci_tarball",
        visibility = visibility,
    )
