
# System Components

Agenet is structured around three core components, each serving a unique role to ensure efficient communication and routing between clients and agents.

## Core Components

### 1. Gateway
The **Gateway** acts as the entry point for client requests. It receives incoming requests from clients, generates a unique ID (UUID) if one is not provided, and routes the request to the appropriate channel in the system.

- **Functionality**: Handles the initial reception of client requests, manages UUID assignment, and initiates message routing.
- **Role**: Serves as the primary interface between clients and Agenet, abstracting the complexities of the backend services from the client side.

### 2. Dispatcher
The **Dispatcher** manages and directs requests to their corresponding agents. It is where agents connect to listen for specific request types, enabling a modular and scalable approach for request handling.

- **Functionality**: Listens for requests from the Gateway and directs them to the appropriate Agent. It ensures requests are delivered to the correct handler, facilitating efficient task distribution.
- **Role**: Acts as the coordinator for requests, maintaining connections with agents and distributing tasks accordingly.

### 3. Agent
The **Agent** is responsible for processing specific requests as required by the client. Each Agent specializes in handling a particular type of request and sends responses back through the Dispatcher to reach the original client.

- **Functionality**: Executes the requested tasks and prepares responses.
- **Role**: The endpoint for request processing within Agenet, ensuring each client request is handled by the appropriate service.

These components work in harmony to provide a seamless routing and task handling system, where the Gateway, Dispatcher, and Agents each play a crucial role in enabling asynchronous, real-time interactions.
