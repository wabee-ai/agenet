# Agenet v2 — Discovery and registration

## The Agent Card is the registration document

v1 agents announced a bare `requestType` string. v2 agents present an A2A 1.0 `AgentCard`,
and the card is the only thing the mesh needs to know about an agent. Required fields for
registration:

| Field | Requirement |
|---|---|
| `name`, `version`, `provider` | `name` is unique per tenant; it becomes the virtual endpoint path `/agents/{name}` |
| `skills[]` | At least one. Each `id` is the routing key. `tags[]`, `inputModes[]`, `outputModes[]`, `description`, `examples[]` are indexed for discovery. |
| `capabilities.streaming`, `pushNotifications` | Honest values; the Gateway advertises the intersection of agent and mesh support |
| `securitySchemes`, `securityRequirements` | Describe what the *agent* expects on the task token (usually the mesh's own scheme) |
| `supportedInterfaces[]` | For dial-out agents, empty on registration; the Dispatcher fills in the virtual endpoint. For pull-mode agents, the URL the Dispatcher will call. |
| `signatures[]` | **Required.** JWS over the RFC 8785 canonical card. The key must chain to the tenant's trust anchors. |
| `capabilities.extensions[]` | Must include an `AgentExtension` with `uri = https://agenet.io/ext/routing/v1` whose `params` carry the Agenet hints below; may include `https://agenet.io/ext/resumable/v1` and `https://agenet.io/ext/mcp-tools/v1` |

A2A 1.0's `AgentCard` and `AgentSkill` have no `metadata` field, so Agenet hints cannot ride on
the card itself. They live in the `params` struct of the routing extension entry, keyed by
skill id where they are per-skill. This survives the unmodified `a2a.v1.AgentCard` imported by
`Register.card` and is covered by the card signature.

```json
{
  "uri": "https://agenet.io/ext/routing/v1",
  "required": false,
  "params": {
    "card": { "maxInflight": 16 },
    "skills": {
      "summarize": {
        "version": "2.1.0",
        "idempotent": true,
        "resumable": true,
        "sideEffects": "none",
        "dataClasses": ["pii"],
        "costHint": { "perCallUsd": 0.05, "tokensPerCall": 4000 },
        "inputSchema": { "type": "object", "properties": { "question": { "type": "string" } } }
      }
    }
  }
}
```

| Key | Scope | Values | Effect |
|---|---|---|---|
| `version` | skill | semver | Skill version for range matching |
| `idempotent` | skill | bool | Allows redelivery after a committed `Ack` |
| `resumable` | skill | bool | Allows a parked task to move to another instance with history |
| `sideEffects` | skill | `none`, `local`, `external` | Mirrors MCP's `readOnlyHint`/`destructiveHint`; policy uses it to require approval |
| `dataClasses` | skill | list | Data classifications the skill may handle (`pii`, `phi`, `secret`); policy compares with the request's labels |
| `costHint` | skill | `{perCallUsd, tokensPerCall}` | Scheduling and budget pre-checks |
| `inputSchema`, `outputSchema` | skill | JSON Schema 2020-12 | Validation and MCP tool projection |
| `maxInflight` | card | int | Default capacity if `Register.capacity` is omitted |

Skills with no entry under `params.skills` get the conservative defaults `idempotent=false`,
`resumable=false`, `sideEffects=external`.

## Pinning and rug-pull defense

The Dispatcher hashes the canonical card at registration and records `card_hash` on every task
it delivers. A re-registration with a different hash and the same `version` is rejected. A
new `version` triggers policy re-evaluation; tenants may require explicit approval of new
cards before any traffic reaches them. This blocks the "rug pull" pattern where a benign
description is swapped for a poisoned one after approval.

## Three attachment modes

### Dial-out (the v1 model, kept)

```
agent ──Attach(Register{card, instance_id, capacity})──▶ dispatcher
agent ◀─Registered{agent_url, card_hash, lease_ttl}──── dispatcher
```

The agent needs no inbound port, no public URL, and no TLS certificate of its own. Identity is
the SPIFFE SVID on the mTLS connection or an OAuth client-credentials token in gRPC metadata,
plus the card signature. Multiple instances register the same card with different
`instance_id`s and share the skill queue.

### Pull mode (existing A2A servers)

An operator registers `{name, url, cardSource: fetched|pinned, auth}` with the Dispatcher.
The Dispatcher fetches `/.well-known/agent-card.json`, verifies the signature, pins the hash,
and thereafter acts as an A2A client. The delivery path depends on the pinned card's
`capabilities`, because A2A requires streaming operations to fail when `streaming=false`:

| Agent capabilities | Delivery | Progress | Recovery |
|---|---|---|---|
| `streaming=true` | `SendStreamingMessage` | events on the stream | `SubscribeToTask` while non-terminal, `GetTask` otherwise |
| `streaming=false`, `pushNotifications=true` | `SendMessage` with a `TaskPushNotificationConfig` pointing at the Dispatcher's webhook | webhook callbacks, each followed by `GetTask` | `GetTask` |
| `streaming=false`, `pushNotifications=false` | `SendMessage` | `GetTask` polling with backoff (`pullPollInterval`, default 2 s, capped at 30 s) | `GetTask` |

In all three rows `CancelTask` is used for cancellation, the Dispatcher converts each observed
status or artifact change into a task-store event with its own offset, and a `SendMessage`
that returns a terminal `Task` inline closes the Agenet task in one step. The card is
re-fetched on the `ttlMs` cadence; a hash change follows the same rules as a dial-out
re-registration.

### Tool agent (MCP servers)

See [06 MCP bridge](06-mcp-bridge.md). The adapter synthesizes a card with one skill per
tool and registers it in dial-out mode.

## What clients see

- `GET /.well-known/agent-card.json` on the Gateway returns the **mesh card**: `name` is the
  mesh, `skills[]` is the union of registered skills the caller's token may see, the routing
  extension's `params.skills[<id>].agents` lists the providing agent names, and
  `capabilities.extensions[]` advertises the routing extension. Unauthenticated callers get
  the public subset; `GetExtendedAgentCard` returns the full view after auth.
- `GET /agents/{name}/.well-known/agent-card.json` returns a **virtual card**: the agent's
  card with `supportedInterfaces[]` rewritten to the virtual endpoint and `securitySchemes`
  replaced by the Gateway's. Because A2A signatures cover the canonical card content, those
  rewrites invalidate the agent's original `signatures[]`. The Gateway therefore strips them
  and signs the virtual card with the **mesh signing key**, which tenants publish as a trust
  anchor alongside their agent-provider keys. The virtual card records
  `params.card.sourceCardHash` on its routing extension entry, the pinned hash of the original
  signed card, and the
  original is served unmodified at `/agents/{name}/.well-known/agent-card.json?source=1` for
  clients that want to verify the agent's own signature. A pure-A2A client uses the virtual
  card and never needs the extension.
- `GET /skills?tag=finance&inputMode=application/pdf` is a convenience search over the same
  index. Results carry `ttlMs` and `cacheScope`, matching MCP's `CacheableResult`.

## Feeding external registries

The Dispatcher can export its live view as:

- MCP Registry `server.json` entries for tool agents (and for any agent via the MCP facade),
  pushed to a private sub-registry.
- AGNTCY OASF records, if a tenant runs an Agent Directory.

The mesh itself is not a registry of record; it only knows what is attached right now.
