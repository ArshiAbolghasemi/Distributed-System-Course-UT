RAFT_DIR = ../src/raft1

.PHONY: \
	test test-race \
	test-leader-election test-leader-election-race \
	test-log test-log-race \
	test-persister test-persister-race \
	test-snapshot test-snapshot-race \
	help

help:
	@echo "Raft Test Targets:"
	@echo "  make test                      - Run all tests"
	@echo "  make test-race                 - Run all tests with race detection"
	@echo "  make test-leader-election      - Run leader election test (3A)"
	@echo "  make test-leader-election-race - Run leader election with race detection"
	@echo "  make test-log                  - Run log test (3B)"
	@echo "  make test-log-race             - Run log with race detection"
	@echo "  make test-persister            - Run persister test (3C)"
	@echo "  make test-persister-race       - Run persister test with race detection"
	@echo "  make test-snapshot             - Run snapshot test (3D)"
	@echo "  make test-snapshot-race        - Run snapshot with race detection"

test:
	cd $(RAFT_DIR) && go test -v -run .

test-race:
	cd $(RAFT_DIR) && go test -race -v -run .

test-leader-election:
	cd $(RAFT_DIR) && go test -v -run 3A

test-leader-election-race:
	cd $(RAFT_DIR) && go test -race -v -run 3A

test-log:
	cd $(RAFT_DIR) && go test -v -run 3B

test-log-race:
	cd $(RAFT_DIR) && go test -race -v -run 3B

test-persister:
	cd $(RAFT_DIR) && go test -v -run 3C

test-persister-race:
	cd $(RAFT_DIR) && go test -race -v -run 3C

test-snapshot:
	cd $(RAFT_DIR) && go test -v -run 3D

test-snapshot-race:
	cd $(RAFT_DIR) && go test -race -v -run 3D
