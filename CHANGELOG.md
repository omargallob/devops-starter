# Changelog

## [0.1.4](https://github.com/omargallob/devops-starter/compare/v0.1.3...v0.1.4) (2026-05-25)


### Features

* **registry:** add lcov and genhtml tools ([#34](https://github.com/omargallob/devops-starter/issues/34)) ([a1a8162](https://github.com/omargallob/devops-starter/commit/a1a8162344a0e72a1bd35e26d9cb1636b918fab5))

## [0.1.3](https://github.com/omargallob/devops-starter/compare/v0.1.2...v0.1.3) (2026-05-25)


### Features

* **installer:** improve test coverage and fix binary format install bug ([#32](https://github.com/omargallob/devops-starter/issues/32)) ([9c1942b](https://github.com/omargallob/devops-starter/commit/9c1942bbbc80e18af283f42c8f5c16be584ba7ac))


### Bug Fixes

* **ci:** push coverage badge to unprotected badges branch ([#30](https://github.com/omargallob/devops-starter/issues/30)) ([dc60295](https://github.com/omargallob/devops-starter/commit/dc60295d13285510a1727a3a6675f5d276a5b352))
* **ci:** use GitHub API to update coverage badge on protected branch ([#28](https://github.com/omargallob/devops-starter/issues/28)) ([4610d66](https://github.com/omargallob/devops-starter/commit/4610d66da7448c413109e3fa4718daf364e4eec9))

## [0.1.2](https://github.com/omargallob/devops-starter/compare/v0.1.1...v0.1.2) (2026-05-25)


### Features

* add interactive guided setup wizard ([#27](https://github.com/omargallob/devops-starter/issues/27)) ([88c35ef](https://github.com/omargallob/devops-starter/commit/88c35efb7c6aeafd228585ae262113d89b1f9cdb))
* dynamically display mise-managed languages in CLI and TUI ([#22](https://github.com/omargallob/devops-starter/issues/22)) ([c685fc2](https://github.com/omargallob/devops-starter/commit/c685fc2dc650b5f700474464f4c59ca4f84cbd2b))

## [0.1.1](https://github.com/omargallob/devops-starter/compare/v0.1.0...v0.1.1) (2026-05-25)


### Features

* add bootstrap install script ([3a3731b](https://github.com/omargallob/devops-starter/commit/3a3731b9e8394099f6484e1d8a73d85b2a1b6293))
* add conventional commits enforcement via pre-commit and commitlint ([#7](https://github.com/omargallob/devops-starter/issues/7)) ([1c5cf1e](https://github.com/omargallob/devops-starter/commit/1c5cf1eede7c3bdbddc57322493fa962a71a03ef))
* add main entry point for devops-starter CLI ([fdd7a07](https://github.com/omargallob/devops-starter/commit/fdd7a0775f7235fe7778b2ab6cb0c503f14111b6))
* **bazel:** add golangci-lint test rule ([2f872d1](https://github.com/omargallob/devops-starter/commit/2f872d159983a467cc0b0b116998287dbf8ac47c))
* **cli:** add cobra command structure ([f8762a8](https://github.com/omargallob/devops-starter/commit/f8762a842ba61178ad1c8fa79b111a56d4ba32ab))
* **config:** add default configuration file ([718b057](https://github.com/omargallob/devops-starter/commit/718b0572e83534c2c3f69d0c8ae4a6942e1b8deb))
* **config:** add YAML configuration management ([68479f8](https://github.com/omargallob/devops-starter/commit/68479f8f624f2f176a1a6553bf26d3de5e6c48b1))
* **doctor:** evaluate .zshrc for PATH config and add --fix flag ([#13](https://github.com/omargallob/devops-starter/issues/13)) ([f0f8452](https://github.com/omargallob/devops-starter/commit/f0f845259655597d3181fd2e6ff9c25289fc0683))
* **dotfiles:** add opinionated shell, git, tmux, starship, and neovim configs ([8ea29e0](https://github.com/omargallob/devops-starter/commit/8ea29e098b5fe274f0476c375fb0ecfeb20a4012))
* **dotfiles:** add symlink manager with backup and dry-run ([82bd9bb](https://github.com/omargallob/devops-starter/commit/82bd9bbd19439f52abc0fc92c758726760ee034c))
* initialize Go module with core dependencies ([e22eb0f](https://github.com/omargallob/devops-starter/commit/e22eb0fefcd476cb23797d74589309456b37fc4e))
* install confirmation, remove command, and TUI adopt/remove support ([#12](https://github.com/omargallob/devops-starter/issues/12)) ([365d6ef](https://github.com/omargallob/devops-starter/commit/365d6eff46ae620d2ad9d442c437cb03736454d3))
* **installer:** add download, checksum, extract, and install logic ([2ce1962](https://github.com/omargallob/devops-starter/commit/2ce196208fadda57fd8d2c16101c10433b9f4b81))
* interactive status TUI with version detection ([#3](https://github.com/omargallob/devops-starter/issues/3)) ([d9a2e6b](https://github.com/omargallob/devops-starter/commit/d9a2e6b72076980d088c3a08de4219792b1ebb57))
* **platform:** add OS/arch/distro detection and system deps ([9d6a280](https://github.com/omargallob/devops-starter/commit/9d6a280c1c037c0ff54714e4031f871cfcabec07))
* **registry:** add complete tool registry with 60+ tools ([dbc5799](https://github.com/omargallob/devops-starter/commit/dbc57998f736bc708c1b7b5911759225285892b8))
* replace release workflow with release-please automation ([#17](https://github.com/omargallob/devops-starter/issues/17)) ([819c3b6](https://github.com/omargallob/devops-starter/commit/819c3b660ec01b21c3e311e5ac0849dd38f4d339))
* rewrite coverage badge generator in Go ([#9](https://github.com/omargallob/devops-starter/issues/9)) ([dd42adb](https://github.com/omargallob/devops-starter/commit/dd42adb87783d441ef4f7b9915ce611737487f68))
* **tooldef:** add core tool definition types ([e83e536](https://github.com/omargallob/devops-starter/commit/e83e53680bd33dd07e7e544f6fbf4e7590430211))
* use bazel lint targets in pre-commit and CI ([#16](https://github.com/omargallob/devops-starter/issues/16)) ([ff75156](https://github.com/omargallob/devops-starter/commit/ff751563bccbf4d22106b4e16127224e49cf8095))


### Bug Fixes

* **bazel:** resolve golangci_lint test failures ([#11](https://github.com/omargallob/devops-starter/issues/11)) ([1f91d11](https://github.com/omargallob/devops-starter/commit/1f91d114cd457459fb500da14a4f9c80a3ba8f81))
* **ci:** use eager image layer handling for rules_img pulls ([ef3dc9a](https://github.com/omargallob/devops-starter/commit/ef3dc9ad0153fbaf7ef4dc64f58f6b867e94e48f))
* **shell:** handle mkcd cd failures for shellcheck ([382057e](https://github.com/omargallob/devops-starter/commit/382057e3ffbc9422d7b502f8dcffd9a95e69b833))


### Code Refactoring

* **bazel:** organize into macros/ and rules/ subdirectories ([73910bd](https://github.com/omargallob/devops-starter/commit/73910bd9f302402a92023c83e545d8303b67be10))
* **bazel:** simplify BUILD files using custom macros ([6e272a6](https://github.com/omargallob/devops-starter/commit/6e272a641204e520e94a4a61f6e7543c8fcbdb0c))
