# HakStore Publishing Spec

HakStore is a personal pak store for NextUI-based TrimUI devices. This document covers everything needed to publish a new app or update an existing one.

---

## How It Works

1. Each app lives in its own GitHub repo with a `pak.json` at the root and a GitHub release containing the installable zip.
2. `storefront_base.json` in this repo lists each app by `id` and `repo_url`.
3. A GitHub Actions workflow (`build_storefront.yml`) fetches each app's `pak.json` and rebuilds `storefront.json` — the live catalog the app reads at runtime.
4. HakStore fetches the catalog from:
   `https://raw.githubusercontent.com/aaronearles/nextui-hak-store/refs/heads/main/storefront.json`

HakStore itself never needs updating when a new app is added or an existing app releases a new version.

---

## App Repo Requirements

### 1. `pak.json` at repo root

```json
{
  "name": "MyApp",
  "version": "v1.0.0",
  "type": "TOOL",
  "description": "Short description shown in the store.",
  "author": "Aaron Earles",
  "repo_url": "https://github.com/aaronearles/nextui-myapp",
  "release_filename": "MyApp.pak.zip",
  "platforms": ["tg5040"],
  "changelog": {
    "v1.0.0": "Initial release."
  }
}
```

**Field reference:**

| Field | Required | Notes |
|---|---|---|
| `name` | Yes | Determines install directory: `/mnt/SDCARD/Tools/{name}.pak/` |
| `version` | Yes | Must exactly match the GitHub release tag (e.g. `v1.0.0`) |
| `type` | Yes | `"TOOL"` → installs to `Tools/`, `"EMU"` → installs to `Emus/` |
| `description` | Yes | Shown in browse/info screens |
| `author` | Yes | Displayed in pak info |
| `repo_url` | Yes | Must point to **this repo** (the fork, not upstream) |
| `release_filename` | Yes | Exact filename of the zip asset in the GitHub release |
| `platforms` | Yes | Array: any of `"tg5040"`, `"tg5050"`, `"my355"` |
| `changelog` | Yes | Map of `"vX.Y.Z"` → description string |
| `screenshots` | No | Array of image URLs shown in pak info |
| `update_ignore` | No | Glob patterns for files/dirs to preserve on update (e.g. user config files) |
| `scripts` | No | See Scripts section below |

### 2. GitHub Release

- Tag must exactly match `pak.json`'s `version` field (e.g. `v1.0.0`)
- Must include an asset named exactly as `release_filename` (e.g. `MyApp.pak.zip`)
- The download URL HakStore constructs is: `{repo_url}/releases/download/{version}/{release_filename}`

### 3. Release Zip Structure

The zip is extracted directly into `{name}.pak/` on the device. Contents should be **flat at the zip root** (not nested in a subfolder):

```
MyApp.pak.zip
├── launch.sh        ← entry point, required for NextUI paks
├── myapp            ← compiled binary
└── res/             ← any resources
    └── ...
```

A typical `launch.sh`:
```sh
#!/bin/sh
SELF_DIR="$(dirname "$0")"
exec "$SELF_DIR/myapp"
```

---

## Adding a New App to HakStore

### Step 1 — Prepare the app repo

Ensure the app repo has:
- `pak.json` at root with `repo_url` pointing at the correct GitHub repo
- A GitHub release tagged to match `version`, with the zip asset

### Step 2 — Add an entry to `storefront_base.json`

```json
{
  "id": "uniqueId10c",
  "storefront_name": "My App",
  "repo_url": "https://github.com/aaronearles/nextui-myapp",
  "categories": ["Tools"]
}
```

**Fields:**

| Field | Notes |
|---|---|
| `id` | Unique 10-char alphanumeric string — never change this after publishing |
| `storefront_name` | Display name in the store (can differ from `name` in pak.json) |
| `repo_url` | GitHub repo URL — the builder fetches `pak.json` from this repo |
| `categories` | Array of category strings — must have at least one or the pak won't appear in Browse |
| `platforms` | Optional override — otherwise taken from pak.json |
| `large_pak` | Set `true` for large downloads to show a warning |
| `disabled` | Set `true` to hide from catalog without removing the entry |
| `previous_names` | Array of old `storefront_name` values — used to detect renamed paks |
| `previous_repo_urls` | Array of old `repo_url` values — used to detect migrated repos |

### Step 3 — Rebuild the catalog

```sh
git add storefront_base.json
git commit -m "Add MyApp to catalog"
git push
gh workflow run build_storefront.yml
```

The workflow fetches each app's `pak.json` and writes `storefront.json` to main. The device picks up the new entry next time HakStore fetches the catalog.

---

## Releasing an Update

1. Make changes in the app repo
2. Bump `version` in `pak.json` and add a changelog entry
3. Commit and push to main in the app repo
4. Create a GitHub release with the new tag and upload the zip asset
5. Trigger the storefront rebuild (step 3 above) — HakStore's catalog will then show the update

No changes to `storefront_base.json` are needed for version updates; the builder always pulls the latest `pak.json` from each repo.

---

## Experimental Paks

Add to `experimental_paks` in `storefront_base.json` instead of `paks`. These are hidden unless the user enables Experimental Mode in HakStore settings.

```json
{
  "experimental_paks": [
    {
      "id": "uniqueId10c",
      "storefront_name": "My Experimental App",
      "repo_url": "https://github.com/aaronearles/nextui-myapp",
      "categories": ["Tools"]
    }
  ]
}
```

---

## Update Ignore (Preserving User Data)

Files matching patterns in `update_ignore` are left untouched during updates. Use this for config files or user-generated data inside the `.pak` folder:

```json
{
  "update_ignore": ["config.json", "saves/*"]
}
```

---

## Scripts

Optional shell scripts that run after install, update, or uninstall. Paths are relative to the `.pak` folder on-device:

```json
{
  "scripts": {
    "post_install": {
      "path": "scripts/post_install.sh",
      "args": []
    },
    "post_update": {
      "path": "scripts/post_update.sh",
      "args": []
    },
    "post_uninstall": {
      "path": "scripts/post_uninstall.sh",
      "args": []
    }
  }
}
```

---

## Reference: HakStore Repo

| File | Purpose |
|---|---|
| `storefront_base.json` | Edit this to add/remove/disable catalog entries |
| `storefront.json` | Auto-generated — do not edit manually |
| `.github/workflows/build_storefront.yml` | Rebuilds `storefront.json` from `storefront_base.json` |
| `.github/workflows/build_pak.yml` | Builds and releases a new version of HakStore itself |
| `models/constants.go` | `StorefrontJsonURL` — the live catalog URL fetched by the app |
