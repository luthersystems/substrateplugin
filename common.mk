PROJECT=substrateplugin
PROJECT_PATH=github.com/luthersystems/${PROJECT}

VERSION=0.1.0-SNAPSHOT

PACKAGE=github.com/luthersystems/${PROJECT}
AWS_REGION=eu-west-2
ECR_HOST=967058059066.dkr.ecr.${AWS_REGION}.amazonaws.com

BUILDENV_TAG=0.0.28
BUILD_IMAGE_GO=${ECR_HOST}/luthersystems/build-go:${BUILDENV_TAG}
BUILD_IMAGE_GODYNAMIC=${ECR_HOST}/luthersystems/build-godynamic:${BUILDENV_TAG}
BUILD_IMAGE_GOEXTRA=${ECR_HOST}/luthersystems/build-goextra:${BUILDENV_TAG}
BUILD_IMAGE_API=${ECR_HOST}/luthersystems/build-api:${BUILDENV_TAG}

GO_PKG_VOLUME=${PROJECT}-build-gopath-pkg
GO_PKG_PATH=/go/pkg
GO_PKG_MOUNT=$(if $(CI),-v $(PWD)/build/pkg:${GO_PKG_PATH},--mount='type=volume,source=${GO_PKG_VOLUME},destination=${GO_PKG_PATH}')

CP=cp
RM=rm
DOCKER=docker
DOCKER_RUN_OPTS=--rm
DOCKER_RUN=${DOCKER} run ${DOCKER_RUN_OPTS}
CHOWN=$(if $(CIRCLECI),sudo chown,chown)
CHOWN_USR=$(LOGNAME)
CHOWN_USR?=$(USER)
CHOWN_GRP=$(if $(or $(IS_WINDOWS),$(CIRCLECI)),,admin)
DOCKER_USER="$(shell id -u ${CHOWN_USR}):$(shell id -g ${CHOWN_USR})"
DOMAKE=$(MAKE) -C $1 $2 # NOTE: this is not used for now as it does not work with -j for some versions of Make
MKDIR_P=mkdir -p
TOUCH=touch
GZIP=gzip
GUNZIP=gunzip
