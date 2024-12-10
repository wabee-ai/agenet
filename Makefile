# Paths
PROTO_SRC=api/proto
PROTO_OUT=pkg/grpc

# Tools
PROTOC=protoc
PROTOC_GEN_GO=protoc-gen-go
PROTOC_GEN_GO_GRPC=protoc-gen-go-grpc

# Commands
GEN_PROTO=$(PROTOC) --proto_path=$(PROTO_SRC) \
	--go_out=$(PROTO_OUT) \
	--go-grpc_out=$(PROTO_OUT) \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative

# Default target
all: upgrade-deps proto

# Generate Go implementations from proto files
proto:
	$(GEN_PROTO) $(PROTO_SRC)/common/*.proto
	$(GEN_PROTO) $(PROTO_SRC)/gateway/*.proto
	$(GEN_PROTO) $(PROTO_SRC)/dispatcher/*.proto

# Clean generated files
clean:
	rm -rf $(PROTO_OUT)

# Dependency management
check-upgrade-deps:
	go list -u -m all

test-upgrade-deps:
	go get -t -u ./...

upgrade-deps:
	go get -u ./...
	go mod tidy
	go mod vendor

v:
	go mod tidy
	go mod vendor

# Testing
test:
	go test ./...

# Help target
help:
	@echo "Usage:"
	@echo "  make proto                - Generate Go implementations for proto files"
	@echo "  make clean                - Remove all generated files"
	@echo "  make check-upgrade-deps   - Check for available dependency upgrades"
	@echo "  make test-upgrade-deps    - Test upgrading dependencies"
	@echo "  make upgrade-deps         - Upgrade dependencies and tidy up modules"
	@echo "  make v                    - Run go mod tidy and go mod vendor"
	@echo "  make test                 - Run all tests"
	@echo "  make help                 - Show this help message"