SRC_MAIN_DIR = ./src/main
SRC_MRAPPS_DIR = ./src/mrapps

APPS = wc indexer crash nocrash rtiming mtiming jobcount early_exit

.PHONY: clean build-run run-sequential

$(APPS):
	@echo "Building $@.so plugin..."
	cd $(SRC_MAIN_DIR) && go build -buildmode=plugin ../mrapps/$@.go
	@echo "$@.so built successfully"

run-sequential:
	@if [ -z "$(APP)" ]; then \
		echo "Please specify an application. Usage: make run-sequential APP=wc"; \
		exit 1; \
	fi
	@echo "Running sequential MapReduce with $(APP)..."
	cd $(SRC_MAIN_DIR) && rm -f mr-out*
	cd $(SRC_MAIN_DIR) && go run mrsequential.go $(APP).so pg-*.txt
	@echo "Output available in $(SRC_MAIN_DIR)/mr-out-0"
	cd $(SRC_MAIN_DIR) && more mr-out-0

run-coordinator:
	cd $(SRC_MAIN_DIR) && rm -f mr-out*
	cd $(SRC_MAIN_DIR) && go run mrcoordinator.go pg-*.txt

run-worker:
	@if [ -z "$(APP)" ]; then \
		echo "Please specify an application. Usage: make run-sequential APP=wc"; \
		exit 1; \
	fi
	@echo "Running worker MapReduce woth $(APP)..."
	cd $(SRC_MAIN_DIR) && go run mrworker.go $(APP).so
	@echo "Output available in $(SRC_MAIN_DIR)/mr-out-*"
	cd $(SRC_MAIN_DIR) && cat mr-out-* | sort | more

test:
	cd $(SRC_MAIN_DIR) && ./test-mr.sh
	$(MAKE) clean

clean:
	@echo "Cleaning up..."
	@cd $(SRC_MAIN_DIR) && rm -f mr-out* *.so && rm -f mr-[0-9]*-[0-9]*.json && rm -rf mr-tmp && \
		rm -f mrcoordinator && rm -f mrworker && rm -f mrsequential
	@cd $(SRC_MRAPPS_DIR) && rm -f *.so
	@echo "Clean complete"

help:
	@echo "Available targets:"	
	@echo "  run-sequential  - Run sequential MapReduce with specified app (make run-sequential APP=wc)"	
	@echo "  clean           - Remove generated files"
	@echo "  run-coordinator - Run MapReduce coordinator"
	@echo "  run-wroker      - Run worker MapReduce with specified app (make run-worker APP=wc)"
	@echo "  test            - run tests"
	@echo "  clean           - clean redundant files"
