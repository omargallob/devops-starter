FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

ARG BAZELISK_VERSION=1.25.0
ARG TARGETARCH=amd64

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    git \
    make \
    python3-pip \
    shellcheck \
    unzip \
    yamllint \
  && rm -rf /var/lib/apt/lists/*

# Bazelisk reads .bazelversion and downloads Bazel 9.0.0 on first use.
# Symlinked to `bazel` so all `bazel` calls transparently use Bazelisk.
RUN curl -fsSL \
    "https://github.com/bazelbuild/bazelisk/releases/download/v${BAZELISK_VERSION}/bazelisk-linux-${TARGETARCH}" \
    -o /usr/local/bin/bazelisk \
  && chmod +x /usr/local/bin/bazelisk \
  && ln -s /usr/local/bin/bazelisk /usr/local/bin/bazel

# mise installed system-wide; runtimes are installed per-user in post-create.sh
RUN MISE_INSTALL_PATH=/usr/local/bin/mise \
    curl -fsSL https://mise.run | sh
