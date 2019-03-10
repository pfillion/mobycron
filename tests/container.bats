#!/usr/bin/env bats

NS=${NS:-pfillion}
IMAGE_NAME=${IMAGE_NAME:-mobycron}
VERSION=${VERSION:-latest}
CONTAINER_NAME="mobycron-${VERSION}"

load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'

function teardown(){
    docker rm -f ${CONTAINER_NAME}
}

@test "mobycron" {
    # Given a json config, when starting the container, then the job is scheduled and executed
    docker run -d -v $(pwd)/tests/configs:/configs --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	sleep 3
    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'job completed successfully.*Hello Joe'
}