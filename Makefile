# Set the default shell to `/bin/sh` for executing commands in the Makefile.
SHELL = /bin/sh

# Define the project name for easy reference.
PROJECT = "hq-go-http"

# --- Go(Golang) ------------------------------------------------------------------------------------

# Define common Go commands with variables for reusability and easier updates.
GOCMD=go # The main Go command.
GOMOD=$(GOCMD) mod # Go mod command for managing modules.
GOGET=$(GOCMD) get # Go get command for retrieving packages.
GOFMT=$(GOCMD) fmt # Go fmt command for formatting Go code.
GOTEST=$(GOCMD) test # Go test command for running tests.

# Define Go build flags for verbosity and linking.
GOFLAGS := -v # Verbose flag for Go commands to print detailed output.
LDFLAGS := -s -w # Linker flags to strip debug information (-s) and reduce binary size (-w).

# Set static linking flags for systems that are not macOS (darwin).
# Static linking allows the binary to include all required libraries in the executable.
ifneq ($(shell go env GOOS),darwin)
	LDFLAGS := -extldflags "-static"
endif

# Define Golangci-lint command for linting Go code.
GOLANGCILINTCMD=golangci-lint
GOLANGCILINTRUN=$(GOLANGCILINTCMD) run

# --- Go Module Management

# Tidy Go modules
# This target cleans up `go.mod` and `go.sum` files by removing any unused dependencies.
# Use this command to ensure that the module files are in a clean state.
.PHONY: go-mod-tidy
go-mod-tidy:
	$(GOMOD) tidy

# Update Go modules
# This target updates the Go module dependencies to their latest versions.
# It fetches and updates all modules, and any indirect dependencies.
.PHONY: go-mod-update
go-mod-update:
	$(GOGET) -f -t -u ./... # Update test dependencies.
	$(GOGET) -f -u ./... # Update other dependencies.

# --- Go Code Quality and Testing

# Format Go code
# This target formats all Go source files according to Go's standard formatting rules using `go fmt`.
.PHONY: go-fmt
go-fmt:
	$(GOFMT) ./...

# Lint Go code
# This target lints the Go source code to ensure it adheres to best practices.
# It uses `golangci-lint` to run various static analysis checks on the code.
# It first runs the `go-fmt` target to ensure the code is properly formatted.
.PHONY: go-lint
go-lint: go-fmt
	$(GOLANGCILINTRUN) $(GOLANGCILINT) ./...

# Run Go tests
# This target runs all Go tests in the current module.
# The `GOFLAGS` flag ensures that tests are run with verbose output, providing more detailed information.
.PHONY: go-test
go-test:
	$(GOTEST) $(GOFLAGS) ./...

# --- Help -----------------------------------------------------------------------------------------

# Display help information
# This target prints out a detailed list of all available Makefile commands for ease of use.
# It's a helpful reference for developers using the Makefile.
.PHONY: help
help:
	@echo ""
	@echo "*****************************************************************************"
	@echo ""
	@echo "PROJECT : $(PROJECT)"
	@echo ""
	@echo "*****************************************************************************"
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo " Go Commands:"
	@echo "  go-mod-tidy .............. Tidy Go modules."
	@echo "  go-mod-update ............ Update Go modules."
	@echo "  go-fmt ................... Format Go code."
	@echo "  go-lint .................. Lint Go code."
	@echo "  go-test .................. Run Go tests."
	@echo ""
	@echo " Help Commands:"
	@echo "  help ..................... Display this help information"
	@echo ""

# Default goal when `make` is run without arguments
.DEFAULT_GOAL = help