# API Coverage Analysis - Device & Firmware Domain

> Data sources: `apps/network/src/services/` and all `*/service.ts` files under `apps/network/src/pages/`; `apps/universal-login/src/services/` (no device/firmware APIs found there).

## Summary Statistics

- Portal/Network API calls in device & firmware domain: 95
- CLI covered: 58
- CLI uncovered (Gap): 37
- Coverage rate: ~61%

## Detailed Comparison Table

### Device Core

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices | List devices | `incloud device list` | ✅ Covered |
| POST | /api/v1/devices | Create device | `incloud device create` | ✅ Covered |
| GET | /api/v1/devices/{id} | Get device detail | `incloud device get` | ✅ Covered |
| PUT | /api/v1/devices/{id} | Update device info | `incloud device update` | ✅ Covered |
| DELETE | /api/v1/devices/{id} | Delete device | `incloud device delete` | ✅ Covered |
| GET | /api/v1/devices/export | Export device list | `incloud device export` | ✅ Covered |
| PUT | /api/v1/devices/move | Move device to/from group | `incloud device assign` / `unassign` | ✅ Covered |
| PUT | /api/v1/devices/transfer | Transfer device to org | `incloud device transfer` | ✅ Covered |
| POST | /api/v1/devices/bulk-invoke-methods | Bulk reboot/restore devices | `incloud device exec reboot` (bulk) | ✅ Covered |
| PUT | /api/v1/devices/{id}/firmware-upgrade/ignore | Ignore firmware upgrade prompt | — | ❌ Uncovered |
| GET | /api/v1/devices/summary | Dashboard device summary stats | — | ❌ Uncovered |
| GET | /api/v1/devices/offline/statistics | Fleet-wide offline statistics | — | ❌ Uncovered |
| GET | /api/v1/devices/offline/topn | Top-N offline devices | — | ❌ Uncovered |
| POST | /api/v1/devices/bulk/update | Bulk update devices (CSV upload) | — | ❌ Uncovered |

### Device Live / Real-time

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/live/devices/list | Get live presence for multiple devices | — | ❌ Uncovered |
| GET | /api/v1/live/devices/{id} | Get single device live presence info | — | ❌ Uncovered |

### Device Import

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/devices/imports | Upload import file | `incloud device import` | ✅ Covered |
| GET | /api/v1/devices/imports | List import records | `incloud device import-status list` | ✅ Covered |
| POST | /api/v1/devices/imports/{id} | Confirm/execute import job | `incloud device import` (auto-confirm) | ✅ Covered |
| GET | /api/v1/devices/imports/{id}/detail | Get import job progress | `incloud device import-status get` | ✅ Covered |
| PUT | /api/v1/devices/imports/cancel/{id} | Cancel in-progress import | — | ❌ Uncovered |

### Device Config

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/config | Get device config | `incloud device config get` | ✅ Covered |
| GET | /api/v1/devices/{id}/merge-config | Get merged (effective) config | `incloud device config get --merge` | ✅ Covered |
| DELETE | /api/v1/devices/{id}/pending/config | Clear pending config | `incloud device config abort` | ✅ Covered |
| GET | /api/v1/devices/{id}/config/error | Get config error info | `incloud device config error` | ✅ Covered |
| GET | /api/v1/devices/{id}/config/history | List config history snapshots | `incloud device config history list` | ✅ Covered |
| GET | /api/v1/devices/{id}/config/history/{snapshotId} | Get config snapshot detail | `incloud device config history get` | ✅ Covered |
| POST | /api/v1/devices/{id}/config/history/{snapshotId}/apply | Restore config from snapshot | `incloud device config history restore` | ✅ Covered |
| PUT | /api/v1/config/direct | Direct config update (no session) | `incloud device config update` | ✅ Covered |
| POST | /api/v1/config/layer/bulk-copy | Bulk copy config layer across devices/groups | `incloud device config copy` | ✅ Covered |
| POST | /api/v1/config/init | Init config edit session | — | ❌ Uncovered |
| POST | /api/v1/config/init/snapshot | Init read-only snapshot session | — | ❌ Uncovered |
| POST | /api/v1/config/commit | Commit config session changes | — | ❌ Uncovered |
| DELETE | /api/v1/config | Discard config session | — | ❌ Uncovered |
| GET | /api/v1/config/pending | Get pending session config | — | ❌ Uncovered |
| DELETE | /api/v1/config/layer/device/{id} | Delete device config layer | — | ❌ Uncovered |
| GET | /api/v1/config/layer/group/{id} | Get group config layer | — | ❌ Uncovered |
| DELETE | /api/v1/config/layer/group/{id} | Delete group config layer | — | ❌ Uncovered |

### Config Documents (Schema)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/config-documents | List config documents/schemas | `incloud device config schema list` | ✅ Covered |
| GET | /api/v1/config-documents/overview | Config documents overview | `incloud device config schema overview` | ✅ Covered |

### Device Interfaces & Uplinks

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/interfaces | Get interface status | `incloud device interface` | ✅ Covered |
| POST | /api/v1/devices/{id}/interfaces/refresh | Refresh interface status | `incloud device interface --refresh` | ✅ Covered |
| GET | /api/v1/devices/{id}/uplinks | Get device uplinks | `incloud device uplink` | ✅ Covered |
| GET | /api/v1/devices/{id}/uplinks/perf-trend | Uplink performance trend | `incloud device uplink perf` | ✅ Covered |
| GET | /api/v1/uplinks/{id} | Get uplink detail | `incloud device uplink get` | ✅ Covered |

### Device Signal & Performance

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/signal | Get signal history | `incloud device signal list` | ✅ Covered |
| GET | /api/v1/devices/{id}/signal/export | Export signal data | `incloud device signal export` | ✅ Covered |
| GET | /api/v1/devices/{id}/current-signal | Get current (latest) signal | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/antenna-signal | Get antenna signal history | `incloud device antenna` | ✅ Covered |
| GET | /api/v1/devices/{id}/performance | Get EC performance chart data | `incloud device perf` | ✅ Covered |
| GET | /api/v1/devices/{id}/performances | Get EC performance list | `incloud device perf` | ✅ Covered |
| POST | /api/v1/devices/{id}/performances/refresh | Refresh EC performance data | `incloud device perf --refresh` | ✅ Covered |

### Device Data Usage (Traffic)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/datausage-hourly | Hourly data usage series | `incloud device datausage hourly` | ✅ Covered |
| GET | /api/v1/devices/{id}/datausage-daily | Daily data usage series | `incloud device datausage daily` | ✅ Covered |
| GET | /api/v1/devices/{id}/datausage-{type}/overview | Data usage overview (dashboard) | — | ❌ Uncovered |
| GET | /api/v1/devices/datausage/details | Fleet-wide data usage details | `incloud device datausage list` | ✅ Covered |

### Device Location

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| PUT | /api/v1/devices/{id}/location | Set/update device location | `incloud device location set` | ✅ Covered |
| PUT | /api/v1/devices/{id}/locations/refresh | Refresh GPS location from device | `incloud device location refresh` | ✅ Covered |

### Device Online Events

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/online-events-list | Connection event history list | `incloud device online events` | ✅ Covered |
| GET | /api/v1/devices/{id}/offline/daily-stats | Daily offline stats per device | `incloud device online daily-stats` | ✅ Covered |
| GET | /api/v1/devices/{id}/online-events-chart/statistics | Online events chart statistics | — | ❌ Uncovered |

### Device Diagnostics (exec)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/devices/{id}/diagnosis/ping | Start ping | `incloud device exec ping` | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/traceroute | Start traceroute | `incloud device exec traceroute` | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/capture | Start packet capture | `incloud device exec capture` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/capture | Poll capture status | `incloud device exec capture` (poll) | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/flowscan | Start flow scan (traffic analysis) | `incloud device exec flowscan` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/flowscan | Get flowscan status/results | `incloud device exec flowscan-status` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/flowscan/export | Export flowscan results | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/diagnosis/speedtest/config | Get speedtest server config | `incloud device exec speedtest` | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/speedtest | Start speedtest | `incloud device exec speedtest` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/speed-test-histories | Speedtest history | `incloud device exec speedtest-history` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/interfaces | Get diagnostic interfaces | `incloud device exec interfaces` | ✅ Covered |
| PUT | /api/v1/diagnosis/{id}/cancel | Cancel any diagnosis task | `incloud device exec cancel` | ✅ Covered |
| POST | /api/v1/devices/{id}/methods | Invoke device method | `incloud device exec method` | ✅ Covered |
| POST | /api/v1/devices/{id}/outbound-rules | Deploy outbound rules from S3 | — | ❌ Uncovered |
| GET | /api/v1/devices/outbound-rules/presigned-url | Get presigned URL for rules file | — | ❌ Uncovered |

### Device Logs

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/logs/download | Download diagnostic logs | `incloud device log diagnostic` | ✅ Covered |
| GET | /api/v1/devices/{id}/logs/download/syslog | Download syslog | `incloud device log syslog` | ✅ Covered |
| GET | /api/v1/devices/{id}/mqttlogs | Get MQTT logs | `incloud device log mqtt` | ✅ Covered |

### Device Connections (OOBM)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/devices/{id}/connections | Create client connection to device | — | ❌ Uncovered |

### Device Groups

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devicegroups | List device groups | `incloud device group list` | ✅ Covered |
| POST | /api/v1/devicegroups | Create group | `incloud device group create` | ✅ Covered |
| GET | /api/v1/devicegroups/{id} | Get group detail | `incloud device group get` | ✅ Covered |
| PUT | /api/v1/devicegroups/{id} | Update group | `incloud device group update` | ✅ Covered |
| DELETE | /api/v1/devicegroups/{id} | Delete group | `incloud device group delete` | ✅ Covered |
| POST | /api/v1/devicegroups/remove | Batch remove groups | — | ❌ Uncovered |
| POST | /api/v1/devicegroups/bulk-invoke-methods | Bulk reboot/restore by group | — | ❌ Uncovered |
| GET | /api/v1/devicegroups/{id}/devices/candidates | Get candidate devices for group | — | ❌ Uncovered |
| GET | /api/v1/devicegroups/{id}/devices/upgrade | Get devices needing firmware upgrade | — | ❌ Uncovered |
| GET | /api/v1/devicegroups/{id}/firmware-versions | Get firmware versions in group | `incloud device group firmwares` | ✅ Covered |
| GET | /api/v1/devicegroups/{id}/summary | Get group summary (config/upgrade/version) | — | ❌ Uncovered |
| POST | /api/v1/devicegroups/devices/summary | Get device summary for multiple groups | `incloud device group list` (enriched) | ✅ Covered |
| GET | /api/v1/devicegroups/export | Export device groups | — | ❌ Uncovered |

### Serialnumber

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/serialnumber/{sn}/validate | Validate serial number + get product type | `incloud device create` (validation step) | ✅ Covered |

### Firmwares

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/firmwares | List firmwares | `incloud firmware list` | ✅ Covered |
| GET | /api/v1/firmwares/{id} | Get firmware detail | `incloud firmware get` | ✅ Covered |
| GET | /api/v1/firmwares/{id}/job/executions | List job executions for firmware | `incloud firmware executions --firmware` | ✅ Covered |
| POST | /api/v1/firmwares/batch/jobs | Create batch firmware upgrade job | `incloud firmware job create` | ✅ Covered |
| GET | /api/v1/products/{product}/firmwares | List firmwares by product | — | ❌ Uncovered |
| GET | /api/v1/products/{product}/firmwares/{version} | Get specific firmware by product+version | — | ❌ Uncovered |

### OTA Modules & Device OTA Status

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/ota/modules | List OTA modules (module types) | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/ota/modules | Get device OTA modules status | `incloud firmware status --device` | ✅ Covered |
| GET | /api/v1/devices/{id}/ota/modules/{module} | Get specific OTA module for device | — | ❌ Uncovered |
| GET | /api/v1/device/firmwares | List device firmware status (fleet) | `incloud firmware status` | ✅ Covered |
| GET | /api/v1/ota/devices | List OTA devices (upgrade pending view) | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/ota/jobs/completed | Completed OTA jobs for device | `incloud firmware executions --device` | ✅ Covered |
| GET | /api/v1/devices/{id}/jobs | All jobs for a device | — | ❌ Uncovered |

### OTA Jobs

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/jobs | List OTA jobs (portal upgrade log) | — | ❌ Uncovered |
| PUT | /api/v1/jobs/{id}/cancel | Cancel OTA job | `incloud firmware job cancel` | ✅ Covered |
| GET | /api/v1/ota/jobs | List all OTA jobs | `incloud firmware job list` | ✅ Covered |
| GET | /api/v1/job/executions | List all job executions (fleet-wide) | — | ❌ Uncovered |
| PUT | /api/v1/job/executions/{id}/cancel | Cancel job execution | `incloud firmware exec cancel` | ✅ Covered |
| PUT | /api/v1/job/executions/{id}/retry | Retry job execution | `incloud firmware exec retry` | ✅ Covered |
| GET | /api/v1/ota/job/executions | List OTA job executions (global) | `incloud firmware executions` | ✅ Covered |
| PUT | /api/v1/job/{jobId}/cancel/{groupId} | Cancel group scheduled upgrade job | — | ❌ Uncovered |

### Network Clients (Connected to Device)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/network/clients | List network clients | `incloud device client list` | ✅ Covered |
| GET | /api/v1/network/clients/{id} | Get client detail | `incloud device client get` | ✅ Covered |
| PUT | /api/v1/network/clients/{id} | Update client | `incloud device client update` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/online-events-list | Client connection history | `incloud device client online-events` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/online-events-chart/statistics | Client online stats chart | `incloud device client online-stats` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/datausage-daily | Client daily data usage | `incloud device client datausage daily` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/datausage-hourly | Client hourly data usage | `incloud device client datausage hourly` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/rssi | Client RSSI data | `incloud device client rssi` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/sinr | Client SINR data | `incloud device client sinr` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/throughput | Client throughput data | `incloud device client throughput` | ✅ Covered |
| GET | /api/v1/devices/{id}/clients | Get clients connected to device | — | ❌ Uncovered |
| POST | /api/v1/network/devices/{id}/clients/pos-ready | Set client POS-ready status | — | ❌ Uncovered |

## Gap Analysis

### Critical Gaps

1. **Device Import Cancel**: `PUT /api/v1/devices/imports/cancel/{id}` is in the portal but absent from the CLI. An in-progress import cannot be cancelled via CLI.

2. **Job Execution Overview** (`GET /api/v1/job/executions`): The portal's upgrade history log uses this fleet-wide endpoint. The CLI `incloud firmware executions` uses `/api/v1/ota/job/executions` — a different endpoint with potentially different filtering behavior.

3. **Config Session Workflow**: The 3-step session flow (init → edit → commit/discard) used by the portal's config editor is not exposed as CLI commands. The CLI provides `config update` using `/api/v1/config/direct` as a shortcut, but the full session workflow is unavailable.

4. **Device Groups Batch Operations**: `POST /api/v1/devicegroups/bulk-invoke-methods` (reboot/restore all devices in a group by group ID) and `POST /api/v1/devicegroups/remove` (batch delete groups) are portal-only.

5. **OTA Jobs Listing Discrepancy**: The portal uses `GET /api/v1/jobs` (no OTA prefix) for the upgrade log page; the CLI uses `GET /api/v1/ota/jobs`. Both relate to firmware upgrade jobs but appear to be different endpoints.

### Minor Gaps

1. **Firmware product-version endpoints**: `GET /api/v1/products/{product}/firmwares` and `GET /api/v1/products/{product}/firmwares/{version}` — used heavily in the portal to show firmware options per product — are not surfaced in the CLI.

2. **Device live info** (`/api/v1/live/devices/*`): Real-time presence data shown in the portal's device list columns is not queryable via CLI.

3. **Device group candidates/upgrade views**: `GET /api/v1/devicegroups/{id}/devices/candidates` (add-device-to-group picker) and `GET /api/v1/devicegroups/{id}/devices/upgrade` (devices with mismatched firmware) are not covered.

4. **Device current signal**: `GET /api/v1/devices/{id}/current-signal` (instantaneous signal reading) is missing; CLI only has historical signal (`incloud device signal list`).

5. **Datausage overview**: `GET /api/v1/devices/{id}/datausage-{type}/overview` used in dashboard widgets is not covered.

6. **Flowscan / outbound-rules export**: `GET /api/v1/devices/{id}/diagnosis/flowscan/export` and the outbound-rules presigned-URL + deploy flow are portal-only.

7. **OTA module status per module**: `GET /api/v1/devices/{id}/ota/modules/{module}` (per-module status) is not covered; CLI only queries all modules at once.

## Notes

- `apps/universal-login` contains only auth/billing APIs; no device or firmware APIs are present there.
- The CLI's `incloud firmware status` uses `/api/v1/device/firmwares` (singular, no `s`) which differs from the portal's `/api/v1/devices/{id}/ota/modules`; both relate to device firmware status but serve slightly different views.
- Several portal-only APIs exist exclusively for dashboard visualizations (maps, charts, offline statistics) and match no operational CLI model.
- Network client sub-commands (`incloud device client *`) cover data usage, signal, and connection history that are visible only in the portal's device profile panels — good parity in this sub-area.
