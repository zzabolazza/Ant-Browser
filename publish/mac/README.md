# macOS Publish Plan

## Purpose

This document defines the macOS packaging plan for Ant Browser.

The goal is to turn the current codebase into a macOS build that can:

- build on a native macOS machine
- launch from `/Applications`
- keep user-writable state outside the `.app` bundle
- avoid breaking existing Windows and Linux packaging flows

Ant Browser only uses Chromium-native proxy links (`direct://` / `http://` / `https://` / `socks5://`),
so packages do **not** bundle any external proxy engine binaries.

## Current Entry Command

The initial internal-build script can be invoked on a real Mac with:

```bash
bash publish/mac/publish-mac.sh --arch arm64
```

For the first iteration, `arm64` is the recommended target.

If no physical Mac is available, use the manual GitHub Actions workflow:

```text
Actions -> Publish macOS Package -> Run workflow
```

The workflow builds the same `arm64` internal package on a macOS runner and uploads the generated files from `publish/output/` as an artifact. The optional `version` input overrides the package version for that run.

This is a plan document only. It does not mean macOS packaging is already implemented.

## Current Status

The repository already has:

- Windows packaging flow
- Linux packaging flow
- initial `publish/mac/publish-mac.sh` scaffold for internal test builds
- initial `publish/config.init.mac.yaml` template

The repository does not yet have:

- signing / notarization flow

The repository now includes the first macOS writable-state implementation for app bundle roots:

- when the app root is inside `.app/Contents/MacOS` or `.app/Contents/Resources`
- writable state is redirected to `~/Library/Application Support/ant-browser`
- config, chrome, and data move to the user state root

## Current Implementation Note

The current initial macOS packaging scaffold intentionally places seed files under:

- `Ant Browser.app/Contents/MacOS/config.yaml`
- `Ant Browser.app/Contents/MacOS/chrome/README.md`

This is not the prettiest final bundle layout, but it matches the current runtime path resolution and avoids a larger refactor in Phase 1.

After the first internal build is stable, the bundle layout can be reviewed and moved toward `Contents/Resources` if needed.

## Why macOS Looks More Complex

macOS is not difficult because of Wails alone. The real complexity comes from three areas:

1. Installed `.app` bundles under `/Applications` should be treated as read-only.
2. User data must not be written inside the `.app` bundle.
3. Public distribution usually requires code signing and notarization, otherwise Gatekeeper may block launch.

## Recommended Scope

### Phase 1: Internal Test Build

Target:

- `darwin/arm64` first
- output `.app` and `.zip`
- unsigned build is acceptable for internal testing

Why:

- Apple Silicon is the mainstream macOS target now
- it keeps the first version smaller and easier to verify
- it avoids spending time on Intel support before the runtime path is stable

### Phase 2: Public Distribution Build

Target:

- signed `.app`
- notarized `.zip` or `.dmg`
- optional `darwin/amd64` or universal build

Why:

- end users expect double-click install and normal launch
- unsigned apps are more likely to be blocked

## Recommended Runtime Layout

### App Bundle

Recommended structure inside the built app:

- `Ant Browser.app/Contents/MacOS/ant-chrome`
- optional placeholder `chrome/README.md` if you want to keep behavior aligned with Linux

### User-Writable State

Recommended macOS state root:

- `~/Library/Application Support/ant-browser`

Recommended contents under the state root:

- `config.yaml`
- `proxies.yaml`
- `data/`
- `chrome/`

Rule:

- config, database, browser cores, logs, and profile data stay in the user state root

## Code Changes Required

### 1. Add macOS Writable State Handling

Current Linux detached-state logic only activates on Linux:

- `backend/internal/apppath/apppath.go`

Required change:

- extend path detection so installed macOS apps also use a detached writable state root
- recommended trigger: when `GOOS=darwin` and app root is not writable, or when running from an `.app` bundle

Expected result:

- app launch from `/Applications` does not try to write config/data into the bundle

### 2. Add macOS Publish Script

New file to add:

- `publish/mac/publish-mac.sh`

Recommended responsibilities:

1. verify host is macOS
2. verify target arch (`arm64` first)
3. install frontend dependencies
4. build frontend
5. run `wails build -platform darwin/arm64`
6. optionally archive to `.zip`
7. optionally sign and notarize when environment variables are provided

Current scaffold status:

- implemented as an unsigned internal-build script
- outputs `.app` plus `.zip`
- requires a native macOS host
- intentionally does not attempt notarization yet

### 3. Signing and Notarization

This is not required for a first internal test build, but is required for a serious public release.

Needed later:

- Apple Developer certificate
- `codesign`
- `notarytool`
- entitlements if runtime behavior requires them

Typical flow:

1. sign the `.app`
2. zip or build dmg
3. notarize
4. staple

## Recommended Implementation Order

1. Deliver `darwin/arm64` internal test build only.
2. Add macOS detached state root.
3. Add `publish/mac/publish-mac.sh`.
4. Verify launch from `/Applications`.
5. Verify browser core placement under user state root.
6. Add signing and notarization only after the unsigned build is stable.
7. Decide whether `darwin/amd64` is worth supporting.

## Validation Checklist

The macOS work should not be considered complete until all items below are verified on a real Mac.

### Packaging

- build completes on native macOS
- output `.app` exists
- output `.zip` or `.dmg` exists

### First Launch

- app launches from Finder
- app launches after copying to `/Applications`
- first launch creates `~/Library/Application Support/ant-browser`
- `config.yaml` is seeded correctly
- database and `data/` are created under the user state root

### Browser Core

- manually placed browser core can be detected
- browser core path persists in config or database
- browser instance can actually start

### Exit Behavior

- window close works
- explicit quit works
- no stuck background process remains after quit

### Regression Safety

- Windows packaging still builds
- Linux packaging still builds
- Linux detached state behavior still works

## Difficulty Assessment

### Internal Test Build

Difficulty: medium

Main blockers:

- detached writable state
- mac packaging script

### Public Release Build

Difficulty: medium-high

Main blockers:

- signing
- notarization
- quarantine / Gatekeeper behavior

## Suggested First Deliverable

The safest first milestone is:

- macOS `arm64`
- native build on a real Mac
- unsigned `.app`
- zipped artifact for internal testing
- detached writable state under `~/Library/Application Support/ant-browser`

Do not start with:

- universal binary
- dmg beautification
- public distribution
- Intel support

Those can come after the app is proven stable on one Mac target first.

## Files Expected To Be Added Or Updated

Likely new files:

- `publish/mac/publish-mac.sh`
- `publish/mac/README.md`

Likely updated files:

- `backend/internal/apppath/apppath.go`
- `backend/internal/apppath/apppath_test.go`
- `backend/runtime_paths.go`
- `backend/app.go`
- `main.go`

## Decision Record

Current recommendation:

- do macOS `arm64` first
- solve writable state before touching signing
- keep Windows and Linux publish flows unchanged unless shared runtime code needs extension
- treat public notarized distribution as Phase 2, not Phase 1
