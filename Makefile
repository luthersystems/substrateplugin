include common.mk

ECR_LOGIN=bash get-ecr-token.sh ${ECR_HOST}

.PHONY: all

.PHONY: default
default: all

.PHONY: clean
clean:
	# docker volume rm will fail if the volume doesn't exist
	-${DOCKER} volume rm ${GO_PKG_VOLUME}
	# make sure it's really gone
	sh -c '! ${DOCKER} volume inspect ${GO_PKG_VOLUME}'

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
		-v ${CURDIR}/guest.sh":"/tmp/guest.sh \
		-e "CGO_LDFLAGS_ALLOW=-Wl,--no-as-needed" \
		-w /go/src/${PROJECT_PATH} \
		--entrypoint /tmp/guest.sh \
		${BUILD_IMAGE_GODYNAMIC} ${GO_PKG_PATH} $(shell id -u) /go/src/${PROJECT_PATH} "make go-test"

.PHONY: docker-login
docker-login:
	${ECR_LOGIN}

.PHONY: lint
lint:
	${DOCKER_RUN} \
		$(shell pinata-ssh-mount) \
		${GO_PKG_MOUNT} \
		-v ${CURDIR}":"/go/src/${PROJECT_PATH} \
		-v ${CURDIR}/guest.sh":"/tmp/guest.sh \
		-e "CGO_LDFLAGS_ALLOW=-Wl,--no-as-needed" \
		-w /go/src/${PROJECT_PATH} \
		--entrypoint /tmp/guest.sh \
		${BUILD_IMAGE_GODYNAMIC} ${GO_PKG_PATH} $(shell id -u) /go/src/${PROJECT_PATH} "make linttarget"

.PHONY: linttarget
linttarget:
	go vet ./...
	golint ./...
