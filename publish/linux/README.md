# Linux Publish

## Targets

- `linux/amd64`
- `linux/arm64`

Output artifacts:

- `publish/output/Facade-<version>-linux-<arch>.tar.gz`
- `publish/output/facade_<version>_<arch>.deb`

## Runtime policy

Facade only uses Chromium-native proxy links (`direct://` / `http://` / `https://` / `socks5://`),
so Linux packages no longer bundle any external proxy engine binaries.

## Commands

Single architecture:

```bash
bash publish/linux/publish-linux.sh --arch amd64
bash publish/linux/publish-linux.sh --arch arm64
```

Batch call:

```bash
bash publish/linux/publish-linux-all.sh
```

## Notes

- Linux packages do **not** include browser cores.
- Build on native architecture runner for stability.
- `.deb` installs app files under `/opt/facade`.
- Writable app state (config, database, profiles, logs) is stored under `~/.local/share/facade` (or `$XDG_DATA_HOME/facade`), not under `/opt/facade`.
- `.deb` registers an application launcher at `/usr/share/applications/facade.desktop`.
- `.deb` installs standard Linux desktop icons under `/usr/share/icons/hicolor/*/apps/facade.png` and `/usr/share/pixmaps/facade.png`, so menus and launchers are more likely to pick up the app icon correctly.
- `.deb` bundles AppStream metadata under `/usr/share/metainfo/facade.metainfo.xml`, which improves recognition in software centers and GUI `.deb` installers.
- Install the `.deb` with `sudo apt install ./facade_<version>_<arch>.deb` so apt pulls runtime Depends (including `libwebkit2gtk-4.1-0` when the package was built against WebKitGTK 4.1). Prefer this over bare `dpkg -i`, which does not download missing dependencies from apt repositories.
- GitHub Actions builds on Ubuntu 24.04 link `libwebkit2gtk-4.1.so.0`. If you see `libwebkit2gtk-4.1.so.0: cannot open shared object file`, install `libwebkit2gtk-4.1-0` or reinstall the `.deb` via `apt`.
- On Debian/Ubuntu desktop environments that already support local `.deb` GUI installers, the package can usually be installed by double-clicking it; if the host has no GUI installer association, use `sudo apt install ./facade_<version>_<arch>.deb`.
- Linux packages currently register the app in the desktop launcher/menu; they do not force-create a shortcut file on each user's desktop.
