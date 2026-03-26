# InCloud CLI Installation Guide

This document is designed for AI assistants (e.g. Claude Code) to read and execute.

## Steps

### 1. Detect Platform

Determine the OS and architecture of the current machine:

| OS      | Architecture | Binary Name             |
|---------|-------------|-------------------------|
| Linux   | x86_64      | `incloud-linux-amd64`   |
| Linux   | arm64       | `incloud-linux-arm64`   |
| macOS   | x86_64      | `incloud-darwin-amd64`  |
| macOS   | arm64       | `incloud-darwin-arm64`  |
| Windows | x86_64      | `incloud-windows-amd64` |

Run `uname -s` and `uname -m` to detect. On macOS, if `uname -m` returns `x86_64`, check `sysctl -n sysctl.proc_translated` — if it returns `1`, the shell is running under Rosetta 2 and the native architecture is `arm64`.

### 2. Download Binary and Checksums

Try **GitHub Releases** first. If it times out or fails, fall back to the **S3 mirror**.

**GitHub Releases (primary):**
```
https://github.com/inhandnet/incloud-cli/releases/latest/download/{BINARY_NAME}
https://github.com/inhandnet/incloud-cli/releases/latest/download/checksums.txt
```

**S3 mirror (fallback):**

First, fetch the latest version tag from the manifest:
```
https://incloud-cli-releases.s3.cn-north-1.amazonaws.com.cn/inhandnet/incloud-cli/manifest.yaml
```

The manifest is a YAML file. Read the `latest` field to get the version tag (e.g. `v0.2.0`), then download:
```
https://incloud-cli-releases.s3.cn-north-1.amazonaws.com.cn/inhandnet/incloud-cli/{TAG}/{BINARY_NAME}
https://incloud-cli-releases.s3.cn-north-1.amazonaws.com.cn/inhandnet/incloud-cli/{TAG}/checksums.txt
```

### 3. Verify Checksum

The `checksums.txt` file contains SHA256 checksums in the format:
```
<hash>  <filename>
```

Verify the downloaded binary:
- macOS: `shasum -a 256 <binary>`
- Linux: `sha256sum <binary>`

Compare the output hash with the corresponding entry in `checksums.txt`. **Do not proceed if the checksum does not match.**

### 4. Install

Make the binary executable and move it to the install path:

1. Try `/usr/local/bin/incloud` — if permission denied, use `sudo` (ask the user first)
2. If the user prefers no sudo, install to `~/.local/bin/incloud` instead (create the directory if needed, and remind the user to add `~/.local/bin` to their PATH if it's not already there)

```bash
chmod +x <binary>
mv <binary> /usr/local/bin/incloud
```

### 5. Verify

Run `incloud version` to confirm the installation succeeded.

### 6. Login

```bash
incloud auth login
```

This opens a browser for OAuth authorization. It works with zero configuration — defaults to the global region and creates a `default` context automatically.

Two production regions are available:

| Region | Short name | Domain           | Command                          |
|--------|-----------|------------------|----------------------------------|
| Global | `global`  | inhandcloud.com  | `incloud auth login` (default)   |
| China  | `cn`      | inhandcloud.cn   | `incloud auth login --host cn`   |

Ask the user which region they need. After login, verify with `incloud auth status`.
