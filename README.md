
# Agenet - Agent Network

## Introduction
Agenet, short for **Agent Network**, is inspired by Usenet's concept of a **User Network**. Just as Usenet created a vast, decentralized system for sharing information between users, Agenet is designed to facilitate seamless communication and task routing among service agents. This system enables clients to send specific requests without needing knowledge of underlying dependencies, service locations, or infrastructure.

Agenet promotes a modular, decoupled, and scalable architecture where each service is independently developed, managed, and maintained. Through this network, clients and agents communicate efficiently, focusing solely on requests while Agenet handles all complexities of routing, processing, and response.

## Documentation Index

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

---

This project offers a robust, scalable, and flexible approach to routing client requests, leveraging real-time communication and unique message tracking to enhance reliability, adaptability, and overall system simplicity.

## Setting Up the Environment for Protocol Buffers (Protobuf)

To generate Go implementations for the `.proto` files in this project, you need to configure your environment by installing the Protocol Buffers compiler (`protoc`) and the Go plugins.

### Installation Steps

#### 1. Install the Protocol Buffers Compiler (`protoc`)
The `protoc` compiler is required to process `.proto` files and generate code for various languages, including Go.

##### On Linux
```bash
sudo apt update
sudo apt install -y protobuf-compiler
```

##### On macOS
```bash
brew install protobuf
```

##### On Windows
1. Download the latest release from the [Protocol Buffers GitHub page](https://github.com/protocolbuffers/protobuf/releases).
2. Extract the downloaded archive.
3. Add the path to the `protoc` executable to your system's `PATH` variable.

##### Verify Installation
After installing, run the following command to confirm `protoc` is installed:
```bash
protoc --version
```

#### 2. Install Go Plugins for Protobuf
You need two plugins to generate Go code for gRPC and Protocol Buffers:

##### Install the Protobuf Go Plugin
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

##### Install the gRPC Go Plugin
```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

##### Add Plugins to System Path
Ensure the `$GOPATH/bin` is included in your `PATH` environment variable:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

##### Verify Plugin Installation
Run the following commands to verify the plugins are installed correctly:
```bash
protoc-gen-go --version
protoc-gen-go-grpc --version
```

### 3. Test the Setup with the Makefile
Once the environment is set up, you can use the provided `Makefile` to generate Go implementations for the `.proto` files.

#### Generate Code
Run the following command from the project root:
```bash
make proto
```

#### Clean Generated Files
To remove all generated files, run:
```bash
make clean
```

### Troubleshooting
- If `protoc` is not found, double-check that it is installed and included in your `PATH`.
- Ensure that the Go plugins are installed and available in your `PATH`.

With this setup, you are ready to work with Protocol Buffers and generate Go code for the project's `.proto` files.