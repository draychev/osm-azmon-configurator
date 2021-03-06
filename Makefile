#!make

TARGETS         := darwin/amd64 linux/amd64 windows/amd64
SHELL           := bash -auexo pipefail
BINNAME         ?= osm
DIST_DIRS       := find * -type d -exec
CTR_REGISTRY    ?= openservicemesh
CTR_TAG         ?= latest

GOPATH = $(shell go env GOPATH)
GOBIN  = $(GOPATH)/bin
GOX    = go run github.com/mitchellh/gox

include .env

VERSION ?= dev
BUILD_DATE=$$(date +%Y-%m-%d-%H:%M)
GIT_SHA=$$(git rev-parse HEAD)
BUILD_DATE_VAR := github.com/draychev/osm-azmon-configurator/pkg/version.BuildDate
BUILD_VERSION_VAR := github.com/draychev/osm-azmon-configurator/pkg/version.Version
BUILD_GITCOMMIT_VAR := github.com/draychev/osm-azmon-configurator/pkg/version.GitCommit

LDFLAGS ?= "-X $(BUILD_DATE_VAR)=$(BUILD_DATE) -X $(BUILD_VERSION_VAR)=$(VERSION) -X $(BUILD_GITCOMMIT_VAR)=$(GIT_SHA) -X main.chartTGZSource=$$(cat -) -s -w"

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/osm-azmon-configurator/osm-azmon-configurator -ldflags "-X $(BUILD_DATE_VAR)=$(BUILD_DATE) -X $(BUILD_VERSION_VAR)=$(VERSION) -X $(BUILD_GITCOMMIT_VAR)=$(GIT_SHA) -s -w" ./cmd/osm-azmon-configurator

DOCKER_DEMO_TARGETS = $(addprefix docker-build-, $(DEMO_TARGETS))
.PHONY: $(DOCKER_DEMO_TARGETS)
$(DOCKER_DEMO_TARGETS): NAME=$(@:docker-build-%=%)
$(DOCKER_DEMO_TARGETS):
	make build-$(NAME)
	docker build -t $(CTR_REGISTRY)/$(NAME):$(CTR_TAG) -f dockerfiles/Dockerfile.$(NAME) demo/bin/$(NAME)

docker-build-osm-azmon-configurator: build
	docker build -t $(CTR_REGISTRY)/osm-azmon-configurator:$(CTR_TAG) -f dockerfiles/Dockerfile.osm-azmon-configurator bin/osm-azmon-configurator

# docker-push-bookbuyer, etc
DOCKER_PUSH_TARGETS = $(addprefix docker-push-, $(DEMO_TARGETS) osm-azmon-configurator)
VERIFY_TAGS = 0
.PHONY: $(DOCKER_PUSH_TARGETS)
$(DOCKER_PUSH_TARGETS): NAME=$(@:docker-push-%=%)
$(DOCKER_PUSH_TARGETS):
	@if [ $(VERIFY_TAGS) != 1 ]; then make docker-build-$(NAME); docker push "$(CTR_REGISTRY)/$(NAME):$(CTR_TAG)" || { echo "Error pushing images to container registry $(CTR_REGISTRY)/$(NAME):$(CTR_TAG)"; exit 1; }; else bash scripts/publish-image.sh $(NAME); fi

.PHONY: docker-push
docker-push: $(DOCKER_PUSH_TARGETS)

.PHONY: shellcheck
shellcheck:
	shellcheck -x $(shell find . -name '*.sh')

check-env:
ifndef CTR_REGISTRY
	$(error CTR_REGISTRY environment variable is not defined; see the .env.example file for more information; then source .env)
endif
ifndef CTR_TAG
	$(error CTR_TAG environment variable is not defined; see the .env.example file for more information; then source .env)
endif
