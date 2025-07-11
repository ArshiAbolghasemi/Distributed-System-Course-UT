.PHONY: test

test:
	@echo "Runnig mapreduce tests..."
	$(MAKE) -C mapreduce test
	@echo "Runnig key value server tests..."
	$(MAKE) -C kvserver test
	@echo "Runnig raft tests..."
	$(MAKE) -C raft test

