#!/usr/bin/env python3
"""Update the S3 mirror manifest.yaml with a new release.

Usage:
    python3 scripts/update-s3-manifest.py \
        --tag v0.3.0 \
        --bin-dir bin \
        --notes-file /tmp/release_notes.md \
        --manifest manifest.yaml

If --manifest doesn't exist, a new one is created.
If the tag already exists in the manifest, it is replaced (re-publish).
"""

import argparse
import glob
import os
import sys
from datetime import datetime, timezone

# PyYAML is available on GitHub Actions ubuntu-latest runners
import yaml


def main():
    parser = argparse.ArgumentParser(description="Update S3 mirror manifest")
    parser.add_argument("--tag", required=True, help="Release tag (e.g. v0.3.0)")
    parser.add_argument("--bin-dir", required=True, help="Directory containing built binaries")
    parser.add_argument("--notes-file", default="", help="Path to release notes file")
    parser.add_argument("--prerelease", action="store_true", help="Mark as prerelease")
    parser.add_argument("--manifest", required=True, help="Path to manifest.yaml (read/write)")
    args = parser.parse_args()

    # Load existing manifest or start fresh
    manifest = {"last_release_id": 0, "last_asset_id": 0, "releases": []}
    if os.path.exists(args.manifest):
        with open(args.manifest) as f:
            loaded = yaml.safe_load(f)
            if loaded:
                manifest = loaded

    releases = manifest.get("releases") or []
    last_release_id = manifest.get("last_release_id", 0)
    last_asset_id = manifest.get("last_asset_id", 0)

    # Remove existing entry for this tag (allows re-publish)
    releases = [r for r in releases if r.get("tag_name") != args.tag]

    # Read release notes
    notes = ""
    if args.notes_file and os.path.exists(args.notes_file):
        with open(args.notes_file) as f:
            notes = f.read().strip()

    # Collect assets: binaries + checksums.txt
    asset_files = sorted(glob.glob(os.path.join(args.bin_dir, "incloud-*")))
    checksums = os.path.join(args.bin_dir, "checksums.txt")
    if os.path.exists(checksums):
        asset_files.append(checksums)

    if not asset_files:
        print(f"ERROR: no assets found in {args.bin_dir}", file=sys.stderr)
        sys.exit(1)

    assets = []
    for filepath in asset_files:
        last_asset_id += 1
        name = os.path.basename(filepath)
        assets.append({
            "id": last_asset_id,
            "name": name,
            "size": os.path.getsize(filepath),
            "url": f"{args.tag}/{name}",  # relative, HttpSource resolves it
        })

    last_release_id += 1
    release = {
        "id": last_release_id,
        "name": args.tag,
        "tag_name": args.tag,
        "url": "",
        "draft": False,
        "prerelease": args.prerelease,
        "published_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "release_notes": notes,
        "assets": assets,
    }

    releases.append(release)

    # Keep only the last 20 releases to avoid unbounded manifest growth
    if len(releases) > 20:
        releases = releases[-20:]

    manifest = {
        "last_release_id": last_release_id,
        "last_asset_id": last_asset_id,
        "releases": releases,
    }

    with open(args.manifest, "w") as f:
        yaml.dump(manifest, f, default_flow_style=False, allow_unicode=True, sort_keys=False)

    print(f"Manifest updated: {args.tag} ({len(assets)} assets)")


if __name__ == "__main__":
    main()
