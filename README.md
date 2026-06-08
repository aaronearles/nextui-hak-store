# HakStore

A personal hard fork of [Pak Store](https://github.com/LoveRetro/nextui-pak-store) for side-loading,
publishing, and testing my own hacky NextUI paks on my devices.

It works exactly like Pak Store, except it points at my own catalog
(`storefront.json`, generated from `storefront_base.json` and published to the
`main` branch) instead of the official one.

## Installing

1. Build and package: `task all`
2. Copy `build/Tools/<platform>/HakStore.pak` to `SD_ROOT/Tools/<platform>` on the device
   (or use `task adb` if the device is connected via ADB).
3. Launch `HakStore` from the `Tools` menu.

## Adding a pak to the catalog

1. Add an entry (`storefront_name`, `repo_url`, `categories`, etc.) to `storefront_base.json`.
2. Push to `main` — the `Build Storefront.json` workflow regenerates `storefront.json`
   and commits it back, which HakStore fetches via raw.githubusercontent.com.
3. Or run `go run app/storefront_builder.go` locally to regenerate it yourself.

## Development

See `taskfile.yml` for build/package/deploy/debug tasks (`task build`, `task package`,
`task adb`, `task debug`, etc.).
