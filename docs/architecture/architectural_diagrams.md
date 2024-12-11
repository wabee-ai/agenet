
# Architectural Diagrams

The following diagram provides a high-level overview of the Agenet system architecture, illustrating the core components and general flow of communication between them.

## High-Level System Architecture

```mermaid
flowchart TB
    Client([Client]):::client
    Gateway([Gateway]):::component
    Dispatcher([Dispatcher]):::component
    Agent([Agent]):::externalAgent
    
    RequestChannel([Request Channel]):::queue
    ResponseUUIDChannel([Response UUID Channel]):::queue

    %% General communication flow without step details
    Client -->|Sends Request| Gateway
    Gateway -->|Routes Request to Request Channel| RequestChannel
    Gateway -.->|Connects to Response UUID Channel| ResponseUUIDChannel
    RequestChannel -.->|Dispatcher Connects to Request Channel| Dispatcher
    Dispatcher -->|Forwards Request to Agent| Agent
    Agent -->|Returns Response| Dispatcher
    Dispatcher -->|Routes Response to Response UUID Channel| ResponseUUIDChannel
    ResponseUUIDChannel -.->|Gateway Receives Response| Gateway
    Gateway -->|Forwards Response to Client| Client

    %% Style definitions
    classDef client fill:#4CAF50,stroke:#333,stroke-width:2px;
    classDef component fill:#2196F3,stroke:#333,stroke-width:2px;
    classDef queue fill:#D2691E,stroke:#333,stroke-width:2px;
    classDef externalAgent fill:#8A2BE2,stroke:#333,stroke-width:2px;

    class Client client;
    class Gateway component;
    class Dispatcher component;
    class Agent externalAgent;
    class RequestChannel queue;
    class ResponseUUIDChannel queue;
```

## Summary
This high-level architecture diagram shows the main components of Agenet and their general communication pathways. The Client interacts with the Gateway to send requests. The Gateway, in turn, routes requests to the appropriate Request Channel, while keeping a connection to a Response UUID Channel. The Dispatcher listens on the Request Channel, forwarding requests to the relevant Agent based on the request type.

Upon receiving a response from an Agent, the Dispatcher routes the response back through the Response UUID Channel, where the Gateway retrieves it and forwards it to the original Client.

These interactions illustrate Agenet's capability to decouple client requests from specific agent details, supporting a scalable and modular system.
