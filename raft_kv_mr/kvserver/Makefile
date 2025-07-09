KVSRV_DIR = ../src/kvsrv1
LOCK_DIR = $(KVSRV_DIR)/lock

.PHONY: test test-reliable test-lock test-lock-reliable help

test:
	@echo "Running basic KV server tests..."
	cd $(KVSRV_DIR) && go test -v

test-reliable:
	@echo "Running reliable KV server tests..."
	cd $(KVSRV_DIR) && go test -v -run reliable

test-lock:
	@echo "Running basic lock tests..."
	cd $(LOCK_DIR) && go test -v

test-lock-reliable:
	@echo "Running reliable lock tests..."
	cd $(LOCK_DIR) && go test -v -run reliable

help:
	@echo "Available targets:"
	@echo "  test              - Run basic KV server tests"
	@echo "  test-reliable     - Run reliable KV server tests"
	@echo "  test-lock         - Run basic lock tests"
	@echo "  test-lock-reliable - Run reliable lock tests"
	@echo "  help              - Show this help message"
