include common.mk

ECR_LOGIN=bash get-ecr-token.sh ${ECR_HOST}

.PHONY: all

.PHONY: default
default: all

.PHONY: clean
clean:
	-${DOCKER} volume rm ${GO_PKG_VOLUME}

GO_TEST_TARGETS=toplevel_go-test

# test is a superset of go-test, and runs some additional tests in containers.
.PHONY: test
test: $(GO_TEST_TARGETS)
	@

# IMPORTANT: go-test does not run any docker containers! It is called by
# bitbucket pipelines and `citest` which mimics pipelines.
.PHONY: go-test
all: go-test
go-test: $(GO_TEST_TARGETS)
	@

.PHONY: toplevel_go-test
toplevel_go-test:
	go test ./...

.PHONY: print-go-env
print-go-env:
	env
	go env

# IMPORTANT: run `make citest` before creating a PR! This mimics the pipelines
# CI run on bitbucket.
.PHONY: citest
citest:
	${DOCKER_RUN} \
		$(shell pinata-ssh-mount) \
		${GO_PKG_MOUNT} \
		-v ${CURDIR}":"/go/src/${PROJECT_PATH} \
		-e "CGO_LDFLAGS_ALLOW=-Wl,--no-as-needed" \
		-w /go/src/${PROJECT_PATH} \
		--entrypoint bash \
		${BUILD_IMAGE_GO} -c 'set -o xtrace; chmod 0777 ${GO_PKG_PATH} && env | egrep "^(PATH|docker_tarball|SSH_AUTH_SOCK|GOPATH|GOCACHE|GOPRIVATE|CGO_LDFLAGS_ALLOW|GOLANG_VERSION)=" | sed -e "s/^/export /" >/tmp/env && cat /tmp/env && useradd --uid $(shell id -u) --gid $(shell id -g) --home-dir /tmp/somebody --create-home somebody && su -l -c "set -o xtrace; . /tmp/env && cd /go/src/${PROJECT_PATH} && make go-test" somebody'

.PHONY: docker-login
docker-login:
	${ECR_LOGIN}

.PHONY: lint
lint:
	${DOCKER_RUN} \
		$(shell pinata-ssh-mount) \
		${GO_PKG_MOUNT} \
		-v ${CURDIR}":"/go/src/${PROJECT_PATH} \
		-e "CGO_LDFLAGS_ALLOW=-Wl,--no-as-needed" \
		-w /go/src/${PROJECT_PATH} \
		--entrypoint bash \
		${BUILD_IMAGE_GO} -c 'set -o xtrace; chmod 0777 ${GO_PKG_PATH} && env | egrep "^(PATH|docker_tarball|SSH_AUTH_SOCK|GOPATH|GOCACHE|GOPRIVATE|CGO_LDFLAGS_ALLOW|GOLANG_VERSION)=" | sed -e "s/^/export /" >/tmp/env && cat /tmp/env && useradd --uid $(shell id -u) --gid $(shell id -g) --home-dir /tmp/somebody --create-home somebody && su -l -c "set -o xtrace; . /tmp/env && cd /go/src/${PROJECT_PATH} && make linttarget" somebody'

.PHONY: linttarget
linttarget:
	go vet ./...
	golint ./...
