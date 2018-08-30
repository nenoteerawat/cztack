SHELL := /bin/bash
MODULES=$(filter-out vendor/ module-template/ scripts/ testutil/,$(sort $(dir $(wildcard */))))
TEST :=./...
export PRIVATE_SUBNETS :=subnet-0e74698925a68c650,subnet-0f6ea862112b067c8
export VPC_ID :=vpc-0442f170b88f8eaf6

all: clean fmt docs lint test

fmt:
	@for m in $(MODULES); do \
		terraform fmt $m; \
	done

lint:
	@for m in $(MODULES); do \
		terraform fmt -check $$m || exit $$?; \
	done;

	@for m in $(MODULES); do \
		ls $$m/*_test.go 2>/dev/null 1>/dev/null || (echo "no test(s) for $$m"; exit $$?); \
	done

docs:
	@for m in $(MODULES); do \
		pushd $$m; \
		../scripts/update-readme.sh update; \
		popd; \
	done;

check-docs:
	@for m in $(MODULES); do \
		pushd $$m; \
		../scripts/update-readme.sh check || exit $$?; \
		popd; \
	done;

clean:
		rm **/*.tfstate*

test: fmt
	GOCACHE=off AWS_PROFILE=cztack-ci-1 go test $(TEST)
