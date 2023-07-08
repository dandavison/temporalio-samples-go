############################# Main targets #############################
# Run all checks, build, and test.
install: clean staticcheck errcheck workflowcheck bins test
########################################################################

##### Variables ######
UNIT_TEST_DIRS := $(sort $(dir $(shell find . -name "*_test.go")))
TEST_TIMEOUT := 20s
COLOR := "\e[1;36m%s\e[0m\n"

define NEWLINE


endef

# If the first argument is "bins"...
ifeq (bins,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "bin"
  BIN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(BIN_ARGS):;@:)
endif

ifeq ($(strip $(BIN_ARGS)),)
	BIN_ARGS := .
endif
MAIN_FILES := $(shell find $(BIN_ARGS) -name "main.go")

##### Targets ######
.PHONY: bins
bins:
	@printf $(COLOR) "Build sample for $(BIN_ARGS)"
	$(foreach MAIN_FILE,$(MAIN_FILES), go build -o ./bin/$(shell dirname "$(MAIN_FILE)") ./$(shell dirname "$(MAIN_FILE)")$(NEWLINE))

test:
	@printf $(COLOR) "Run unit tests..."
	@rm -f test.log
	$(foreach UNIT_TEST_DIR,$(UNIT_TEST_DIRS),\
		@go test -timeout $(TEST_TIMEOUT) -race $(UNIT_TEST_DIR) | tee -a test.log \
	$(NEWLINE))
	@! grep -q "^--- FAIL" test.log

staticcheck:
	@printf $(COLOR) "Run static check..."
	@go install honnef.co/go/tools/cmd/staticcheck@2022.1.3
	@staticcheck ./...

errcheck:
	@printf $(COLOR) "Run error check..."
	@GO111MODULE=off go get -u github.com/kisielk/errcheck
	@errcheck ./...

workflowcheck:
	@printf $(COLOR) "Run workflow check..."
	@go install go.temporal.io/sdk/contrib/tools/workflowcheck
	@workflowcheck -show-pos ./...

update-sdk:
	go get -u go.temporal.io/api@main
	go get -u go.temporal.io/sdk@main
	go mod tidy

clean:
	rm -rf bin
	
ci-build: staticcheck errcheck workflowcheck bins test
GO111MODULE=off go get -u honnef.co/go/tools/cmd/staticcheck