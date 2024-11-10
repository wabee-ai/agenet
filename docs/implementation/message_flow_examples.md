
# Message Flow Examples

To better understand the inner workings of Agenet, this section provides example flows showing how requests are routed from clients to agents, processed, and returned.

## Example Flow: "Retrieve File" Request

In this example, a client sends a **Retrieve File** request. This flow illustrates how the system manages the request and routes it to the appropriate agent.

1. **Client** sends a `"Retrieve File"` request to the **Gateway** via a bidirectional gRPC connection.
2. **Gateway** generates a unique ID (UUID) for the request if the client didn’t provide one and assigns a timestamp to the request.
3. **Gateway** routes the request to the **Request Channel** specifically designated for `"Retrieve File"`.
4. **Gateway** then establishes a persistent connection to a **Response UUID Channel** uniquely identified by the UUID, where it will listen for a response tied to the request ID.
5. **Dispatcher** connects to the **Request Channel** for `"Retrieve File"`, where it receives the request and forwards it to the **Agent**.
6. **Agent** processes the request by performing the necessary operation (e.g., retrieving the file) and prepares a response.
7. **Agent** sends the response back to the **Dispatcher**.
8. **Dispatcher** routes the response to the **UUID-based Response Channel**, associating it with the correct request ID for tracking.
9. **Gateway** listens on the **UUID-based Response Channel**, where it identifies the response by the request ID.
10. **Gateway** forwards the response to the **Client** through the gRPC connection, completing the cycle.

## Example Flow Diagram

The following diagram illustrates this flow in a visual format.

```mermaid
flowchart TB
    Client([Client: sends 'Retrieve File' request]):::client
    Gateway([Gateway]):::component
    Dispatcher([Dispatcher]):::component
    Agent([File Agent]):::externalAgent

    RequestChannel([Request Channel - Retrieve File]):::queue
    ResponseUUIDChannel([Response UUID Channel]):::queue

    Step1((1)):::step
    Step2((2)):::step
    Step3((3)):::step
    Step4((4)):::step
    Step5((5)):::step
    Step6((6)):::step
    Step7((7)):::step
    Step8((8)):::step
    Step9((9)):::step

    Client --> Step1
    Step1 -->|Send Request: 'Retrieve File'| Gateway
    Gateway --> Step2
    Step2 -->|Route to 'Retrieve File' Request Channel| RequestChannel
    Gateway -.-> Step3
    Step3 -.->|Connect to Response UUID Channel| ResponseUUIDChannel
    RequestChannel -.-> Step4
    Step4 -.->|Connect to 'Retrieve File' Request Channel| Dispatcher
    Dispatcher --> Step5
    Step5 -->|Forward Request to File Agent| Agent
    Agent --> Step6
    Step6 -->|Return Response to Dispatcher| Dispatcher
    Dispatcher --> Step7
    Step7 -->|Route to Response UUID Channel| ResponseUUIDChannel
    ResponseUUIDChannel -.-> Step8
    Step8 -.->|Receive Response| Gateway
    Gateway --> Step9
    Step9 -->|Forward Response to Client| Client

    %% Style definitions
    classDef client fill:#4CAF50,stroke:#333,stroke-width:2px;
    classDef component fill:#2196F3,stroke:#333,stroke-width:2px;
    classDef queue fill:#D2691E,stroke:#333,stroke-width:2px;
    classDef externalAgent fill:#8A2BE2,stroke:#333,stroke-width:2px;
    classDef step fill:#FF0000,stroke:#333,stroke-width:2px,color:#FFFFFF;

    class Client client;
    class Gateway component;
    class Dispatcher component;
    class Agent externalAgent;
    class RequestChannel queue;
    class ResponseUUIDChannel queue;
    class Step1,Step2,Step3,Step4,Step5,Step6,Step7,Step8,Step9 step;
```

This example demonstrates Agenet’s ability to manage request routing and response delivery effectively, maintaining a persistent connection from client to agent and back.
