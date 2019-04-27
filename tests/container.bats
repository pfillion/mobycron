#!/usr/bin/env bats

NS=${NS:-pfillion}
IMAGE_NAME=${IMAGE_NAME:-mobycron}
VERSION=${VERSION:-latest}
CONTAINER_NAME="mobycron-${VERSION}"
DOER_CONTAINER_NAME="mobycron-doer-${VERSION}"

load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'

function teardown(){
    docker rm -f ${CONTAINER_NAME}
    docker rm -f ${DOER_CONTAINER_NAME} || true
}

@test "mobycron config file only" {
    docker run -d -e MOBYCRON_DOCKER_MODE=false -e MOBYCRON_PARSE_SECOND=true -e MOBYCRON_CONFIG_FILE=/configs/config.json -v $(pwd)/tests/configs:/configs --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	sleep 2
    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'job completed successfully.*Hello Joe'
}

@test "mobycron parse second not permitted" {
    docker run -d -e MOBYCRON_DOCKER_MODE=false -e MOBYCRON_PARSE_SECOND=false -e MOBYCRON_CONFIG_FILE=/configs/config.json -v $(pwd)/tests/configs:/configs --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	sleep 1
    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'failed to add jobs fron config file'
}

@test "mobycron start container" {
    docker create --name ${DOER_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='start' busybox echo 'Do job'
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	sleep 2

    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'container action completed successfully'
    run docker logs ${DOER_CONTAINER_NAME}
	assert_output --regexp 'Do job'
}

@test "mobycron restart container" {
    docker run -d --name ${DOER_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='restart' busybox echo 'Do job'
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	sleep 2
    
    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'container action completed successfully'
    run docker logs ${DOER_CONTAINER_NAME}
	assert_output --regexp 'Do job.*Do job'
}

@test "mobycron stop container" {
    docker run -d --name ${DOER_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='stop' busybox sh -c ' sleep 100 && echo ''Do job'''
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	sleep 15
    # TODO: set timeout to 1s
    
    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'container action completed successfully'
    run docker logs ${DOER_CONTAINER_NAME}
	assert_output ''
}

@test "mobycron exec container" {
    docker run -d --name ${DOER_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='exec' -l mobycron.command='echo ''Do job''' busybox sleep 100
    # docker run -d --name ${DOER_CONTAINER_NAME} -l mobycron.schedule='0/5 * * * * *' -l mobycron.action='exec' -l mobycron.command='backup.sh' -e 'MYSQL_ROOT_PASSWORD=rootpw' pfillion/mariadb
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	sleep 2
    # TODO: set timeout to 1s
    
    run docker logs ${CONTAINER_NAME}
    assert_output --regexp 'Do job'
	assert_output --regexp 'container action completed successfully'
}
