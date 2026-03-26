# API Coverage Analysis - Admin & Users Domain

## Summary Stats

- Total unique API endpoints in Portal (admin/users domain): 102
- CLI covered: 26
- CLI not covered (Gap): 76
- Coverage rate: ~25%

## Detailed Comparison Table

### User Management (`/api/v1/users*`, `/api/v1/user/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/users | List users (console/universal-login) | `incloud user list` | ✅ Covered |
| GET | /api/v1/users/{id} | Get user detail | `incloud user get <id>` | ✅ Covered |
| POST | /api/v1/users | Create user | `incloud user create` | ✅ Covered |
| PUT | /api/v1/users/{id} | Update user | `incloud user update <id>` | ✅ Covered |
| DELETE | /api/v1/users/{id} | Delete user | `incloud user delete <id>` | ✅ Covered |
| PUT | /api/v1/users/{id}/lock | Lock user | `incloud user lock <id>` | ✅ Covered |
| PUT | /api/v1/users/{id}/unlock | Unlock user | `incloud user unlock <id>` | ✅ Covered |
| GET | /api/v1/users/me | Get current user profile | `incloud user me` | ✅ Covered |
| PUT | /api/v1/users/me | Update current user | — | ❌ Not covered |
| GET | /api/v1/user/identities | List user identities across orgs | `incloud user identity list` | ✅ Covered |
| GET | /api/v1/users/{id}/roles | Get user roles | — | ❌ Not covered |
| POST | /api/v1/users/{id}/roles | Assign roles to user | — | ❌ Not covered |
| DELETE | /api/v1/users/{id}/roles | Remove roles from user | — | ❌ Not covered |
| PUT | /api/v1/users/{id}/password | Set user password | — | ❌ Not covered |
| GET | /api/v1/users/{id}/groups | Get user groups | — | ❌ Not covered |
| PUT | /api/v1/users/{id}/groups | Add user to groups | — | ❌ Not covered |
| PUT | /api/v1/users/{id}/groups/delete | Remove user from groups | — | ❌ Not covered |
| DELETE | /api/v1/users/{id}/mfa/{type}/associate | Delete MFA association | — | ❌ Not covered |
| POST | /api/v1/users/invite | Invite external user | — | ❌ Not covered |
| POST | /api/v1/users/invitations | Get invitation link | — | ❌ Not covered |
| PUT | /api/v1/users/invitations/reset | Reset invitation link | — | ❌ Not covered |
| POST | /api/v1/users/remove | Bulk remove users | — | ❌ Not covered |
| GET | /api/v1/users/{id}/resend-invite | Resend invite email | — | ❌ Not covered |
| POST | /api/v1/users/{id}/impersonate/login | Impersonate user | — | ❌ Not covered |
| POST | /api/v1/users/impersonate/logout | Stop impersonation | — | ❌ Not covered |
| PUT | /api/v1/users/me/locale | Update user locale | — | ❌ Not covered |
| PUT | /api/v1/users/me/password | Update own password | — | ❌ Not covered |
| POST | /api/v1/users/bind-phone/verification | Send phone bind code | — | ❌ Not covered |
| POST | /api/v1/users/unbind-phone/verification | Send phone unbind code | — | ❌ Not covered |
| POST | /api/v1/users/bind-phone | Bind phone number | — | ❌ Not covered |
| POST | /api/v1/users/unbind-phone | Unbind phone number | — | ❌ Not covered |
| POST | /api/v1/users/orgs-summary | Get org user summary | — | ❌ Not covered |
| POST | /api/v1/users/list | List users (POST search) | — | ❌ Not covered |
| GET | /api/v1/users/passkeys | List passkeys | — | ❌ Not covered |
| POST | /api/v1/users/passkeys/register/start | Start passkey registration | — | ❌ Not covered |
| POST | /api/v1/users/passkeys/register | Register passkey | — | ❌ Not covered |
| PUT | /api/v1/users/passkeys/{id} | Update passkey name | — | ❌ Not covered |
| GET | /api/v1/user/features | List feature compatibilities | — | ❌ Not covered |
| POST | /api/v1/user/features | Create feature compatibility | — | ❌ Not covered |
| PUT | /api/v1/user/features/{id} | Update feature compatibility | — | ❌ Not covered |
| DELETE | /api/v1/user/features/{id} | Delete feature compatibility | — | ❌ Not covered |
| GET | /api/v1/user/features/{id} | Get feature compatibility | — | ❌ Not covered |

### Organization Management (`/api/v1/orgs/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/orgs | List organizations | `incloud org list` | ✅ Covered |
| GET | /api/v1/orgs/{id} | Get organization detail | `incloud org get <id>` | ✅ Covered |
| POST | /api/v1/orgs | Create organization | `incloud org create` | ✅ Covered |
| PUT | /api/v1/orgs/{id} | Update organization | `incloud org update <id>` | ✅ Covered |
| DELETE | /api/v1/orgs/{id} | Delete organization | `incloud org delete <id>` | ✅ Covered |
| GET | /api/v1/orgs/self | Get current org info | `incloud org self` | ✅ Covered |
| PUT | /api/v1/orgs/self | Update current org | `incloud org update-self` | ✅ Covered |
| POST | /api/v1/orgs/{id}/branding | Set org branding/customization | — | ❌ Not covered |
| GET | /api/v1/orgs/{id}/branding | Get org branding | — | ❌ Not covered |
| PUT | /api/v1/orgs/{id}/accessible | Update org subscription | — | ❌ Not covered |
| GET | /api/v1/orgs/{id}/apps | List org SaaS apps | — | ❌ Not covered |
| PUT | /api/v1/orgs/{id}/apps | Enable/disable SaaS app | — | ❌ Not covered |
| GET | /api/v1/orgs/{id}/addresses | List org addresses | — | ❌ Not covered |
| POST | /api/v1/orgs/{id}/addresses | Create org address | — | ❌ Not covered |
| DELETE | /api/v1/orgs/{id}/addresses/{id} | Delete org address | — | ❌ Not covered |
| PUT | /api/v1/orgs/{id}/addresses/{id} | Update org address | — | ❌ Not covered |
| GET | /api/v1/orgs/{id}/contacts | List org contacts | — | ❌ Not covered |
| POST | /api/v1/orgs/{id}/contacts | Add org contact | — | ❌ Not covered |
| PUT | /api/v1/orgs/{id}/contacts/{id} | Update org contact | — | ❌ Not covered |
| DELETE | /api/v1/orgs/{id}/contacts/{id} | Delete org contact | — | ❌ Not covered |
| PUT | /api/v1/orgs/{id}/bill-address | Update billing address | — | ❌ Not covered |
| GET | /api/v1/orgs/{id}/billing-policy | Get billing policy | — | ❌ Not covered |
| GET | /api/v1/orgs/{id}/users/{id} | Remove user from org (bulk) | — | ❌ Not covered |
| GET | /api/v1/orgs/{id}/admin-users | Get org admin users | — | ❌ Not covered |
| GET | /api/v1/orgs/global-summary | Get org global summary | — | ❌ Not covered |

### Role Management (`/api/v1/roles/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/roles | List roles | `incloud role list` | ✅ Covered |
| GET | /api/v1/roles/{id} | Get role detail | — | ❌ Not covered |
| POST | /api/v1/roles | Create role | — | ❌ Not covered |
| PUT | /api/v1/roles/{id} | Update role | — | ❌ Not covered |
| DELETE | /api/v1/roles/{id} | Delete role | — | ❌ Not covered |
| GET | /api/v1/roles/{id}/permissions | Get role permissions | — | ❌ Not covered |
| POST | /api/v1/roles/{id}/permissions | Add permission to role | — | ❌ Not covered |
| DELETE | /api/v1/roles/{id}/permissions | Remove permission from role | — | ❌ Not covered |
| GET | /api/v1/roles/{id}/users | Get users in role | — | ❌ Not covered |
| POST | /api/v1/roles/{id}/users | Assign role to users | — | ❌ Not covered |
| DELETE | /api/v1/roles/{id}/users | Remove role from users | — | ❌ Not covered |

### Product Management (`/api/v1/products/*`, `/api/v1/product-types/*`, `/api/v1/product-compatibilities/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/products | List products | `incloud product list` | ✅ Covered |
| GET | /api/v1/products/{id} | Get product detail | `incloud product get <id>` | ✅ Covered |
| POST | /api/v1/products | Create product | `incloud product create` | ✅ Covered |
| PUT | /api/v1/products/{id} | Update product | `incloud product update <id>` | ✅ Covered |
| DELETE | /api/v1/products/{id} | Delete product | `incloud product delete <id>` | ✅ Covered |
| GET | /api/v1/products/serialNumber-patterns | List SN patterns | — | ❌ Not covered |
| POST | /api/v1/products/serialNumber-patterns | Add SN pattern | — | ❌ Not covered |
| PUT | /api/v1/products/serialNumber-patterns/{id} | Update SN pattern | — | ❌ Not covered |
| DELETE | /api/v1/products/serialNumber-patterns/{id} | Delete SN pattern | — | ❌ Not covered |
| GET | /api/v1/products/edge/models | List edge AI models | — | ❌ Not covered |
| GET | /api/v1/product-types | List product types | — | ❌ Not covered |
| POST | /api/v1/product-types | Create product type | — | ❌ Not covered |
| PUT | /api/v1/product-types/{id} | Update product type | — | ❌ Not covered |
| DELETE | /api/v1/product-types/{id} | Delete product type | — | ❌ Not covered |
| GET | /api/v1/product-types/{id}/nm-micro-app | Get product type micro-app | — | ❌ Not covered |
| PUT | /api/v1/product-types/{id}/nm-micro-app | Update product type micro-app | — | ❌ Not covered |
| DELETE | /api/v1/product-types/{id}/nm-micro-app | Delete product type micro-app | — | ❌ Not covered |
| GET | /api/v1/product-compatibilities | List product compatibilities | — | ❌ Not covered |
| POST | /api/v1/product-compatibilities | Create product compatibility | — | ❌ Not covered |
| PUT | /api/v1/product-compatibilities/{id} | Update product compatibility | — | ❌ Not covered |
| DELETE | /api/v1/product-compatibilities/{id} | Delete product compatibility | — | ❌ Not covered |
| GET | /api/v1/product-compatibilities/{id} | Get product compatibility | — | ❌ Not covered |
| GET | /api/v1/product-compatibilities/{id}/products | Get products for compatibility | — | ❌ Not covered |

### Billing, Licensing & Orders (`/api/v1/billing/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/billing/licenses | List licenses | — | ❌ Not covered |
| POST | /api/v1/billing/licenses | Create license | — | ❌ Not covered |
| DELETE | /api/v1/billing/licenses/{id} | Delete license | — | ❌ Not covered |
| GET | /api/v1/billing/licenses/status-summary | License status stats | — | ❌ Not covered |
| PUT | /api/v1/billing/licenses/align | Align license dates | — | ❌ Not covered |
| POST | /api/v1/billing/licenses/move | Transfer licenses | — | ❌ Not covered |
| GET | /api/v1/billing/licenses/{id}/history | License change log | — | ❌ Not covered |
| GET | /api/v1/billing/license-types | List license types | — | ❌ Not covered |
| GET | /api/v1/billing/license-types/{slug}/prices | Get license type prices | — | ❌ Not covered |
| GET | /api/v1/billing/subscriptions | List subscriptions | — | ❌ Not covered |
| POST | /api/v1/billing/subscriptions | Create subscription | — | ❌ Not covered |
| DELETE | /api/v1/billing/subscriptions/{id} | Delete subscription | — | ❌ Not covered |
| GET | /api/v1/billing/subscriptions/{id} | Get subscription detail | — | ❌ Not covered |
| GET | /api/v1/billing/services | List billing services | — | ❌ Not covered |
| GET | /api/v1/billing/prices | List pricing | — | ❌ Not covered |
| POST | /api/v1/billing/invoice-info | Update invoice info | — | ❌ Not covered |
| GET | /api/v1/billing/invoice-info | Get invoice info | — | ❌ Not covered |

### SIM / Link Orders (`/api/v1/link/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| PUT | /api/v1/link/cards/imports | Import SIM cards | — | ❌ Not covered |
| GET | /api/v1/link/coupons | List SIM coupons | — | ❌ Not covered |
| POST | /api/v1/link/coupons | Create coupon | — | ❌ Not covered |
| PUT | /api/v1/link/coupons/{id}/close | Close coupon | — | ❌ Not covered |
| GET | /api/v1/link/coupons/{id} | Get coupon detail | — | ❌ Not covered |
| GET | /api/v1/link/coupons/{id}/plans | Get coupon plans | — | ❌ Not covered |
| GET | /api/v1/link/plans | List SIM plans | — | ❌ Not covered |
| POST | /api/v1/link/plans/applyCoupon | Apply coupon to plans | — | ❌ Not covered |
| GET | /api/v1/link/plan-types | List plan types | — | ❌ Not covered |
| GET | /api/v1/link/products | List SIM products | — | ❌ Not covered |
| GET | /api/v1/link/carriers | List carriers | — | ❌ Not covered |
| GET | /api/v1/link/orders | List SIM orders | — | ❌ Not covered |
| POST | /api/v1/link/orders | Create SIM order | — | ❌ Not covered |
| GET | /api/v1/link/orders/{id} | Get order detail | — | ❌ Not covered |
| PUT | /api/v1/link/orders/{id}/cancel | Cancel order | — | ❌ Not covered |
| GET | /api/v1/link/orders/coupons | Get order coupons | — | ❌ Not covered |
| GET | /api/v1/link/payment/orders | List prepaid orders | — | ❌ Not covered |
| GET | /api/v1/link/payment/orders/{id} | Get prepaid order | — | ❌ Not covered |
| POST | /api/v1/link/payment/orders/{id}/apply-invoice | Apply invoice to order | — | ❌ Not covered |
| GET | /api/v1/link/invoices | List invoices | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id} | Get invoice detail | — | ❌ Not covered |
| PUT | /api/v1/link/invoices/{id}/status | Update invoice status | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id}/download | Download invoice | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id}/receipt | Download receipt | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id}/checkout/stripe | Stripe checkout | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id}/payment-link | External payment link | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id}/sim-usage | Invoice SIM usage | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id}/transactions | Invoice transactions | — | ❌ Not covered |
| GET | /api/v1/link/invoices/{id}/transactions/{tradeNo} | Transaction detail | — | ❌ Not covered |
| GET | /api/v1/link/payment-links/{code}/checkout/stripe | Payment link stripe checkout | — | ❌ Not covered |
| GET | /api/v1/link/payment-links/{code}/invoice | Payment link invoice | — | ❌ Not covered |
| GET | /api/v1/link/payment-links/{code}/transactions/{tradeNo} | Payment link transaction | — | ❌ Not covered |

### Audit Logs (`/api/v1/audit/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/audit/logs | List audit logs | `incloud activity list` | ✅ Covered |
| GET | /api/v1/audit/logs/list | Export audit logs (list) | — | ⚠️ Partial (same endpoint, different alias) |

### Reports (`/api/v1/incloud/report*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/incloud/reports | List reports | — | ❌ Not covered |
| DELETE | /api/v1/incloud/reports/{id} | Delete report | — | ❌ Not covered |
| GET | /api/v1/incloud/reports/{id}/download | Download report | — | ❌ Not covered |
| POST | /api/v1/incloud/reports/{id}/recreate | Re-run report | — | ❌ Not covered |
| GET | /api/v1/incloud/report/policies | List report policies | — | ❌ Not covered |
| POST | /api/v1/incloud/report/policies | Create report policy | — | ❌ Not covered |
| PUT | /api/v1/incloud/report/policies/{id} | Update report policy | — | ❌ Not covered |
| DELETE | /api/v1/incloud/report/policies/{id} | Delete report policy | — | ❌ Not covered |
| GET | /api/v1/incloud/report/policies/{id} | Get report policy | — | ❌ Not covered |

### AI / Copilot (`/api/v1/ai/*`, `/api/v1/copilot/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| POST | /api/v1/copilot/chat | Copilot streaming chat | — | ❌ Not covered |
| POST | /api/v1/copilot/cancel | Cancel copilot request | — | ❌ Not covered |
| POST | /api/v1/copilot/score | Score copilot response | — | ❌ Not covered |
| GET | /api/v1/copilot/sessions | List chat sessions | — | ❌ Not covered |
| GET | /api/v1/copilot/sessions/{id} | Get session detail | — | ❌ Not covered |
| DELETE | /api/v1/copilot/sessions/{id} | Delete chat session | — | ❌ Not covered |
| POST | /api/v1/ai/description/generate | AI generate description | — | ❌ Not covered |
| POST | /api/v1/ai/description/optimize | AI optimize description | — | ❌ Not covered |
| POST | /api/v1/ai/description/calculate | AI calculate score | — | ❌ Not covered |
| POST | /api/v1/ai/config-explain | AI explain device config | — | ❌ Not covered |
| GET | /api/v1/ai/config/{sessionId} | AI config session | — | ❌ Not covered |
| PUT | /api/v1/ai/config/{sessionId} | AI config session update | — | ❌ Not covered |
| GET | /api/v1/ai/config/{sessionId}/is_supported | AI config support check | — | ❌ Not covered |

### Tokens & OAuth Clients (`/api/v1/tokens*`, `/api/v1/clients*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/tokens | List API tokens | — | ❌ Not covered |
| POST | /api/v1/tokens | Create API token | — | ❌ Not covered |
| PUT | /api/v1/tokens/{id} | Update API token | — | ❌ Not covered |
| DELETE | /api/v1/tokens/{id} | Delete API token | — | ❌ Not covered |
| GET | /api/v1/clients | List OAuth clients | — | ❌ Not covered |
| POST | /api/v1/clients | Create OAuth client | — | ❌ Not covered |
| GET | /api/v1/clients/{id} | Get OAuth client | — | ❌ Not covered |
| PUT | /api/v1/clients/{id} | Update OAuth client | — | ❌ Not covered |
| DELETE | /api/v1/clients/{id} | Delete OAuth client | — | ❌ Not covered |
| PUT | /api/v1/clients/{id}/secret | Reset client secret | — | ❌ Not covered |

### Stats & Data Usage (`/api/v1/stats/*`, `/api/v1/datausage/*`, `/api/v1/uplinks/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/datausage/overview | Traffic usage overview | `incloud overview traffic` | ✅ Covered |
| GET | /api/v1/datausage/topk | Top-K devices by data usage | `incloud overview traffic` | ✅ Covered |
| GET | /api/v1/uplinks/{id} | Get uplink detail | — | ⚠️ Partial (in `device uplink-get`, not admin module) |
| GET | /api/v1/stats/{name}/data | Get metric data series | — | ❌ Not covered |
| GET | /api/v1/stats/bulk-data | Bulk metric data | — | ❌ Not covered |
| GET | /api/v1/stats/metrics | List custom metrics | — | ❌ Not covered |
| POST | /api/v1/stats/metrics | Create custom metric | — | ❌ Not covered |
| PUT | /api/v1/stats/metrics/{id} | Update custom metric | — | ❌ Not covered |
| DELETE | /api/v1/stats/metrics/{id} | Delete custom metric | — | ❌ Not covered |

### System Messages (`/api/v1/system/messages*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/system/messages | List system messages | — | ❌ Not covered |
| POST | /api/v1/system/messages/confirm | Mark messages as read | — | ❌ Not covered |
| GET | /api/v1/system/messages/{id} | Get message detail | — | ❌ Not covered |

### Alert Notification Policies (`/api/v1/alert/notify/policies*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/alert/notify/policies | List notification policies | — | ❌ Not covered |
| POST | /api/v1/alert/notify/policies | Create notification policy | — | ❌ Not covered |
| GET | /api/v1/alert/notify/policies/{id} | Get policy detail | — | ❌ Not covered |
| PUT | /api/v1/alert/notify/policies/{id} | Update policy | — | ❌ Not covered |
| DELETE | /api/v1/alert/notify/policies/{id} | Delete policy | — | ❌ Not covered |
| PUT | /api/v1/alert/notify/policies/{id}/enable | Enable policy | — | ❌ Not covered |
| PUT | /api/v1/alert/notify/policies/{id}/disable | Disable policy | — | ❌ Not covered |

### Frontend Settings & Gateway (`/api/v1/frontend/settings`, `/api/v1/gateway/*`)

| Method | API Path | Frontend Purpose | CLI Command | Status |
|--------|---------|-----------------|-------------|--------|
| GET | /api/v1/frontend/settings | Get frontend settings | — | ⚠️ Partial (used internally for auth detection, not exposed as user command) |
| GET | /api/v1/gateway/scopes | List available OAuth scopes | — | ❌ Not covered |

---

## Gap Analysis

### Critical Gaps (impact daily admin operations)

1. **Role management (CRUD)** — Portal exposes full role lifecycle (create/update/delete/permissions). CLI only has `role list`. Admins managing platform roles must use the web UI.
2. **User role assignment** — `/api/v1/users/{id}/roles` (GET/POST/DELETE) and `/api/v1/roles/{id}/users` (GET/POST/DELETE) have no CLI equivalent. Bulk user-role operations require the portal.
3. **Billing & license management** — All of `/api/v1/billing/*` (licenses, subscriptions, license types, pricing) lacks CLI coverage. This entire domain is portal-only.
4. **API Tokens management** — `/api/v1/tokens` CRUD has no CLI equivalent despite being a key developer workflow.
5. **OAuth Clients management** — `/api/v1/clients` CRUD (Hydra OAuth clients) has no CLI equivalent.
6. **Set user password** — `/api/v1/users/{id}/password` (admin reset password) is missing. Useful for account recovery scripting.
7. **Invite user** — `/api/v1/users/invite` and `/api/v1/users/invitations` (invitation links) are missing; onboarding workflows require the portal.

### Secondary Gaps (moderate impact)

8. **Org extended management** — Branding (`/orgs/{id}/branding`), SaaS app activation (`/orgs/{id}/apps`), addresses (`/orgs/{id}/addresses`), contacts (`/orgs/{id}/contacts`) all lack CLI support.
9. **Report system** — `/api/v1/incloud/report*` (scheduled reports and policies) is entirely uncovered.
10. **Notification policies** — `/api/v1/alert/notify/policies*` CRUD has no CLI equivalent; alert configuration requires the portal.
11. **System messages** — `/api/v1/system/messages*` is uncovered; useful for scripting notification checks.
12. **AI/Copilot** — All `/api/v1/ai/*` and `/api/v1/copilot/*` endpoints have no CLI equivalent.
13. **Stats metrics management** — Custom metric CRUD (`/api/v1/stats/metrics`) is missing.
14. **Product-types & product-compatibilities** — Both sub-domains have no CLI support despite the portal having full CRUD.
15. **SIM/Link domain** — The entire `/api/v1/link/*` domain (SIM orders, invoices, coupons, carriers) is uncovered.
16. **User impersonation** — Sudo/impersonate endpoints (`/api/v1/users/{id}/impersonate/login`) have no CLI equivalent.
17. **User identity features** — `/api/v1/user/features` (feature compatibility per org) has no CLI equivalent.

### Low-priority Gaps

18. **MFA management** — Passkeys, TOTP, SMS association endpoints are UI-specific but could be useful for automated account setup.
19. **Uplinks** — `/api/v1/uplinks/{id}` detail exists in the device command module, not explicitly in admin domain.
20. **Phone validation** — `/api/v1/phone/validate` is a helper endpoint used by registration forms only.

## Notes

- The CLI covers the core user/org/product CRUD lifecycle and audit logs well, providing ~25% coverage of the admin domain.
- The largest uncovered domains are billing/licensing, the SIM/Link ecosystem, and AI/Copilot — all of which require dedicated modules to implement.
- The `incloud overview` subcommand covers traffic data usage (`/api/v1/datausage/*`) and alert statistics but not stats metrics management.
- The `incloud activity list` command maps to the primary audit log query endpoint.
- Source files analyzed:
  - Portal services: `apps/console/src/pages/iam/`, `apps/console/src/pages/micro/`, `apps/network/src/`, `apps/universal-login/src/`, `apps/oms/src/`, `apps/link/src/`, `apps/aihub/src/`, `packages/components/src/services/`, `packages/modules/src/`
  - CLI modules: `internal/cmd/user/`, `internal/cmd/org/`, `internal/cmd/role/`, `internal/cmd/product/`, `internal/cmd/activity/`, `internal/cmd/overview/`
