# Windows Publish

## Targets

- `windows/amd64`
- `windows/arm64`

Output artifacts (in `publish/output/`):

- `Facade-<version>-windows-<arch>.zip` — portable build (`facade.exe` + `config.yaml`)
- `Facade-Setup-<version>-<arch>.exe` — NSIS installer (only when `makensis` is available)

## Runtime policy

Facade only uses Chromium-native proxy links (`direct://` / `http://` / `https://` / `socks5://`),
so Windows packages do **not** bundle browser cores or any external proxy engine binaries.

Writable app state (config, database, profiles, logs) is stored under
`%LOCALAPPDATA%\facade`, not inside the install directory.

## Commands

Run on a native Windows host from Git Bash / MSYS:

```bash
bash publish/windows/publish-windows.sh --arch amd64
bash publish/windows/publish-windows.sh --arch arm64
```

Options:

- `--version <ver>` override the version (default reads `wails.json`)
- `--skip-build` reuse an existing `build/bin/facade.exe`
- `--skip-installer` produce only the portable zip
- `--keep-staging` keep the staging directory

## Prerequisites

- Go (matching `go.mod`) and Node 20
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0`)
- NSIS in `PATH` for the installer step (`choco install nsis`); omit or use `--skip-installer` to build the zip only

## CI

`.github/workflows/publish-windows.yml` builds both architectures on native
Windows runners (`windows-latest` for amd64, `windows-11-arm` for arm64),
uploads the artifacts, and attaches them to the GitHub Release when a `v*` tag
is pushed. It can also be triggered manually via **Actions → Publish Windows
Packages → Run workflow**.

> `windows-11-arm` must be available to your repository/organization for the
> arm64 job to run; otherwise remove that matrix entry or build arm64 locally.
