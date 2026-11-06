# The agentic landscape since Agenet v1 (Dec 2024 → Nov 2026)

Agenet v1 was written in November–December 2024, two weeks after Anthropic open-sourced the
Model Context Protocol and four months before Google announced A2A. At that time "agent" in
this repo meant a *service agent*: a worker process that handles one `requestType`. There was
no shared wire format for agents, no agent identity model, and no standard for long-running
work. All three now exist, are governed by one foundation, and are shipping in every major
cloud. This document records what changed and what it implies for a v2.

Dates and claims below were gathered in November 2026 from primary sources where available.
Items marked *(secondary)* come from vendor blogs or press and were not confirmed against
a primary release note.

## 1. Timeline

| Date | Event | Why it matters for Agenet |
|---|---|---|
| 2024-11-25 | MCP launches (rev `2024-11-05`): JSON-RPC 2.0, stdio + HTTP/SSE, tools/resources/prompts, `initialize` handshake. [spec](https://modelcontextprotocol.io/specification/2024-11-05) | Established JSON-RPC 2.0 as the agent-ecosystem envelope. |
| 2025-03 | IBM ACP, Cisco AGNTCY, OpenAI Agents SDK all launch. | Start of the "agent protocol" race. |
| 2025-03-26 | MCP rev 2: OAuth 2.1, Streamable HTTP (single POST endpoint, `Mcp-Session-Id`, `Last-Event-ID` resumption), tool annotations (`readOnlyHint`, `destructiveHint`, `idempotentHint`). [changelog](https://modelcontextprotocol.io/specification/2025-03-26/changelog) | Stateful sessions over HTTP, structurally similar to Agenet's bidi streams. |
| 2025-04-09 | Google announces A2A with 50+ partners. | First horizontal agent-to-agent standard. |
| 2025-04 | Invariant Labs names *tool poisoning*; May 2025 GitHub MCP private-repo exfiltration via prompt injection. [timeline](https://authzed.com/blog/timeline-mcp-breaches) | Descriptions and payloads are attack surface. |
| 2025-06-18 | MCP rev 3: servers are OAuth resource servers (RFC 9728), clients MUST send RFC 8707 resource indicators, structured tool output, elicitation, Security Best Practices (confused deputy, token passthrough, session hijacking). [changelog](https://modelcontextprotocol.io/specification/2025-06-18/changelog) | Audience-bound tokens per hop became the norm. |
| 2025-06 | A2A donated to the Linux Foundation. | |
| 2025-07-30 | A2A 0.3: gRPC binding, signed Agent Cards (JWS over RFC 8785 canonical JSON), `agent-card.json`. [changelog](https://github.com/a2aproject/A2A/blob/main/CHANGELOG.md) | A2A now has a protobuf schema Agenet can import. |
| 2025-07-29 | AGNTCY (OASF schema, Agent Directory, SLIM gRPC transport) donated to LF. [press](https://www.linuxfoundation.org/press/linux-foundation-welcomes-the-agntcy-project-to-standardize-open-multi-agent-system-infrastructure-and-break-down-ai-agent-silos) | SLIM is the closest analog to Agenet and positions itself as an A2A *binding*. |
| 2025-08-25 | agentgateway (Solo.io) donated to LF; speaks MCP and A2A natively. [blog](https://agentgateway.dev/blog/2025-08-25-solo-contributes-agentgateway-to-lf/) | "Agent gateway" is now a product category. |
| 2025-08-29 | IBM ACP merges into A2A. [discussion](https://github.com/orgs/i-am-bee/discussions/5) | Consolidation: one horizontal protocol. |
| 2025-09-08 | MCP Registry preview (`server.json`, sub-registries). [blog](https://blog.modelcontextprotocol.io/posts/2025-09-08-mcp-registry-preview/) | |
| 2025-10-13 | AWS Bedrock AgentCore GA: Runtime (microVM per session, 8h), Gateway, Identity (token vault), Memory, Observability; Policy (Cedar) preview Dec 2025. [blog](https://aws.amazon.com/blogs/aws/introducing-amazon-bedrock-agentcore-securely-deploy-and-operate-ai-agents-at-any-scale) | Clouds bundle routing + identity + policy + tracing as one layer. |
| 2025-11-25 | MCP rev 4: experimental Tasks (`working/input_required/completed/failed/cancelled`), sampling with tools, Client ID Metadata Documents, extensions. [changelog](https://modelcontextprotocol.io/specification/2025-11-25/changelog) | Long-running work gets a task handle. |
| 2025-12-09 | **Agentic AI Foundation (AAIF)** formed under LF by Anthropic (MCP), OpenAI (AGENTS.md), Block (goose); platinum members include AWS, Google, Microsoft, Cloudflare. [blog](https://blog.modelcontextprotocol.io/posts/2025-12-09-mcp-joins-agentic-ai-foundation/) | Single governance home for agent standards. |
| 2025-12-09 | OWASP Top 10 for Agentic Applications: ASI03 identity/privilege abuse, ASI07 insecure inter-agent communication, ASI08 cascading failures. | Routing layers are explicitly in scope. |
| 2026-01-26 | MCP Apps (SEP-1865) finalized as a Stable extension. | |
| 2026-03-12 | **A2A 1.0** ships; TSC of AWS, Cisco, Google, IBM, Microsoft, Salesforce, SAP, ServiceNow; SDKs in Python, Go, Java, JS, .NET, Rust. [post](https://a2a-protocol.org/latest/blog/2026/03/12/a2a-protocol-ships-v10-production-ready-standard-for-agent-to-agent-communication/) | The agent-to-agent contract is now stable. |
| 2026-03-23 | Temporal + OpenAI Agents SDK integration GA. *(secondary)* | Durable execution engines drive agent retries. |
| 2026-04 | Microsoft Agent Framework 1.0 GA (MCP + A2A); Anthropic Managed Agents public beta; Google rebrands Vertex Agent Engine to Gemini Enterprise Agent Platform. *(secondary)* | Vendor SDKs consume MCP/A2A; none defines a new wire format. |
| 2026-05 | IETF ID-JAG draft 04 (identity assertion authorization grant); shipped as Okta Cross App Access, Auth0, Descope in Aug–Sep 2026. [Okta](https://developer.okta.com/blog/2025/09/03/cross-app-access) | Standard on-behalf-of delegation for agents. |
| 2026-06-12 | OTel GenAI semantic conventions v1.42.0 add MCP conventions; `invoke_agent`, `execute_tool`, `gen_ai.usage.*` attributes. [repo](https://github.com/open-telemetry/semantic-conventions-genai) | A shared vocabulary for agent traces. |
| 2026-06-23 | Envoy AI Gateway 1.0 GA (MCP aggregation, CEL authz); renamed Agent Router, joined AAIF 2026-09-10; A2A fronting Sep 2026. [release](https://tetrate.io/blog/envoy-ai-gateway-v1-0-release) | |
| **2026-07-28** | **MCP rev 5, "stateless"**: sessions and `initialize` removed, every request self-describes in `_meta`, `server/discover`, Multi Round-Trip Requests (`resultType: input_required` + sealed `requestState`), Tasks moved to an extension with polling, `Mcp-Method`/`Mcp-Name` headers for L7 routing, OTel `traceparent` in `_meta`, Roots/Sampling/Logging deprecated, 12-month feature lifecycle. [changelog](https://modelcontextprotocol.io/specification/2026-07-28/changelog) | **The strongest signal for Agenet**: the ecosystem tried stateful bidirectional sessions for 18 months and abandoned them as the protocol core. |
| 2026-08-27 | A2A joins AAIF. [blog](https://a2a-protocol.org/latest/blog/2026/08/27/a-new-chapter-for-a2a-joining-the-agentic-ai-foundation/) | MCP and A2A under one roof. |
| 2026-09-15 | A2A roadmap: v1.1 task timeline and event filtering, bidirectional streaming, elicitation. [roadmap](https://a2a-protocol.org/latest/roadmap) | |
| 2026-10-12 | AGNTCY proposes joining AAIF. [proposal](https://github.com/aaif/project-proposals/issues/57) | |

## 2. What the ecosystem converged on

1. **Two protocols, one foundation.** MCP is vertical (agent → tools and data). A2A is horizontal
   (agent → agent). Everything else either merged into them (IBM ACP), positioned itself as a
   binding under them (AGNTCY SLIM), or stayed academic (ANP, Agora, NLIP).
2. **JSON-RPC 2.0 envelope with gRPC and REST projections.** A2A defines each operation once with
   three bindings and standardizes errors on `google.rpc.Status`.
3. **Stateless requests plus durable task handles.** MCP removed sessions; A2A never had them.
   Long work is a `Task` with `GetTask`, `SubscribeToTask`, `CancelTask`, and webhook push.
   Streams are a delivery optimization, not the contract.
4. **A declared capability catalog instead of opaque routing keys.** A2A `AgentCard.skills[]`
   with `id`, `tags`, `inputModes`, `outputModes`; MCP `tools/list` with JSON Schema. Gateways
   (Agent Router, Kong, agentgateway, AgentCore Gateway) filter and govern on that catalog.
5. **Audience-bound identity on every hop.** OAuth 2.1 resource servers, RFC 8707 resource
   indicators, no token passthrough, token exchange or ID-JAG for delegation, SPIFFE for
   workload identity. Agent Cards are signed. Session IDs are not identity.
6. **Pause states for humans and auth.** A2A `INPUT_REQUIRED` and `AUTH_REQUIRED`; MCP
   elicitation and MRTR; LangChain `interrupted`.
7. **Tracing and token accounting as metadata.** W3C `traceparent` carried in `_meta`;
   OTel `invoke_agent` CLIENT/SERVER spans; `gen_ai.usage.input_tokens/output_tokens` per hop.
8. **Payload is untrusted input.** Tool poisoning, the "lethal trifecta", OWASP ASI07. Pin
   descriptors by hash, keep control fields out of the payload, put a policy decision point in
   the call path (Cedar, OPA, CEL).

## 3. Where Agenet v1 stands against that

| Agenet v1 (Dec 2024) | 2026 expectation | Gap |
|---|---|---|
| `requestType` string | `AgentCard.skills[].id` + tags + MIME modes, versioned | No catalog, no schemas, no versioning |
| `uuid` doubles as response channel | `taskId` (unit of work) + `contextId` (conversation) + `messageId` (idempotency) | One ID doing three jobs |
| `data` + `content-type` | `Message{parts[]}` and `Artifact{parts[]}`; `Part = text|raw|url|data` | No multimodal or multi-part payloads |
| delivered → partial* → completed/error | `SUBMITTED, WORKING, INPUT_REQUIRED, AUTH_REQUIRED, COMPLETED, FAILED, CANCELED, REJECTED` | No pause, cancel, or reject |
| Persistent response channel "until delivered" | Durable task store; `GetTask`/`SubscribeToTask`/push survive disconnects | Correctness tied to stream liveness |
| `metadata.clientId` | Authenticated principal, delegation chain, tenant, audience-bound token per hop | No identity model at all |
| None | Policy decision point on structured attributes | No authorization |
| None | `traceparent`, OTel GenAI spans, usage per hop | No observability |
| gRPC only, bespoke JSON | A2A gRPC/JSON-RPC/REST bindings; MCP Streamable HTTP | Not interoperable with anything |
| Agents dial out to the Dispatcher | A2A assumes agents are reachable HTTP servers | **v1's genuine differentiator**: agents behind NAT, on laptops, at the edge |

## 4. Conclusion for v2

A bespoke v2 message format would compete with A2A and lose. The useful move is the one
AGNTCY SLIM made: become a **transport binding and routing fabric for A2A 1.0**, keep the
dial-out registration model that A2A lacks, and add the operational layer that A2A and MCP
deliberately leave to infrastructure: durable tasks, brokered identity, policy, tracing,
backpressure, and an MCP bridge. See [`../v2/00-overview.md`](../v2/00-overview.md).
