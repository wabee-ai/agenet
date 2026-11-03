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
| `capabilities.extensions[]` | Must include `https://agenet.io/ext/routing/v1`; may include `https://agenet.io/ext/resumable/v1` and `https://agenet.io/ext/mcp-tools/v1` |

Agenet-specific hints live in `AgentSkill.metadata` and `AgentCard.metadata`:

| Key | Values | Effect |
|---|---|---|
| `agenet.io/version` | semver | Skill version for range matching |
| `agenet.io/idempotent` | bool | Allows redelivery after `Ack` |
| `agenet.io/resumable` | bool | Allows a parked task to move to another instance with history |
| `agenet.io/sideEffects` | `none`, `local`, `external` | Mirrors MCP's `readOnlyHint`/`destructiveHint`; policy uses it to require approval |
| `agenet.io/dataClasses` | list | Data classifications the skill may handle (`pii`, `phi`, `secret`); policy compares with the request's labels |
| `agenet.io/costHint` | `{perCallUsd, tokensPerCall}` | Scheduling and budget pre-checks |
| `agenet.io/maxInflight` | int | Default capacity if `Register.capacity` is omitted |

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
and thereafter acts as an A2A client: `SendStreamingMessage` for delivery, `SubscribeToTask`
to recover after a dropped stream, `CancelTask` for cancellation. The card is re-fetched on
the `ttlMs` cadence; a hash change follows the same rules as a dial-out re-registration.

### Tool agent (MCP servers)

See [06 MCP bridge](06-mcp-bridge.md). The adapter synthesizes a card with one skill per
tool and registers it in dial-out mode.

## What clients see

- `GET /.well-known/agent-card.json` on the Gateway returns the **mesh card**: `name` is the
  mesh, `skills[]` is the union of registered skills the caller's token may see, each skill's
  `metadata["agenet.io/agents"]` lists the providing agent names, and
  `capabilities.extensions[]` advertises the routing extension. Unauthenticated callers get
  the public subset; `GetExtendedAgentCard` returns the full view after auth.
- `GET /agents/{name}/.well-known/agent-card.json` returns the agent's own card with
  `supportedInterfaces[]` rewritten to the virtual endpoint and `securitySchemes` replaced by
  the Gateway's. A pure-A2A client uses this and never needs the extension.
- `GET /skills?tag=finance&inputMode=application/pdf` is a convenience search over the same
  index. Results carry `ttlMs` and `cacheScope`, matching MCP's `CacheableResult`.

## Feeding external registries

The Dispatcher can export its live view as:

- MCP Registry `server.json` entries for tool agents (and for any agent via the MCP facade),
  pushed to a private sub-registry.
- AGNTCY OASF records, if a tenant runs an Agent Directory.

The mesh itself is not a registry of record; it only knows what is attached right now.
