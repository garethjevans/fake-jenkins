TOOLS_DIR := tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
CERT_DIR := cert
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
GINKGO := $(TOOLS_BIN_DIR)/ginkgo

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: linux
linux: ## Build Linux executable.
	GOARCH=amd64 GOOS=linux go build -o ./build/linux/fake-jenkins main.go

.PHONY: darwin
darwin: ## Build Darwin executable.
	GOARCH=amd64 GOOS=darwin go build -o ./build/darwin/fake-jenkins main.go

.PHONY: clean
clean: ## Clean build detritus.
	rm -rf build
	rm -rf $(TOOLS_BIN_DIR)
	rm -rf $(CERT_DIR)

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint the plugin.
	$(GOLANGCI_LINT) run -v --timeout 5m

$(CERT_DIR) $(TOOLS_BIN_DIR):
	-mkdir -p $@

$(GOLANGCI_LINT): $(TOOLS_BIN_DIR)
	go build -tags=tools -o $@ github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: test
test: ## Run the golang test suite.
	go test ./...

$(GINKGO): $(TOOLS_BIN_DIR)
	go build -tags=tools -o $@ github.com/onsi/ginkgo/v2/ginkgo

.PHONY: cert
ca_cert: $(CERT_DIR) ## Generate self-signed CA certificate
	openssl req -nodes -new -x509 -keyout $(CERT_DIR)/key.pem -out $(CERT_DIR)/ca_cert.pem -days 365 -subj '/CN=127.0.0.1'

.PHONY: run-with-ca
run_with_ca: ca_cert darwin ## Run fake jenkins server with CA certificate
	./build/darwin/fake-jenkins -local -key $(CERT_DIR)/key.pem -cert $(CERT_DIR)/ca_cert.pem
