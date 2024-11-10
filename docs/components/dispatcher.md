
# Dispatcher Component

The **Dispatcher** in Agenet acts as the intermediary between the Gateway and the specific Agent responsible for processing a request. It listens to the designated request channels and ensures that each request reaches the correct Agent based on its **RequestType**. The Dispatcher also maintains a bidirectional gRPC stream with each Agent, enabling efficient and reliable communication.

## Key Responsibilities

### 1. Establishing Connection with Agent
When an Agent connects to the Dispatcher via a bidirectional gRPC stream, it specifies the **RequestType** it is responsible for processing. This connection allows the Dispatcher to route incoming requests to the appropriate Agent in real time.

### 2. Request Handling and Message Delivery Notification
Once a request is routed to the Dispatcher, it is forwarded to the connected Agent responsible for that specific **RequestType**. The Dispatcher sends a "message delivered" notification on the channel associated with the request’s **UUID** to indicate to the client that the request has been successfully handed over to the Agent.

### 3. Agent Response and Completion Handling
As the Agent processes the request, it may send one or multiple responses back to the Dispatcher via the gRPC stream. Each response is forwarded to the UUID-based response channel. When the Agent completes processing, it sends a "completed" message to the Dispatcher, which is then routed to the response channel, signaling the end of the response flow.

### 4. Error Handling
If an error occurs while the Agent processes the request, the Agent sends an error message to the Dispatcher through the gRPC stream. The Dispatcher then forwards this error to the UUID-based response channel, allowing the client to receive detailed information about the issue.

## Example Request Flow with Agent Communication

To illustrate how the Dispatcher operates with an Agent, here’s an example workflow for a "Retrieve File" request:

1. **Agent Connection**: The Agent responsible for `"retrieve_file"` connects to the Dispatcher via a gRPC stream, specifying its **RequestType**.
2. **Request Forwarding**: 
   - **Gateway** receives a request with RequestType `"retrieve_file"` and routes it to the corresponding request channel.
   - **Dispatcher** listens on the request channel for `"retrieve_file"` and forwards the request to the connected Agent.
3. **Message Delivered Notification**: Dispatcher sends a "message delivered" notification on the UUID-based response channel.
4. **Processing and Response Flow**: 
   - **Agent** processes the request and may send one or more responses to the Dispatcher through the gRPC stream.
   - Each response is routed to the UUID-based response channel for client retrieval.
5. **Completion or Error**: 
   - **Completion**: When processing completes, the Agent sends a "completed" message to the Dispatcher, which is forwarded to the response channel.
   - **Error**: If an error occurs, the Agent sends an error message to the Dispatcher, which is also forwarded to the response channel.

## Example of Request Attributes Handled by Dispatcher

To clarify the request structure managed by the Dispatcher, here’s an example payload:

```json
{
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "requestType": "retrieve_file",
    "content-type": "application/json",
    "data": {
        "filePath": "/path/to/file.txt",
        "options": {
            "readOnly": true,
            "cache": false
        }
    },
    "metadata": {
        "clientId": "12345",
        "timestamp": "2024-11-06T15:30:00Z"
    }
}
```

- **uuid**: Unique identifier to track the request throughout the system.
- **requestType**: Specifies the type of request for correct Agent routing.
- **content-type**: Specifies the format of the data (e.g., `application/json` or `text/plain`).
- **data**: Contains the core payload for the request.
- **metadata**: Additional information for tracking and context.

## Benefits of the Dispatcher’s Role in Agenet

- **Efficient Request Handling**: By categorizing and routing requests by RequestType, the Dispatcher ensures each Agent only processes relevant requests.
- **Real-Time Communication**: The bidirectional gRPC stream enables seamless interaction between Dispatcher and Agent, supporting dynamic request handling and response delivery.
- **Precise Response Tracking**: UUIDs allow the Dispatcher to track responses accurately, ensuring clients receive the correct responses without collisions.
- **Scalable Architecture**: The Dispatcher’s modular design allows for easy scaling, with additional Dispatchers able to handle more request types or increased load.

The Dispatcher component is a key part of Agenet's architecture, facilitating organized, real-time request processing and response management.
