# Targets to build (from the cmd/ folder)
TARGETS := contactcache

#Go Build Args
GOBUILD := CGO_ENABLED=0 go build
BUILD_DIR := ./build
BUILD_FLAGS := -ldflags="-s -w" -trimpath

# Docker Args
DOCKER := docker
DOCKERBUILD := $(DOCKER) build
DOCKERPUSH := $(DOCKER) push
DOCKERREPO := tcfw
DOCKEROPTS := --compress=true

########
# Build pipelines
########

include ./scripts/make_vars.mk

MAKEFLAGS += '-j 2'

default: all

.PHONY: test
test:
	go test -v github.com/tcfw/test-contactcache/pkg/contactcache

.PHONY: clean
clean:
	@echo -e '$(BROOM_EMJ) Cleaning up binaries...'
	@rm -f $(BUILD_DIR)/*

clean-docker:
	@echo -e '$(BROOM_EMJ) Cleaning up docker containers...'
	@docker image remove $(TARGETS:%=$(DOCKERREPO)/%:latest)

.PHONY: all
all: $(TARGETS)
	@echo
	@echo -e '$(STAR_EMJ) Builds finished!'

.PHONY: build
build: all

.PHONY: docker
docker: 
	@echo -e '$(PACKAGE_EMJ) Making all containers...'
	@$(MAKE) -s $(TARGETS:=-docker)
	@echo
	@echo -e '$(STAR_EMJ) Docker builds finished!'

$(TARGETS:=-docker):
ifneq ("$(which buildah)",)
	$(eval DOCKERBUILD := buildah build-using-dockerfile )
	$(info Using buildah as build tool)
endif
	@echo -e '$(BUILD_EMJ) Building Container $(PACKAGE_EMJ) $@'
	@sed -e "s/_SERVICE_/$(@:-docker=)/g" ./deployments/Dockerfile | \
		$(DOCKERBUILD) $(DOCKEROPTS) -t $(DOCKERREPO)/$(@:-docker=) -f - .

.PHONY: publish
publish:
	@echo -e 'Publishing containers $(PACKAGE_EMJ)'
	@$(MAKE) -s $(TARGETS:=-docker-publish)

$(TARGETS:=-docker-publish):
ifneq ("$(which buildah)",)
	$(eval DOCKERPUSH := buildah push )
	$(info Using buildah as build tool)
endif
	@echo 'Publishing $(@:-docker-publish=)'
	@$(DOCKERPUSH) $(DOCKERREPO)/$(@:-docker-publish=)


$(WEBTARGETS:=-publish):
ifneq ("$(which buildah)",)
	$(eval DOCKERPUSH := buildah push )
	$(info Using buildah as build tool)
endif
	@echo 'Publishing $(@:-docker-publish=)'
	@$(DOCKERPUSH) $(DOCKERREPO)/$(@:-publish=)

$(TARGETS): 
	@echo -e '$(BUILD_EMJ) Building $@'
	@$(GOBUILD) $(BUILD_FLAGS) -o ./build/$@ ./cmd/$@
	@strip --strip-unneeded ./build/$@
	@echo -e '$(TICK_EMJ) Built $@'