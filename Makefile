### -----------------------
# --- Make variables
### -----------------------

# only evaluated if required by a recipe
# http://make.mad-scientist.net/deferred-simple-variable-expansion/

# go module name (as in go.mod)
GO_MODULE_NAME = $(eval GO_MODULE_NAME := $$(shell \
	(mkdir -p tmp 2> /dev/null && cat .modulename 2> /dev/null) \
	|| (gsdev modulename 2> /dev/null | tee .modulename) || echo "unknown" \
))$(GO_MODULE_NAME)

MODULE_NAME = $(eval MODULE_NAME := $(shell echo $(GO_MODULE_NAME) | awk -F'/' '{print $$3}'))$(MODULE_NAME)

# https://medium.com/the-go-journey/adding-version-information-to-go-binaries-e1b79878f6f2
ARG_COMMIT = $(eval ARG_COMMIT := $$(shell \
	(git rev-list -1 HEAD 2> /dev/null) \
	|| (echo "unknown") \
))$(ARG_COMMIT)

ARG_BUILD_DATE = $(eval ARG_BUILD_DATE := $$(shell \
	(date -Is 2> /dev/null || date 2> /dev/null || echo "unknown") \
))$(ARG_BUILD_DATE)

# https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
LDFLAGS = $(eval LDFLAGS := "\
-X '$(GO_MODULE_NAME)/internal/config.ModuleName=$(MODULE_NAME)'\
-X '$(GO_MODULE_NAME)/internal/config.Commit=$(ARG_COMMIT)'\
-X '$(GO_MODULE_NAME)/internal/config.BuildDate=$(ARG_BUILD_DATE)'\
")$(LDFLAGS)

### -----------------------
# --- Building
### -----------------------

CGO_ENABLED ?= 0

go-build: ##- (opt) Runs go build.
	CGO_ENABLED=$(CGO_ENABLED) go build -mod=vendor -ldflags $(LDFLAGS) -o bin/${MODULE_NAME}

proto:
	rm -rf pkg/pb/*.go
	protoc --go_out=plugins=grpc+retag:. pkg/proto/*.proto pkg/proto/ag_algo/*.proto

tidy: ##- (opt) Tidy our go.sum file.
	go mod tidy

vendor:
	go mod vendor -o vendor-tmp
	cp -rvf vendor-tmp/* vendor/
	rm -rf vendor-tmp

### -----------------------
# --- Helpers
### -----------------------

clean: ##- Cleans ./tmp and ./api/tmp folder.
	@echo "make clean"
	@rm -rf tmp 2> /dev/null
	@rm -rf api/tmp 2> /dev/null

get-module-name: ##- Prints current go module-name (pipeable).
	@echo "${GO_MODULE_NAME}"

info-module-name: ##- (opt) Prints current go module-name.
	@echo "go module-name: '${GO_MODULE_NAME}'"

set-module-name: ##- Wizard to set a new go module-name.
	@rm -rf .modulename
	@echo "Enter new go module-name:" \
		&& read new_module_name \
		&& echo "new go module-name: '$${new_module_name}'" \
		&& echo -n "Are you sure? [y/N]" \
		&& read ans && [ $${ans:-N} = y ] \
		&& echo -n "Please wait..." \
		&& find . -not -path '*/\.*' -not -path './Makefile' -type f -exec sed -i "s|${GO_MODULE_NAME}|172.21.5.249/airtrans/$${new_module_name}|g" {} \; \
		&& echo "172.21.5.249/airtrans/$${new_module_name}" >> .modulename \
		&& echo "new go module-name: '$${new_module_name}'!"

get-go-ldflags: ##- (opt) Prints used -ldflags as evaluated in Makefile used in make go-build
	@echo $(LDFLAGS)

tools: ##- (opt) Install packages as specified in tools.go.
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -P $$(nproc) -L 1 -tI % go install %	

docker:
	docker build \
		--no-cache \
		--build-arg SVC=$(MODULE_NAME) \
		--tag=harbor.vht.vn/c4i/$(MODULE_NAME):dev \
		-f Dockerfile . \

docker-push:
	make docker 
	docker push harbor.vht.vn/c4i/$(MODULE_NAME):dev

release:
	make docker
	$(eval version = $(shell git describe --tags))
	docker tag harbor.vht.vn/c4i/$(MODULE_NAME) harbor.vht.vn/c4i/$(MODULE_NAME):$(version)
	docker push harbor.vht.vn/c4i/$(MODULE_NAME):$(version)
	docker rmi harbor.vht.vn/c4i/$(MODULE_NAME)
	docker rmi harbor.vht.vn/c4i/$(MODULE_NAME):$(version)	

swag:
	swag init --parseDependency  --parseInternal --parseDepth 100  -g main.go
	swag fmt	

server:
	@make swag
	@make go-build && ./bin/$(MODULE_NAME) start

.PHONY: tools vendor proto
