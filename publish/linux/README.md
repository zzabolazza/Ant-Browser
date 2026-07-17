# Linux Publish

## Targets

- `linux/amd64`
- `linux/arm64`

Output artifacts:

- `publish/output/AntBrowser-<version>-linux-<arch>.tar.gz`
- `publish/output/ant-browser_<version>_<arch>.deb`

## Runtime policy

Ant Browser only uses Chromium-native proxy links (`direct://` / `http://` / `https://` / `socks5://`),
so Linux packages no longer bundle any external proxy engine binaries.

## Commands

Single architecture:

```bash
bash publish/linux/publish-linux.sh --arch amd64
bash publish/linux/publish-linux.sh --arch arm64
```

From Windows wrapper (`bat\publish.bat L`), Linux publish is executed through Docker Desktop using `publish/linux/linux-builder.Dockerfile`.

Batch call:

```bash
bash publish/linux/publish-linux-all.sh
```

## Notes

- Linux packages do **not** include browser cores (`chrome/` is not bundled).
- Build on native architecture runner for stability.
- `.deb` installs app files under `/opt/ant-browser`.
- Linux packages keep an empty `chrome/` placeholder with `README.md`, but do **not** bundle browser core binaries.
- `.deb` registers an application launcher at `/usr/share/applications/ant-browser.desktop`.
- `.deb` installs standard Linux desktop icons under `/usr/share/icons/hicolor/*/apps/ant-browser.png` and `/usr/share/pixmaps/ant-browser.png`, so menus and launchers are more likely to pick up the app icon correctly.
- `.deb` bundles AppStream metadata under `/usr/share/metainfo/ant-browser.metainfo.xml`, which improves recognition in software centers and GUI `.deb` installers.
- On Debian/Ubuntu desktop environments that already support local `.deb` GUI installers, the package can usually be installed by double-clicking it; if the host has no GUI installer association, use `sudo apt install ./ant-browser_<version>_<arch>.deb`.
- Linux packages currently register the app in the desktop launcher/menu; they do not force-create a shortcut file on each user's desktop.
