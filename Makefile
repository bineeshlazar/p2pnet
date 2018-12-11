#Name of binary to be generated
BIN = mdns

#Name of docker image to be created
IMAGE = mdns

BUILDER=golang:latest

ENV=`pwd`/../build/env
PWDMOUNT = -v `pwd`:/usr/src/${BIN}
ENVMOUNT = -v ${ENV}:/go
WORKDIR = -w /usr/src/${BIN}

build: env
	@docker run --rm ${PWDMOUNT} ${ENVMOUNT} ${WORKDIR} ${BUILDER} /bin/sh build.sh ${BIN}

image: build
	@docker build -t ${IMAGE} --build-arg BIN=${BIN}  .
	@echo "Docker image ${IMAGE} created"
	@docker image ls ${IMAGE}

env:
	@mkdir -p ${ENV}
	@echo "environment created"

clean:
	@rm -f ${BIN}

distclean: clean
	@docker run --rm ${ENVMOUNT} ${BUILDER} /bin/sh  -c "ls | xargs rm -rf"
	@rm -rf ${ENV}
	- @docker rmi ${IMAGE}
