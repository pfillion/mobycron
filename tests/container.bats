#!/usr/bin/env bats

NS=${NS:-pfillion}
IMAGE_NAME=${IMAGE_NAME:-mobycron}
VERSION=${VERSION:-latest}
CONTAINER_NAME="mobycron-${VERSION}"
DOER1_CONTAINER_NAME="mobycron-doer1-${VERSION}"
DOER2_CONTAINER_NAME="mobycron-doer2-${VERSION}"

load helpers
load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'

function teardown(){
    docker rm -f ${CONTAINER_NAME}
    docker rm -f ${DOER1_CONTAINER_NAME} || true
    docker rm -f ${DOER2_CONTAINER_NAME} || true
}

function job_completed_successfully() {
	[ $(docker logs $1 | grep -c 'job completed successfully') -ge $2 ]
}

function container_action_completed_successfully() {
	[ $(docker logs $1 | grep -c 'container action completed successfully') -ge $2 ]
}

function exit_fatal() {
	[ $(docker logs $1 | grep -c 'fatal') -ge $2 ]
}

@test "mobycron config file only with multiple jobs" {
    # Act
    # Prevent Bug: Only last job is executed. It is due to use of range and invalid use of pointer in Cron.AddJobs
    docker run -d -e MOBYCRON_DOCKER_MODE=false -e MOBYCRON_PARSE_SECOND=true -e MOBYCRON_CONFIG_FILE=/configs/config.json -v $(pwd)/tests/configs:/configs --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	
    # Assert
    retry 5 1 job_completed_successfully ${CONTAINER_NAME} 3
    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'job completed successfully.*Hello Joe'
    assert_output --regexp 'job completed successfully.*Hello Bob'
}

@test "mobycron parse second not permitted" {
    # Act
    docker run -d -e MOBYCRON_DOCKER_MODE=false -e MOBYCRON_PARSE_SECOND=false -e MOBYCRON_CONFIG_FILE=/configs/config.json -v $(pwd)/tests/configs:/configs --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	
    # Assert
    retry 5 1 exit_fatal ${CONTAINER_NAME} 1
    run docker logs ${CONTAINER_NAME}
	assert_output --regexp 'failed to add jobs fron config file'
}

@test "mobycron start container" {
    # Arrange
    docker create --name ${DOER1_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='start' busybox echo 'Do job'
    
    # Act
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	
    # Assert
    retry 5 1 container_action_completed_successfully ${CONTAINER_NAME} 1
    run docker logs ${DOER1_CONTAINER_NAME}
	assert_output --regexp 'Do job'
}

@test "mobycron restart container" {
    # Arrange
    docker run -d --name ${DOER1_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='restart' busybox echo 'Do job'

    # Act
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	
    # Assert
    retry 5 1 container_action_completed_successfully ${CONTAINER_NAME} 1
    run docker logs ${DOER1_CONTAINER_NAME}
	assert_output --regexp 'Do job.*Do job'
}

@test "mobycron stop container" {
    # Arrange
    docker run -d --name ${DOER1_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.timeout='1' -l mobycron.action='stop' busybox sh -c ' sleep 100 && echo ''Do job'''
    
    # Act
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	
    # Assert
    retry 5 1 container_action_completed_successfully ${CONTAINER_NAME} 1
    run docker logs ${DOER1_CONTAINER_NAME}
	assert_output ''
}

@test "mobycron exec container" {
    # Arrange
    docker run -d --name ${DOER1_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='exec' -l mobycron.command='echo ''Do job''' busybox sleep 100
   
    # Act
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	
    # Assert
    retry 5 1 container_action_completed_successfully ${CONTAINER_NAME} 1
    run docker logs ${CONTAINER_NAME}
    assert_output --regexp 'Do job'
}

@test "mobycron multiple containers" {
    # Arrange
    docker run -d --name ${DOER1_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='exec' -l mobycron.command='echo ''Do job1''' busybox sleep 100
    docker run -d --name ${DOER2_CONTAINER_NAME} -l mobycron.schedule='* * * * * *' -l mobycron.action='exec' -l mobycron.command='echo ''Do job2''' busybox sleep 100
    
    # Act
    docker run -d -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true -v /var/run/docker.sock:/var/run/docker.sock --name ${CONTAINER_NAME} ${NS}/${IMAGE_NAME}:${VERSION}
	
    # Assert
    retry 5 1 container_action_completed_successfully ${CONTAINER_NAME} 1
    run docker logs ${CONTAINER_NAME}
    assert_output --regexp 'Do job1'
    assert_output --regexp 'Do job2'
}