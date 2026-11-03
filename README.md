
# Agenet - Agent Network

## Introduction
Agenet, short for **Agent Network**, is inspired by Usenet's concept of a **User Network**. Just as Usenet created a vast, decentralized system for sharing information between users, Agenet is designed to facilitate seamless communication and task routing among service agents. This system enables clients to send specific requests without needing knowledge of underlying dependencies, service locations, or infrastructure.

Agenet promotes a modular, decoupled, and scalable architecture where each service is independently developed, managed, and maintained. Through this network, clients and agents communicate efficiently, focusing solely on requests while Agenet handles all complexities of routing, processing, and response.

## Agenet v2 (draft, November 2026)

v1 (December 2024) predates A2A, the Agentic AI Foundation, and every revision of MCP after the first. v2 repositions Agenet as an **agent mesh**: a routing fabric and transport binding for A2A 1.0 that keeps v1's dial-out registration model, replaces the bespoke message format with A2A's, and adds durable tasks, brokered identity, policy, tracing, and an MCP bridge.

- [What changed in the agentic landscape, 2024–2026](./docs/research/landscape-2024-2026.md)
- [00 Overview and principles](./docs/v2/00-overview.md)
- [01 Architecture](./docs/v2/01-architecture.md)
- [02 Wire contract](./docs/v2/02-wire-contract.md)
- [03 Task lifecycle](./docs/v2/03-task-lifecycle.md)
- [04 Discovery and registration](./docs/v2/04-discovery-and-registration.md)
- [05 Identity, policy, observability](./docs/v2/05-identity-policy-observability.md)
- [06 MCP bridge](./docs/v2/06-mcp-bridge.md)
- [07 Migration from v1 and open questions](./docs/v2/07-migration-and-open-questions.md)

## v1 documentation (December 2024, historical)

### 1. Overview
- [Purpose](./docs/overview/purpose.md)
- [Technical Benefits](./docs/overview/technical_benefits.md)
- [Practical Benefits](./docs/overview/practical_benefits.md)

### 2. Architecture
- [System Components](./docs/architecture/system_components.md)
- [Communication Patterns](./docs/architecture/communication_patterns.md)
- [Architectural Diagrams](./docs/architecture/architectural_diagrams.md)

### 3. Implementation Details
- [Message Flow Examples](./docs/implementation/message_flow_examples.md)
- [Component Specifications](./docs/implementation/component_specifications.md)
