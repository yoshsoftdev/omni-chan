# API v1 Contract — Omni-Chan

**Phase 7 — Design specification only.**  
No implementation. Aligns frontend domain layers (Campaign, Template, Channel) with backend wire contracts.

---

## 1. Conventions

### 1.1 Base URL & versioning

| Version | Prefix | Status |
|---------|--------|--------|
| v1 | `/api/v1` | Current |
| v2 | `/api/v2` | Future breaking changes only |

All new clients MUST use `/api/v1/*`.  
Breaking changes require a new major version; v1 remains stable.

### 1.2 Envelope (all JSON responses)

**Success**

```json
{
  "success": true,
  "message": "human-readable summary",
  "data": {}
}
```

**Error**

```json
{
  "success": false,
  "message": "error description"
}
```

HTTP status codes: `400` validation, `401` unauthenticated, `403` forbidden, `404` not found, `409` conflict, `422` business rule, `500` internal.

Optional future extension (non-breaking in v1): `errors?: { field: string; message: string }[]` for validation details.

### 1.3 Authentication

| Mechanism | Usage |
|-----------|--------|
| `Authorization: Bearer <access_token>` | All protected routes |
| `POST /api/v1/auth/refresh` | Rotate access token |
| JWT claims | `sub` (user_id), `team_id` (workspace_id), `role`, `permissions[]` |

Workspace scoping: every resource is implicitly scoped to `team_id` from JWT. Cross-tenant access returns `403`.

### 1.4 Pagination

Query parameters (list endpoints):

| Param | Type | Default | Max |
|-------|------|---------|-----|
| `page` | integer | `1` | — |
| `page_size` | integer | `20` | `100` |

List response `data` shape:

```json
{
  "items": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 142,
    "total_pages": 8
  }
}
```

Legacy list endpoints that return bare arrays (e.g. `{ campaigns: [] }`) MAY remain during migration; new endpoints use `items` + `pagination`.

### 1.5 Filtering & sorting

Common query params:

| Param | Example | Notes |
|-------|---------|-------|
| `search` | `?search=promo` | Full-text / name match |
| `status` | `?status=APPROVED` | Resource-specific enum |
| `channel_id` | `?channel_id=uuid` | Formerly `device_id` |
| `device_id` | deprecated alias | v1 compat only |
| `sort` | `?sort=-created_at` | `-` prefix = DESC |
| `from` / `to` | ISO-8601 date range | Reports, inbox |

### 1.6 Identifiers & timestamps

- All primary keys: **UUID** strings
- All timestamps: **ISO-8601 UTC** (`2026-06-22T09:30:00Z`)
- Soft delete: resources expose `deleted_at` only in admin/detail views (optional v1)

### 1.7 Domain ↔ DTO mapping (frontend)

| Frontend domain | API DTO | Boundary module |
|-----------------|---------|-----------------|
| `Channel` | `ChannelDTO` | `channel.boundary.ts` |
| `Template` | `TemplateDTO` | `template.boundary.ts` |
| `Campaign` | `CampaignDTO` | `campaign.boundary.ts` |

**v1 compatibility:** existing `DeviceDTO` ≡ `ChannelDTO` with field `name` (maps to domain `label`).  
Canonical path going forward: `/channels`. Legacy `/devices` retained as alias until v2 removal.

---

## 2. Domain models (wire / OpenAPI components)

### 2.1 Workspace

```yaml
WorkspaceDTO:
  type: object
  required: [id, name, slug, created_at]
  properties:
    id: { type: string, format: uuid }
    name: { type: string }
    slug: { type: string }
    owner_id: { type: string, format: uuid }
    owner_name: { type: string }
    timezone: { type: string, example: "Asia/Jakarta" }
    country: { type: string, example: "ID" }
    created_at: { type: string, format: date-time }
    updated_at: { type: string, format: date-time }

WorkspaceReadinessDTO:
  type: object
  properties:
    score: { type: integer, minimum: 0, maximum: 100 }
    completed_count: { type: integer }
    total_count: { type: integer }
    status: { type: string, enum: [not_started, in_progress, ready] }
    checks:
      type: array
      items:
        type: object
        properties:
          id: { type: string, enum: [channel_connected, template_synced, contacts_available, team_members_added, campaign_created] }
          label: { type: string }
          completed: { type: boolean }
```

### 2.2 Channel (Device)

```yaml
ChannelStatus:
  type: string
  enum: [disconnected, connecting, connected]

ChannelConnectionType:
  type: string
  enum: [manual, embedded_coex]

ChannelOnboardingDTO:
  type: object
  properties:
    status: { type: string, enum: [pending, syncing_contacts, syncing_history, completed, failed] }
    contacts_synced_count: { type: integer }
    history_messages_synced: { type: integer }
    history_sync_progress: { type: integer, minimum: 0, maximum: 100 }
    error: { type: string, nullable: true }
    completed_at: { type: string, format: date-time, nullable: true }

ChannelDTO:
  type: object
  required: [id, workspace_id, name, provider, status, created_at, updated_at]
  properties:
    id: { type: string, format: uuid }
    workspace_id: { type: string, format: uuid }
    name: { type: string, description: "Display label; maps to Channel.label in frontend" }
    phone_number: { type: string, nullable: true }
    waba_id: { type: string, nullable: true }
    phone_number_id: { type: string, nullable: true }
    provider: { type: string, enum: [meta], default: meta }
    status: { $ref: '#/components/schemas/ChannelStatus' }
    connection_type: { $ref: '#/components/schemas/ChannelConnectionType' }
    is_coex: { type: boolean, default: false }
    onboarding: { $ref: '#/components/schemas/ChannelOnboardingDTO' }
    last_connected_at: { type: string, format: date-time, nullable: true }
    created_at: { type: string, format: date-time }
    updated_at: { type: string, format: date-time }

ConnectChannelManualRequest:
  type: object
  required: [channel_id, phone_number_id, waba_id, access_token]
  properties:
    channel_id: { type: string, format: uuid, description: "Shell channel record" }
    phone_number_id: { type: string }
    waba_id: { type: string }
    access_token: { type: string, writeOnly: true }

ConnectChannelCoexRequest:
  type: object
  required: [channel_id, code, waba_id]
  properties:
    channel_id: { type: string, format: uuid }
    code: { type: string, description: "Meta embedded signup OAuth code" }
    waba_id: { type: string }
    phone_number_id: { type: string, nullable: true }

ChannelSyncRequest:
  type: object
  required: [channel_id]
  properties:
    channel_id: { type: string, format: uuid }
    sync_contacts: { type: boolean, default: true }
    sync_history: { type: boolean, default: true }
```

### 2.3 Template

```yaml
TemplateApiStatus:
  type: string
  enum: [APPROVED, PENDING, REJECTED]

TemplateComponentButtonDTO:
  type: object
  required: [type, text]
  properties:
    type: { type: string, enum: [URL, QUICK_REPLY, PHONE_NUMBER, OTP, COPY_CODE] }
    text: { type: string }
    url: { type: string }
    phone_number: { type: string }
    otp_type: { type: string }

TemplateComponentDTO:
  oneOf:
    - type: object
      required: [type]
      properties:
        type: { const: HEADER }
        format: { type: string, enum: [TEXT, IMAGE, VIDEO, DOCUMENT] }
        text: { type: string }
    - type: object
      required: [type, text]
      properties:
        type: { const: BODY }
        text: { type: string }
        example: { type: object }
    - type: object
      required: [type, text]
      properties:
        type: { const: FOOTER }
        text: { type: string }
    - type: object
      required: [type, buttons]
      properties:
        type: { const: BUTTONS }
        buttons: { type: array, items: { $ref: '#/components/schemas/TemplateComponentButtonDTO' } }

TemplateDTO:
  type: object
  required: [id, channel_id, waba_id, meta_template_id, name, language, components, created_at, updated_at]
  properties:
    id: { type: string, format: uuid }
    channel_id: { type: string, format: uuid, description: "Formerly device_id" }
    device_id: { type: string, format: uuid, deprecated: true }
    waba_id: { type: string }
    meta_template_id: { type: string }
    name: { type: string }
    language: { type: string, example: "id_ID" }
    category: { type: string, enum: [MARKETING, UTILITY, AUTHENTICATION] }
    status: { $ref: '#/components/schemas/TemplateApiStatus' }
    preview: { type: string }
    components: { type: array, items: { $ref: '#/components/schemas/TemplateComponentDTO' } }
    variable_count: { type: integer }
    last_sync_at: { type: string, format: date-time, nullable: true }
    created_at: { type: string, format: date-time }
    updated_at: { type: string, format: date-time }

CreateTemplateRequest:
  type: object
  required: [channel_id, name, language, category, components]
  properties:
    channel_id: { type: string, format: uuid }
    name: { type: string }
    language: { type: string }
    category: { type: string }
    components: { type: array, items: { $ref: '#/components/schemas/TemplateComponentDTO' } }

SyncTemplatesRequest:
  type: object
  required: [channel_id]
  properties:
    channel_id: { type: string, format: uuid }

SyncTemplatesResponse:
  type: object
  properties:
    synced: { type: integer }
    channel_id: { type: string, format: uuid }
```

### 2.4 Campaign

```yaml
CampaignApiStatus:
  type: string
  enum: [draft, scheduled, queued, sending, completed, failed, cancelled]

CampaignDTO:
  type: object
  required: [id, name, status, channel_id, template_id, total_contacts, sent_count, delivered_count, read_count, failed_count, created_at, updated_at]
  properties:
    id: { type: string, format: uuid }
    name: { type: string }
    description: { type: string, nullable: true }
    status: { $ref: '#/components/schemas/CampaignApiStatus' }
    channel_id: { type: string, format: uuid }
    device_id: { type: string, format: uuid, deprecated: true }
    channel_name: { type: string, nullable: true }
    channel_phone: { type: string, nullable: true }
    device_name: { type: string, deprecated: true }
    device_phone: { type: string, deprecated: true }
    template_id: { type: string, format: uuid }
    template_name: { type: string }
    template_category: { type: string }
    template_language: { type: string }
    segment_id: { type: string, format: uuid, nullable: true }
    segment_name: { type: string, nullable: true }
    segment_contact_count: { type: integer }
    variable_mapping:
      type: object
      properties:
        body: { type: object, additionalProperties: { type: string } }
    sample_values: { type: object, additionalProperties: { type: string } }
    timezone: { type: string, nullable: true }
    send_mode: { type: string, enum: [now, later], nullable: true }
    scheduled_at: { type: string, format: date-time, nullable: true }
    started_at: { type: string, format: date-time, nullable: true }
    completed_at: { type: string, format: date-time, nullable: true }
    total_contacts: { type: integer }
    sent_count: { type: integer }
    delivered_count: { type: integer }
    read_count: { type: integer }
    failed_count: { type: integer }
    reply_count: { type: integer, default: 0 }
    clicked_count: { type: integer, default: 0 }
    created_by: { type: string, nullable: true }
    created_at: { type: string, format: date-time }
    updated_at: { type: string, format: date-time }

CreateCampaignRequest:
  type: object
  required: [name, channel_id, template_id]
  properties:
    name: { type: string }
    description: { type: string }
    channel_id: { type: string, format: uuid }
    template_id: { type: string, format: uuid }
    segment_ids: { type: array, items: { type: string, format: uuid } }
    tag_ids: { type: array, items: { type: string, format: uuid } }
    contact_ids: { type: array, items: { type: string, format: uuid } }
    variable_mapping: { type: object }
    scheduled_at: { type: string, format: date-time }
    timezone: { type: string }
    send_mode: { type: string, enum: [now, later] }

PatchCampaignRequest:
  type: object
  properties:
    name: { type: string }
    description: { type: string }
    variable_mapping: { type: object }
    scheduled_at: { type: string, format: date-time, nullable: true }

ScheduleCampaignRequest:
  type: object
  required: [scheduled_at]
  properties:
    scheduled_at: { type: string, format: date-time }
    timezone: { type: string, default: "Asia/Jakarta" }
```

**Frontend mapping note:** API statuses `queued` / `sending` map to domain `running` via `campaign.boundary.ts`.

### 2.5 Schedule

Schedules are campaigns with `status=scheduled` or dedicated schedule metadata. No separate entity in v1.

```yaml
ScheduleDTO:
  allOf:
    - $ref: '#/components/schemas/CampaignDTO'
    - type: object
      properties:
        schedule_status: { type: string, enum: [scheduled, running, completed, cancelled] }
```

### 2.6 Reports

```yaml
ReportDashboardDTO:
  type: object
  properties:
    stats:
      type: object
      properties:
        total_campaigns: { type: integer }
        messages_sent: { type: integer }
        delivered_rate: { type: number }
        read_rate: { type: number }
        reply_rate: { type: number }
        changes: { type: object }
    performance_trend: { type: array, items: { type: object } }
    channel_breakdown: { type: array, items: { type: object } }

CampaignReportDTO:
  type: object
  properties:
    campaign: { $ref: '#/components/schemas/CampaignDTO' }
    funnel:
      type: object
      properties:
        sent: { type: integer }
        delivered: { type: integer }
        read: { type: integer }
        clicked: { type: integer }
        replies: { type: integer }
    links: { type: array, items: { type: object } }
    locations: { type: array, items: { type: object } }
    devices: { type: array, items: { type: object } }
    delivery_breakdown: { type: array, items: { type: object } }
    hourly_performance: { type: array, items: { type: object } }
```

### 2.7 Contact

```yaml
ContactDTO:
  type: object
  required: [id, name, phone, created_at, updated_at]
  properties:
    id: { type: string, format: uuid }
    name: { type: string }
    phone: { type: string }
    email: { type: string, nullable: true }
    notes: { type: string, nullable: true }
    city: { type: string, nullable: true }
    consent_status: { type: string, enum: [opt_in, opt_out, unknown] }
    source: { type: string, nullable: true }
    tag_ids: { type: array, items: { type: string, format: uuid } }
    created_at: { type: string, format: date-time }
    updated_at: { type: string, format: date-time }

CreateContactRequest:
  type: object
  required: [name, phone]
  properties:
    name: { type: string }
    phone: { type: string }
    email: { type: string }
    notes: { type: string }
    tag_ids: { type: array, items: { type: string, format: uuid } }

ContactImportResultDTO:
  type: object
  properties:
    import_id: { type: string, format: uuid }
    total_rows: { type: integer }
    imported: { type: integer }
    skipped: { type: integer }
    failed: { type: integer }
    errors: { type: array, items: { type: object } }

ContactImportJobDTO:
  type: object
  properties:
    id: { type: string, format: uuid }
    filename: { type: string }
    source: { type: string, enum: [csv, excel, manual] }
    status: { type: string, enum: [pending, processing, completed, failed] }
    imported_count: { type: integer }
    created_at: { type: string, format: date-time }
```

### 2.8 Segment

```yaml
SegmentRuleDTO:
  type: object
  required: [field, operator, value]
  properties:
    id: { type: string, format: uuid }
    field: { type: string, enum: [tag, city, last_activity, consent_status, source] }
    operator: { type: string, enum: [equals, not_equals, contains, less_than, greater_than] }
    value: { type: string }

SegmentDTO:
  type: object
  required: [id, name, contact_count, created_at, updated_at]
  properties:
    id: { type: string, format: uuid }
    name: { type: string }
    description: { type: string }
    color: { type: string }
    contact_count: { type: integer }
    rules: { type: array, items: { $ref: '#/components/schemas/SegmentRuleDTO' } }
    created_at: { type: string, format: date-time }
    updated_at: { type: string, format: date-time }

CreateSegmentRequest:
  type: object
  required: [name, rules]
  properties:
    name: { type: string }
    description: { type: string }
    color: { type: string }
    rules: { type: array, items: { $ref: '#/components/schemas/SegmentRuleDTO' } }
```

### 2.9 Tag

```yaml
TagDTO:
  type: object
  required: [id, name, created_at]
  properties:
    id: { type: string, format: uuid }
    name: { type: string }
    color: { type: string }
    contact_count: { type: integer }
    created_at: { type: string, format: date-time }
```

### 2.10 Team

```yaml
TeamMemberDTO:
  type: object
  required: [id, user_id, email, name, role, created_at]
  properties:
    id: { type: string, format: uuid }
    user_id: { type: string, format: uuid }
    email: { type: string }
    name: { type: string }
    role: { type: string, enum: [owner, admin, supervisor, agent] }
    role_id: { type: string, format: uuid }
    status: { type: string, enum: [active, inactive, invited] }
    last_active_at: { type: string, format: date-time, nullable: true }
    created_at: { type: string, format: date-time }

RoleDTO:
  type: object
  properties:
    id: { type: string, format: uuid }
    name: { type: string }
    description: { type: string }
    is_system: { type: boolean }

PermissionDTO:
  type: object
  properties:
    id: { type: string, format: uuid }
    code: { type: string, example: "campaign.create" }
    name: { type: string }
    description: { type: string }

InviteTeamMemberRequest:
  type: object
  required: [email, name, role]
  properties:
    email: { type: string, format: email }
    name: { type: string }
    role: { type: string, enum: [admin, supervisor, agent] }
```

### 2.11 Inbox

```yaml
ConversationStatus:
  type: string
  enum: [open, assigned, closed]

ConversationDTO:
  type: object
  required: [id, contact_id, channel, status, updated_at]
  properties:
    id: { type: string, format: uuid }
    contact_id: { type: string, format: uuid }
    contact_name: { type: string }
    contact_phone: { type: string }
    channel: { type: string, enum: [whatsapp, instagram, other] }
    channel_id: { type: string, format: uuid, nullable: true }
    status: { $ref: '#/components/schemas/ConversationStatus' }
    assigned_to: { type: string, format: uuid, nullable: true }
    assigned_name: { type: string, nullable: true }
    unread_count: { type: integer }
    last_message_preview: { type: string }
    last_message_at: { type: string, format: date-time }
    tags: { type: array, items: { type: string } }
    updated_at: { type: string, format: date-time }

MessageDTO:
  type: object
  required: [id, conversation_id, direction, content, created_at]
  properties:
    id: { type: string, format: uuid }
    conversation_id: { type: string, format: uuid }
    channel_id: { type: string, format: uuid, nullable: true }
    wamid: { type: string, nullable: true }
    sender: { type: string }
    recipient: { type: string, nullable: true }
    content: { type: string }
    direction: { type: string, enum: [inbound, outbound] }
    status: { type: string, enum: [sent, delivered, read, failed], nullable: true }
    received_at: { type: string, format: date-time }
    created_at: { type: string, format: date-time }

SendMessageRequest:
  type: object
  required: [channel_id, recipient, content]
  properties:
    channel_id: { type: string, format: uuid }
    recipient: { type: string, description: "E.164 phone" }
    content: { type: string }
```

---

## 3. Endpoint inventory

### 3.1 Auth & workspace

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| POST | `/api/v1/auth/register` | Register user + workspace | public |
| POST | `/api/v1/auth/login` | Login | public |
| POST | `/api/v1/auth/refresh` | Refresh token | authenticated |
| POST | `/api/v1/auth/logout` | Revoke refresh token | authenticated |
| GET | `/api/v1/workspace` | Current workspace detail | `user.view` |
| PATCH | `/api/v1/workspace` | Update workspace settings | `team.manage` |
| GET | `/api/v1/workspace/readiness` | Onboarding readiness score | `user.view` |
| GET | `/api/v1/meta/embedded-config` | Meta embedded signup config | `channel.manage` |

**Legacy alias:** `GET /api/v1/team/workspace` → same as `/api/v1/workspace`

### 3.2 Channels

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/channels` | List channels | `channel.view` |
| POST | `/api/v1/channels` | Create channel shell | `channel.manage` |
| GET | `/api/v1/channels/{id}` | Get channel | `channel.view` |
| PATCH | `/api/v1/channels/{id}` | Rename channel | `channel.manage` |
| DELETE | `/api/v1/channels/{id}` | Delete channel | `channel.manage` |
| POST | `/api/v1/channels/connect` | Manual Meta connect | `channel.manage` |
| POST | `/api/v1/channels/coex/start` | Embedded signup connect | `channel.manage` |
| POST | `/api/v1/channels/{id}/disconnect` | Disconnect channel | `channel.manage` |
| GET | `/api/v1/channels/onboarding-status` | Coex onboarding status (`?channel_id=`) | `channel.view` |
| POST | `/api/v1/channels/onboarding/sync` | Start coex contact/history sync | `channel.manage` |
| POST | `/api/v1/channels/sync` | Post-connect asset sync | `channel.manage` |

**v1 legacy aliases (deprecated, same handlers):**

| Legacy | Canonical |
|--------|-----------|
| `GET /api/v1/devices` | `GET /api/v1/channels` |
| `POST /api/v1/devices/{id}/connect` | `POST /api/v1/channels/connect` |
| `POST /api/v1/devices/{id}/connect/embedded` | `POST /api/v1/channels/coex/start` |
| `GET /api/v1/devices/{id}/onboarding/status` | `GET /api/v1/channels/onboarding-status` |

**Filters (GET /channels):** `status`, `provider`, `search`, `page`, `page_size`

### 3.3 Templates

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/templates` | List templates | `template.view` |
| GET | `/api/v1/templates/{id}` | Get template | `template.view` |
| POST | `/api/v1/templates` | Submit template to Meta | `template.create` |
| POST | `/api/v1/templates/sync` | Sync from Meta | `template.sync` |

**Filters:** `channel_id`, `status`, `language`, `category`, `search`, `page`, `page_size`

### 3.4 Campaigns

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/campaigns` | List campaigns | `campaign.view` |
| GET | `/api/v1/campaigns/{id}` | Get campaign | `campaign.view` |
| POST | `/api/v1/campaigns` | Create campaign (draft) | `campaign.create` |
| PATCH | `/api/v1/campaigns/{id}` | Update draft/scheduled | `campaign.update` |
| POST | `/api/v1/campaigns/{id}/send` | Queue immediate send | `campaign.send` |
| POST | `/api/v1/campaigns/{id}/schedule` | Schedule campaign | `campaign.schedule` |
| POST | `/api/v1/campaigns/{id}/cancel` | Cancel campaign | `campaign.cancel` |
| DELETE | `/api/v1/campaigns/{id}` | Delete draft | `campaign.delete` |

**Filters:** `status`, `channel_id`, `template_id`, `segment_id`, `search`, `from`, `to`, `page`, `page_size`

**Async rule:** `send` and `schedule` return `202 Accepted` with updated `CampaignDTO`; worker processes queue.

### 3.5 Schedules

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/schedules` | List scheduled/running campaigns |
| GET | `/api/v1/schedules/{id}` | Schedule detail |

Alias of `GET /campaigns?status=scheduled,running,sending` with calendar-friendly fields.

### 3.6 Reports

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/reports/dashboard` | Workspace KPI dashboard | `report.view` |
| GET | `/api/v1/reports/campaigns` | Campaign report list | `report.view` |
| GET | `/api/v1/reports/campaigns/{id}` | Campaign report detail | `report.view` |

**Filters (dashboard):** `from`, `to`, `channel_id`

### 3.7 Contacts

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/contacts` | List contacts | `contact.view` |
| POST | `/api/v1/contacts` | Create contact | `contact.create` |
| GET | `/api/v1/contacts/{id}` | Get contact | `contact.view` |
| PATCH | `/api/v1/contacts/{id}` | Update contact | `contact.update` |
| DELETE | `/api/v1/contacts/{id}` | Delete contact | `contact.delete` |
| POST | `/api/v1/contacts/import/csv` | Import CSV (multipart) | `contact.import` |
| POST | `/api/v1/contacts/import/excel` | Import Excel (multipart) | `contact.import` |
| GET | `/api/v1/contacts/imports` | List import jobs | `contact.view` |
| GET | `/api/v1/contacts/imports/{id}` | Import job status | `contact.view` |

**Filters:** `search`, `tag_id`, `segment_id`, `consent_status`, `page`, `page_size`

### 3.8 Segments

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/segments` | List segments | `segment.view` |
| POST | `/api/v1/segments` | Create segment | `segment.create` |
| GET | `/api/v1/segments/{id}` | Get segment + rules | `segment.view` |
| PATCH | `/api/v1/segments/{id}` | Update segment | `segment.update` |
| DELETE | `/api/v1/segments/{id}` | Delete segment | `segment.delete` |
| POST | `/api/v1/segments/{id}/preview` | Preview audience count | `segment.view` |
| GET | `/api/v1/segments/{id}/contacts` | Paginated segment members | `segment.view` |

### 3.9 Tags

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/tags` | List tags | `tag.view` |
| POST | `/api/v1/tags` | Create tag | `tag.create` |
| PATCH | `/api/v1/tags/{id}` | Update tag | `tag.update` |
| DELETE | `/api/v1/tags/{id}` | Delete tag | `tag.delete` |

### 3.10 Team

| Method | Path | Permission |
|--------|------|------------|
| GET | `/api/v1/team/users` | `user.view` |
| POST | `/api/v1/team/users/invite` | `user.invite` |
| PUT | `/api/v1/team/users/{id}/role` | `user.update` |
| PATCH | `/api/v1/team/users/{id}/deactivate` | `user.update` |
| DELETE | `/api/v1/team/users/{id}` | `user.delete` |
| GET | `/api/v1/team/roles` | `user.view` |
| GET | `/api/v1/team/roles/{id}` | `user.view` |
| GET | `/api/v1/team/permissions` | `team.manage` |

**Legacy:** `/api/v1/teams`, `/api/v1/teams/members` (unchanged v1)

### 3.11 Inbox

| Method | Path | Description | Permission |
|--------|------|-------------|------------|
| GET | `/api/v1/conversations` | List conversations | `inbox.view` |
| GET | `/api/v1/conversations/{id}` | Conversation detail | `inbox.view` |
| PATCH | `/api/v1/conversations/{id}` | Assign / close / tags | `inbox.update` |
| GET | `/api/v1/conversations/{id}/messages` | Message history | `inbox.view` |
| POST | `/api/v1/messages` | Send outbound message | `inbox.send` |
| GET | `/api/v1/inbox/unread-count` | Unread badge count | `inbox.view` |

**Filters (conversations):** `status`, `assigned_to`, `channel_id`, `search`, `tag`, `page`, `page_size`

### 3.12 Webhooks (inbound, no auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/webhooks/meta` | Meta verification challenge |
| POST | `/api/v1/webhooks/meta` | Meta WhatsApp events |

### 3.13 WebSocket

| Path | Description |
|------|-------------|
| `WS /api/v1/ws` | Real-time inbox events (JWT query param or subprotocol) |

---

## 4. Request / response examples

### 4.1 List channels

```http
GET /api/v1/channels?status=connected&page=1&page_size=20
Authorization: Bearer <token>
```

```json
{
  "success": true,
  "message": "channels retrieved",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "workspace_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
        "name": "Skinsure Glow Official",
        "phone_number": "+6281234567890",
        "waba_id": "123456789",
        "phone_number_id": "987654321",
        "provider": "meta",
        "status": "connected",
        "connection_type": "embedded_coex",
        "is_coex": true,
        "onboarding": {
          "status": "completed",
          "contacts_synced_count": 1240,
          "history_messages_synced": 8420,
          "history_sync_progress": 100
        },
        "last_connected_at": "2026-06-22T08:00:00Z",
        "created_at": "2026-06-20T10:00:00Z",
        "updated_at": "2026-06-22T08:00:00Z"
      }
    ],
    "pagination": { "page": 1, "page_size": 20, "total": 1, "total_pages": 1 }
  }
}
```

### 4.2 Create campaign

```http
POST /api/v1/campaigns
Content-Type: application/json
```

```json
{
  "name": "Promo Juni 2026",
  "description": "Diskon 20% pelanggan VIP",
  "channel_id": "550e8400-e29b-41d4-a716-446655440000",
  "template_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "segment_ids": ["seg_001"],
  "variable_mapping": {
    "body": { "1": "name", "2": "discount" }
  },
  "send_mode": "later",
  "scheduled_at": "2026-06-25T09:00:00Z",
  "timezone": "Asia/Jakarta"
}
```

### 4.3 Sync templates

```http
POST /api/v1/templates/sync
```

```json
{
  "channel_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

```json
{
  "success": true,
  "message": "templates synced",
  "data": { "synced": 12, "channel_id": "550e8400-e29b-41d4-a716-446655440000" }
}
```

---

## 5. Webhook format (Meta WhatsApp)

### 5.1 Verification (GET)

```
GET /api/v1/webhooks/meta?hub.mode=subscribe&hub.verify_token=<token>&hub.challenge=<challenge>
→ 200 text/plain challenge
```

### 5.2 Event payload (POST)

```json
{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "<WABA_ID>",
      "changes": [
        {
          "field": "messages",
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "6281234567890",
              "phone_number_id": "<PHONE_NUMBER_ID>"
            },
            "contacts": [{ "profile": { "name": "Aulia" }, "wa_id": "628111222333" }],
            "messages": [
              {
                "from": "628111222333",
                "id": "wamid.xxx",
                "timestamp": "1719043200",
                "type": "text",
                "text": { "body": "Halo" }
              }
            ]
          }
        }
      ]
    }
  ]
}
```

**Internal processing (future):** webhook → normalize → `MessageDTO` → NATS → inbox worker → WebSocket `inbox.message.created`.

**Status webhooks:** `field: "message_template_status_update"` for template approval events.

---

## 6. Error catalog (recommended)

| HTTP | Message pattern | When |
|------|-----------------|------|
| 400 | `validation failed: <field>` | Invalid input |
| 401 | `unauthorized` | Missing/expired token |
| 403 | `permission denied` | RBAC failure |
| 403 | `channel not connected` | Template sync / send blocked |
| 404 | `<resource> not found` | Invalid ID |
| 409 | `campaign already sending` | Invalid state transition |
| 422 | `template not approved` | Campaign send blocked |
| 422 | `no connected channels` | Workspace readiness |
| 429 | `rate limit exceeded` | Meta / API throttle |
| 500 | `internal server error` | Unexpected |

---

## 7. State machines (contract rules)

### 7.1 Channel

```
disconnected → connecting → connected
connected → disconnected (disconnect)
coex: connected → onboarding(pending → syncing_* → completed|failed)
```

### 7.2 Campaign

```
draft → scheduled → queued → sending → completed|failed
draft|scheduled → cancelled
```

Send/schedule MUST be rejected unless:
- `channel.status = connected`
- `template.status = APPROVED`
- audience count > 0

### 7.3 Template

Synced from Meta only. `POST /templates` submits to Meta → `PENDING` until webhook/status poll → `APPROVED|REJECTED`.

---

## 8. Frontend repository mapping (target)

| Frontend service | API repository | Base path |
|------------------|----------------|-----------|
| `channelQueryService` | `channel.repository.api` | `/channels` |
| `templateQueryService` | `template.repository.api` | `/templates` |
| `campaignQueryService` | `campaign.repository.api` | `/campaigns` |
| `segmentService` | `segment.repository.api` | `/segments` |
| `contactService` | `contact.repository.api` | `/contacts` |
| `teamService` | `team.repository.api` | `/team` |
| `inboxService` | `inbox.repository.api` | `/conversations`, `/messages` |

All API repositories MUST route responses through domain boundaries before returning to UI.

---

## 9. Migration notes (devices → channels)

| Concern | v1 strategy |
|---------|-------------|
| URL | Dual-register `/devices` and `/channels` |
| JSON field | Accept `device_id` in requests; respond with both `channel_id` and `device_id` |
| JSON field | `name` in API = `label` in Channel domain |
| Removal | v2 drops `device_id` and `/devices` paths |

---

## 10. OpenAPI artifact (summary)

Full OpenAPI 3.1 document SHOULD be generated at `docs/openapi/v1.yaml` in a future phase from this contract. This document is the authoritative design source for v1.

**Tags:** `Auth`, `Workspace`, `Channels`, `Templates`, `Campaigns`, `Schedules`, `Reports`, `Contacts`, `Segments`, `Tags`, `Team`, `Inbox`, `Webhooks`

**Security scheme:**

```yaml
securitySchemes:
  bearerAuth:
    type: http
    scheme: bearer
    bearerFormat: JWT
```

---

*End of Phase 7 contract design — awaiting review before backend implementation.*
