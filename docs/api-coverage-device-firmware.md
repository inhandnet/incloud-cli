# API Coverage Analysis - Device & Firmware Domain

## Summary Statistics

- Portal/Network APIs (device & firmware domain): 78
- CLI covered: 53
- CLI uncovered (Gap): 25
- Coverage rate: ~68%

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
| PUT | /api/v1/devices/move | Move device to group | `incloud device assign` / `unassign` | ✅ Covered |
| PUT | /api/v1/devices/transfer | Transfer device to org | `incloud device transfer` | ✅ Covered |
| POST | /api/v1/devices/bulk-invoke-methods | Bulk reboot/restore | `incloud device exec reboot` (bulk) | ✅ Covered |
| PUT | /api/v1/devices/{id}/firmware-upgrade/ignore | Ignore firmware upgrade prompt | — | ❌ Uncovered |
| GET | /api/v1/devices/summary | Dashboard device summary | — | ❌ Uncovered |
| GET | /api/v1/devices/locations | Map locations | — | ❌ Uncovered |
| GET | /api/v1/devices/offline/statistics | Offline stats | — | ❌ Uncovered |
| GET | /api/v1/devices/offline/topn | Offline top-N | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/offline/daily-stats | Device offline daily stats | `incloud device online daily-stats` | ✅ Covered |
| GET | /api/v1/devices/orgs-summary | Orgs device summary (console) | — | ❌ Uncovered |

### Device Import

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/devices/imports | Upload import file | `incloud device import` | ✅ Covered |
| GET | /api/v1/devices/imports | List import records | `incloud device import-status list` | ✅ Covered |
| POST | /api/v1/devices/imports/{id} | Confirm import | `incloud device import` (auto-confirm) | ✅ Covered |
| GET | /api/v1/devices/imports/{id}/detail | Get import progress | `incloud device import-status get` | ✅ Covered |
| PUT | /api/v1/devices/imports/cancel/{id} | Cancel import | — | ❌ Uncovered |

### Device Live / Real-time

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/live/devices/list | Get live info for multiple devices | — | ❌ Uncovered |
| GET | /api/v1/live/devices/{id} | Get device live info | — | ❌ Uncovered |

### Device Config

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/config | Get device config | `incloud device config get` | ✅ Covered |
| GET | /api/v1/devices/{id}/merge-config | Get merged config | `incloud device config get --merge` | ✅ Covered |
| DELETE | /api/v1/devices/{id}/pending/config | Clear pending config | `incloud device config abort` | ✅ Covered |
| GET | /api/v1/devices/{id}/config/error | Get config error | `incloud device config error` | ✅ Covered |
| GET | /api/v1/devices/{id}/config/history | List config history | `incloud device config history list` | ✅ Covered |
| GET | /api/v1/devices/{id}/config/history/{snapshotId} | Get config snapshot | `incloud device config history get` | ✅ Covered |
| POST | /api/v1/devices/{id}/config/history/{snapshotId}/apply | Restore config snapshot | `incloud device config history restore` | ✅ Covered |
| PUT | /api/v1/config/direct | Direct config update | `incloud device config update` | ✅ Covered |
| POST | /api/v1/config/init | Init config session | — | ❌ Uncovered |
| POST | /api/v1/config/init/snapshot | Init snapshot config session | — | ❌ Uncovered |
| POST | /api/v1/config/commit | Commit config session | — | ❌ Uncovered |
| GET | /api/v1/config | Get session config | — | ❌ Uncovered |
| GET | /api/v1/config/pending | Get pending config | — | ❌ Uncovered |
| POST | /api/v1/config/layer/bulk-copy | Bulk copy config layer | `incloud device config copy` | ✅ Covered |
| DELETE | /api/v1/config/layer/device/{id} | Delete device config layer | — | ❌ Uncovered |
| GET | /api/v1/config/layer/group/{id} | Get group config layer | — | ❌ Uncovered |
| DELETE | /api/v1/config/layer/group/{id} | Delete group config layer | — | ❌ Uncovered |
| GET | /api/v1/config/default | Get default config | — | ❌ Uncovered |
| PUT | /api/v1/config/default | Update default config | — | ❌ Uncovered |

### Config Documents (Schema)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/config-documents | List config documents | `incloud device config schema list` | ✅ Covered |
| POST | /api/v1/config-documents | Add config document | — | ❌ Uncovered |
| PUT | /api/v1/config-documents/{id} | Update config document | — | ❌ Uncovered |
| DELETE | /api/v1/config-documents/{id} | Delete config document | — | ❌ Uncovered |
| GET | /api/v1/config-documents/overview | Config documents overview | `incloud device config schema overview` | ✅ Covered |
| POST | /api/v1/config-documents/overview | Upsert config doc overview | — | ❌ Uncovered |
| GET | /api/v1/config-documents/export | Export config documents | — | ❌ Uncovered |
| POST | /api/v1/config-documents/import | Import config documents | — | ❌ Uncovered |

### Device Interfaces & Uplinks

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/interfaces | Get interfaces | `incloud device interface` | ✅ Covered |
| POST | /api/v1/devices/{id}/interfaces/refresh | Refresh interfaces | `incloud device interface --refresh` | ✅ Covered |
| GET | /api/v1/devices/{id}/uplinks | Get uplinks | `incloud device uplink` | ✅ Covered |
| GET | /api/v1/devices/{id}/uplinks/perf-trend | Uplink performance trend | `incloud device uplink perf` | ✅ Covered |
| GET | /api/v1/uplinks/{id} | Get uplink detail | `incloud device uplink get` | ✅ Covered |

### Device Signal & Performance

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/signal | Get signal history | `incloud device signal list` | ✅ Covered |
| GET | /api/v1/devices/{id}/current-signal | Get current signal | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/antenna-signal | Get antenna signal | `incloud device antenna` | ✅ Covered |
| GET | /api/v1/devices/{id}/performance | Get EC performance (chart) | `incloud device perf` | ✅ Covered |
| GET | /api/v1/devices/{id}/performances | Get EC performance list | `incloud device perf` | ✅ Covered |
| POST | /api/v1/devices/{id}/performances/refresh | Refresh EC performance | `incloud device perf --refresh` | ✅ Covered |
| GET | /api/v1/devices/{id}/signal/export | Export signal data | `incloud device signal export` | ✅ Covered |

### Device Data Usage (Traffic)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/datausage-hourly | Hourly data usage | `incloud device datausage hourly` | ✅ Covered |
| GET | /api/v1/devices/{id}/datausage-daily | Daily data usage | `incloud device datausage daily` | ✅ Covered |
| GET | /api/v1/devices/{id}/datausage-monthly | Monthly data usage | `incloud device datausage monthly` | ✅ Covered |
| GET | /api/v1/devices/{id}/datausage-hourly/overview | Hourly data usage overview | — | ❌ Uncovered |
| GET | /api/v1/devices/datausage/details | Fleet-wide data usage details | `incloud device datausage list` | ✅ Covered |

### Device Location

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| PUT | /api/v1/devices/{id}/location | Update location | `incloud device location set` | ✅ Covered |
| PUT | /api/v1/devices/{id}/locations/refresh | Refresh GPS location | `incloud device location refresh` | ✅ Covered |

### Device Online Events

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/online-events-list | Get connect history list | `incloud device online events` | ✅ Covered |
| GET | /api/v1/devices/{id}/online-events-chart/statistics | Online events chart stats | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/online-events-chart | Online events chart (console) | — | ❌ Uncovered |

### Device Diagnostics (exec)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/devices/{id}/diagnosis/ping | Start ping | `incloud device exec ping` | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/traceroute | Start traceroute | `incloud device exec traceroute` | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/capture | Start packet capture | `incloud device exec capture` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/capture | Get capture status | `incloud device exec capture` (poll) | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/flowscan | Start flow scan | `incloud device exec flowscan` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/flowscan | Get flowscan status | `incloud device exec flowscan-status` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/speedtest/config | Get speedtest config | `incloud device exec speedtest` | ✅ Covered |
| POST | /api/v1/devices/{id}/diagnosis/speedtest | Start speedtest | `incloud device exec speedtest` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/speed-test-histories | Speedtest history | `incloud device exec speedtest-history` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/interfaces | Get diagnostic interfaces | `incloud device exec interfaces` | ✅ Covered |
| PUT | /api/v1/diagnosis/{id}/cancel | Cancel diagnosis task | `incloud device exec cancel` | ✅ Covered |
| POST | /api/v1/devices/{id}/methods | Invoke device method | `incloud device exec method` | ✅ Covered |
| GET | /api/v1/devices/{id}/diagnosis/flowscan/export | Export flowscan results | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/outbound-rules | Get outbound rules | — | ❌ Uncovered |
| POST | /api/v1/devices/outbound-rules/presigned-url | Get presigned URL for rules | — | ❌ Uncovered |

### Device Logs

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/logs/download | Download diagnostic logs | `incloud device log diagnostic` | ✅ Covered |
| GET | /api/v1/devices/{id}/logs/download/syslog | Download syslog | `incloud device log syslog` | ✅ Covered |
| GET | /api/v1/devices/{id}/mqttlogs | Get MQTT logs | `incloud device log mqtt` | ✅ Covered |
| GET | /api/v1/devices/{id}/logs/local | Get local logs (console IoT) | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/cloudwatch/logs | Get CloudWatch logs | — | ❌ Uncovered |

### Device Shadow (IoT - console)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/shadow | Get device shadow | `incloud device shadow get` | ✅ Covered |
| POST | /api/v1/devices/{id}/shadow | Update shadow | `incloud device shadow update` | ✅ Covered |
| PUT | /api/v1/devices/{id}/shadow | Update shadow (alt) | — | ⚠️ Partially Covered |
| DELETE | /api/v1/devices/{id}/shadow | Delete shadow | `incloud device shadow delete` | ✅ Covered |
| GET | /api/v1/devices/{id}/shadow/names | List shadow names | `incloud device shadow list` | ✅ Covered |
| GET | /api/v1/devices/{id}/events | Get device events (IoT) | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/properties | Get device properties | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/groups | Get device groups (IoT) | — | ❌ Uncovered |
| POST | /api/v1/devices/{id}/groups | Add device to groups | — | ❌ Uncovered |
| POST | /api/v1/devices/{id}/groups/delete | Remove device from groups | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/methods | List device methods | — | ❌ Uncovered |

### Device Groups

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devicegroups | List device groups | `incloud device group list` | ✅ Covered |
| POST | /api/v1/devicegroups | Create group | `incloud device group create` | ✅ Covered |
| GET | /api/v1/devicegroups/{id} | Get group | `incloud device group get` | ✅ Covered |
| PUT | /api/v1/devicegroups/{id} | Update group | `incloud device group update` | ✅ Covered |
| DELETE | /api/v1/devicegroups/{id} | Delete group | `incloud device group delete` | ✅ Covered |
| POST | /api/v1/devicegroups/remove | Batch remove groups | — | ❌ Uncovered |
| POST | /api/v1/devicegroups/bulk-invoke-methods | Bulk reboot/restore group | — | ❌ Uncovered |
| GET | /api/v1/devicegroups/{id}/devices/candidates | Get candidate devices for group | — | ❌ Uncovered |
| GET | /api/v1/devicegroups/{id}/devices/upgrade | Get devices needing upgrade | — | ❌ Uncovered |
| GET | /api/v1/devicegroups/{id}/firmware-versions | Get group firmware versions | `incloud device group firmwares` | ✅ Covered |
| GET | /api/v1/devicegroups/{id}/summary | Get group summary | — | ❌ Uncovered |
| POST | /api/v1/devicegroups/devices/summary | Get device summary for groups | `incloud device group list` (enriched) | ✅ Covered |
| GET | /api/v1/devicegroups/export | Export device groups | — | ❌ Uncovered |

### Serialnumber

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/serialnumber/{sn}/validate | Validate serial number | `incloud device create` (validation step) | ✅ Covered |

### Firmwares

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/firmwares | List firmwares | `incloud firmware list` | ✅ Covered |
| POST | /api/v1/firmwares | Create firmware | — | ❌ Uncovered |
| GET | /api/v1/firmwares/{id} | Get firmware detail | `incloud firmware get` | ✅ Covered |
| PUT | /api/v1/firmwares/{id} | Update firmware | — | ❌ Uncovered |
| DELETE | /api/v1/firmwares/{id} | Delete firmware | — | ❌ Uncovered |
| GET | /api/v1/firmwares/{id}/stats | Get firmware stats | — | ❌ Uncovered |
| PUT | /api/v1/firmwares/{id}/publish | Publish firmware | — | ❌ Uncovered |
| PUT | /api/v1/firmwares/{id}/deprecate | Deprecate firmware | — | ❌ Uncovered |
| PUT | /api/v1/firmwares/{id}/latest | Mark as latest | — | ❌ Uncovered |
| PUT | /api/v1/firmwares/{id}/order | Reorder firmware | — | ❌ Uncovered |
| GET | /api/v1/firmwares/{id}/delta-packages | Get delta packages | — | ❌ Uncovered |
| GET | /api/v1/firmwares/{id}/delta-packages/{version} | Get specific delta package | — | ❌ Uncovered |
| GET | /api/v1/firmwares/{id}/full-package | Get full package | — | ❌ Uncovered |
| PUT | /api/v1/firmwares/{id}/full-package | Upload full package | — | ❌ Uncovered |
| POST | /api/v1/firmwares/presigned-upload-url | Get presigned upload URL | — | ❌ Uncovered |
| GET | /api/v1/firmwares/global-summary | Global firmware summary | — | ❌ Uncovered |
| POST | /api/v1/firmwares/batch/jobs | Create batch firmware upgrade job | `incloud firmware job create` | ✅ Covered |
| GET | /api/v1/firmwares/{id}/jobs | List jobs for firmware | — | ❌ Uncovered |
| GET | /api/v1/firmwares/{id}/job/executions | List job executions for firmware | `incloud firmware executions --firmware` | ✅ Covered |
| GET | /api/v1/products/{product}/firmwares | List firmwares by product | — | ❌ Uncovered |
| GET | /api/v1/products/{product}/firmwares/{version} | Get firmware by product+version | — | ❌ Uncovered |

### OTA Modules

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/ota/modules | List OTA modules | — | ❌ Uncovered |
| POST | /api/v1/ota/modules | Create OTA module | — | ❌ Uncovered |
| PUT | /api/v1/ota/modules/{id} | Update OTA module | — | ❌ Uncovered |
| DELETE | /api/v1/ota/modules/{id} | Delete OTA module | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/ota/modules | Get device OTA modules | `incloud firmware status --device` | ✅ Covered |
| GET | /api/v1/devices/{id}/ota/modules/{module} | Get specific OTA module for device | — | ❌ Uncovered |
| GET | /api/v1/device/firmwares | List device firmware status | `incloud firmware status` | ✅ Covered |
| GET | /api/v1/ota/devices | List OTA devices (upgrade view) | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/ota/jobs/completed | Get completed OTA jobs for device | `incloud firmware executions --device` | ✅ Covered |
| GET | /api/v1/devices/{id}/ota/jobs/next | Get next pending OTA job | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/jobs | List all jobs for device | — | ❌ Uncovered |
| POST | /api/v1/devices/{deviceId}/jobs/{jobId}/cancel | Cancel device job | — | ❌ Uncovered |

### OTA Jobs

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/jobs | List OTA jobs | — | ❌ Uncovered |
| PUT | /api/v1/jobs/{id}/cancel | Cancel OTA job | `incloud firmware job cancel` | ✅ Covered |
| GET | /api/v1/ota/jobs | List all OTA jobs (CLI uses this) | `incloud firmware job list` | ✅ Covered |
| GET | /api/v1/job/executions | List job executions | — | ❌ Uncovered |
| PUT | /api/v1/job/executions/{id}/cancel | Cancel job execution | `incloud firmware exec cancel` | ✅ Covered |
| PUT | /api/v1/job/executions/{id}/retry | Retry job execution | `incloud firmware exec retry` | ✅ Covered |
| GET | /api/v1/ota/job/executions | List OTA job executions (global) | `incloud firmware executions` | ✅ Covered |
| PUT | /api/v1/job/{jobId}/cancel/{groupId} | Cancel group scheduled job | — | ❌ Uncovered |

### Device Clients (Network clients)

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| GET | /api/v1/devices/{id}/clients | Get connected clients | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/connections | Get client connections | — | ❌ Uncovered |
| GET | /api/v1/network/clients | List network clients | `incloud device client list` | ✅ Covered |
| GET | /api/v1/network/clients/{id} | Get client detail | `incloud device client get` | ✅ Covered |
| PUT | /api/v1/network/clients/{id} | Update client | `incloud device client update` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/online-events-list | Client connection history | `incloud device client online-events` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/online-events-chart/statistics | Client online stats chart | `incloud device client online-stats` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/datausage-daily | Client daily usage | `incloud device client datausage daily` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/datausage-hourly | Client hourly usage | `incloud device client datausage hourly` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/rssi | Client RSSI data | `incloud device client rssi` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/sinr | Client SINR data | `incloud device client sinr` | ✅ Covered |
| GET | /api/v1/network/clients/{id}/throughput | Client throughput data | `incloud device client throughput` | ✅ Covered |

### Bulk Device Operations

| Method | API Path | Portal Usage | CLI Command | Status |
|--------|---------|-------------|-------------|--------|
| POST | /api/v1/devices/bulk/update | Bulk update device info (CSV) | — | ❌ Uncovered |
| GET | /api/v1/devices/{id}/clients | Get clients for device | — | ❌ Uncovered |
| GET | /api/v1/devices/reset-service-status | Reset service status | — | ❌ Uncovered |

## Gap Analysis

### Critical Gaps (Impact on Daily Operations)

1. **Firmware CRUD (create/update/delete/publish)**: The CLI has no commands to create, update, delete, publish, deprecate, or manage firmware packages. Operators must use the console portal for all firmware lifecycle management.

2. **OTA Module Management**: No CLI commands for creating/updating/deleting OTA module types. The portal's `apps/console` fully manages these, but the CLI can only query device OTA status.

3. **Job Execution Overview**: `GET /api/v1/job/executions` (fleet-wide upgrade history) is accessible via portal but not directly via CLI. The CLI has `incloud firmware executions` which uses `/api/v1/ota/job/executions` — a slightly different endpoint.

4. **Config Session Workflow**: The 3-step config session flow (init → edit → commit) used by the portal's config editor is not exposed as CLI commands. The CLI provides direct config update (`config update`) which calls `/api/v1/config/direct` instead.

5. **Device Groups Batch Operations**: `POST /api/v1/devicegroups/bulk-invoke-methods` (reboot/restore all devices in groups via group IDs) and `POST /api/v1/devicegroups/remove` (bulk delete groups) are not covered.

6. **Device Import Cancel**: `PUT /api/v1/devices/imports/cancel/{id}` is available in the portal but not in the CLI, making it impossible to cancel an in-progress import from the command line.

### Minor Gaps

1. **Firmware package upload flow**: `POST /api/v1/firmwares/presigned-upload-url` + package upload is console-only; needed for uploading new firmware binaries.

2. **Config Documents CRUD**: While the CLI can read config documents/schemas, it cannot create, update, delete, or import/export them.

3. **Device live info** (`/api/v1/live/devices/*`): Real-time device presence data used by the portal's device list is not surfaced in the CLI.

4. **Device groups candidate/upgrade views**: `GET /api/v1/devicegroups/{id}/devices/candidates` and `GET /api/v1/devicegroups/{id}/devices/upgrade` used to show devices eligible for group membership or firmware upgrade are not in CLI.

5. **Signal/data usage overview endpoints**: `GET /api/v1/devices/{id}/current-signal` and datausage overview (`/datausage-hourly/overview`) are not covered.

6. **IoT-specific device features** (console only): Device events, properties, methods list, CloudWatch logs, and device group membership management are not represented in the CLI.

## Notes

- CLI endpoint `/api/v1/ota/jobs` (used by `incloud firmware job list`) differs from `/api/v1/jobs` used by the portal, but they appear to return equivalent data.
- The CLI's `incloud firmware status` uses `/api/v1/device/firmwares` (singular) which is distinct from `/api/v1/devices/{id}/ota/modules` used by the portal; both relate to device firmware status.
- Several portal-only APIs exist exclusively for dashboard visualizations (maps, charts, offline statistics) and have no equivalent in CLI's operational model.
- Network client sub-commands (`incloud device client *`) cover data usage, signal, and connection history features that are only visible in the portal's device profile panels.
- The `apps/console` (IoT platform) has additional device APIs not covered in `apps/network` (e.g., shadow, events, properties, CloudWatch logs). These represent IoT-specific features separate from network device management.
