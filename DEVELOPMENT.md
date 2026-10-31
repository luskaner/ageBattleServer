# Development

## Requirements

- Go 1.27 or newer.
- Git.
- A network environment in which the test listeners can bind to local ports.

The project is a standalone Go module:

```text
github.com/luskaner/ageBattleServer
```

## Setup

Clone the repository and work from its root:

```powershell
git clone https://github.com/luskaner/ageBattleServer.git
Set-Location ageBattleServer
go mod download
```

Do not commit credentials, private keys, local environment files, generated
binaries, or captured traffic. Keep environment-specific configuration outside
the repository.

## Build

Build all packages:

```powershell
$env:CGO_ENABLED = "0"
go build ./...
```

Build the command-line server:

```powershell
go build -o battleserver.exe ./cmd/battleserver
```

The other directories under `cmd/` contain supporting tools used during
development and testing.

## Test

Run the static checks and the regular test suite:

```powershell
$env:CGO_ENABLED = "0"
go vet ./...
go test ./... -count=1
```

Run the race detector for the server and session packages:

```powershell
$env:CGO_ENABLED = "1"
go test -race ./internal/server/ ./internal/session/ -count=1
```

The complete simulation suite can take substantially longer than the unit
tests because it replays recorded-duration scenarios. Use focused package or
test selectors when iterating locally, then run the full checks before
publishing a release.

## Code organization

- The root package exposes the embeddable server API.
- `cmd/battleserver` contains the executable entry point.
- `internal/server` owns listeners and server lifecycle.
- `internal/session` owns sessions, peers, and observer state.
- `internal/protocol` defines message encoding and decoding.
- `internal/sim` contains replay and integration test support.

Keep protocol behavior and public API changes covered by tests. Prefer small,
focused changes that preserve existing wire compatibility.

## Contributions

Before opening a pull request:

1. Run `gofmt` on changed Go files.
2. Run `go vet ./...`.
3. Run `go test ./... -count=1`.
4. Run the race tests when changing concurrent server or session code.
5. Review the diff for secrets, generated files, and unrelated changes.

Contributions are licensed under the terms of the
[GNU Affero General Public License v3.0](LICENSE).

## Release

Releases are created from semantic version tags such as `v1.0.0`. The GitHub
Actions workflow runs a Go generator to create the GoReleaser configuration,
then publishes a draft release. Windows 7 uses the legacy Go toolchain,
separately from modern Windows:

- Windows 7: `386`, `amd64`
- Modern Windows: `386`, `amd64`, `arm64`
- Linux: `386`, `amd64`, `arm64`, and `arm` (`GOARM` 5, 6)
- macOS: one universal binary archive

The `Taskfile.yml` release tasks invoke the Go tool
`tools/scripts/cmd/generateGoreleaserConfig.go`, which generates
`.goreleaser.yaml`. Set `GOROOT_LEGACY` to a Go 1.27.3 Windows 7-compatible
toolchain:

```powershell
$env:GOROOT_LEGACY = "C:\path\to\go-legacy-win7"
task release-no-sign
```

`task build` creates a local snapshot build. `task release` creates a signed
snapshot release, while `task release-no-sign` skips checksum signing.

For a signed release, configure `GPG_FINGERPRINT` and
`GPG_PASSPHRASE`, and make the signing key available to GPG before running:

```powershell
$env:GPG_FINGERPRINT = "your-key-fingerprint"
$env:GPG_PASSPHRASE = "your-key-passphrase"
task release
```

The tasks require GoReleaser 2.x. The generated artifacts are written to
`dist/`, matching the upstream release layout. Each archive includes
`LICENSE`, `README.md`, and `DEVELOPMENT.md`, and the release produces a
checksum file. Archive names follow the upstream layout, including the
OS names such as `win7`, `win10`, and `mac`, and architecture or
instruction-set suffixes.

## License

Copyright and license information is in [`LICENSE`](LICENSE). By contributing
to this repository, you agree that your contribution is provided under the
same license.
