#!/usr/bin/env bats

NS=${NS:-pfillion}
IMAGE_NAME=${IMAGE_NAME:-mobycron}
VERSION=${VERSION:-latest}
# Swarm don't use local image, need a specific digest
IMAGE_DIGEST=$(docker image inspect ${NS}/${IMAGE_NAME}:${VERSION} --format={{.ID}})
SERVICE_NAME="mobycron-${VERSION}"
DOER1_SERVICE="mobycron-doer1-${VERSION}"
DOER2_SERVICE="mobycron-doer2-${VERSION}"

load helpers
load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'

function setup(){
   init_swarm
}

function teardown(){
    docker service rm ${SERVICE_NAME} || true
    docker service rm ${DOER1_SERVICE} || true
    docker service rm ${DOER2_SERVICE} || true
}

function service_is_ready() {
	[ $(docker service inspect $1 | grep -c "$2") -eq 1 ] &&
	[ $(docker service ps $1 -f "desired-state=$3" --format {{.Name}} | grep -c "$1") -eq $4 ]
}

function job_completed_successfully() {
	[ $(docker service logs $1 | grep -c 'job completed successfully') -ge $2 ]
}

function container_action_completed_successfully() {
	[ $(docker service logs $1 | grep -c 'container action completed successfully') -ge $2 ]
}

function remove_container_job_from_cron() {
	[ $(docker service logs $1 | grep -c 'remove container job from cron') -ge $2 ]
}

@test "swarm scan - start all replicas from service" {
    # Arrange
    docker service create --name ${DOER1_SERVICE} --replicas=2 --restart-condition=none --container-label=mobycron.schedule='* * * * * *' --container-label=mobycron.action='start' busybox sleep 6

	# Act
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}

	# Assert
	retry 10 1 container_action_completed_successfully ${SERVICE_NAME} 4
    run docker service logs ${SERVICE_NAME}
    assert_output --regexp 'add container job to cron.*add container job to cron'
    assert_output --regexp ${DOER1_SERVICE}'\.1.*container action completed successfully'
    assert_output --regexp ${DOER1_SERVICE}'\.2.*container action completed successfully'
}

@test "swarm scan - start container only from active service task" {
	# Arrange
    docker service create --name ${DOER1_SERVICE} --replicas=1 --restart-condition=none --container-label=mobycron.schedule='*/2 * * * * *' --container-label=mobycron.action='start' busybox sleep 6
    docker service update --container-label-add=mobycron.schedule='* * * * * *' ${DOER1_SERVICE}

	# Act
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}
  
	# Assert
	retry 10 1 container_action_completed_successfully ${SERVICE_NAME} 4
    
    run docker service logs ${SERVICE_NAME}
    assert_output --regexp 'add container job to cron.*remove container job from cron'
    assert_output --regexp 'container action completed successfully.*\* \* \* \* \* \*'
    refute_output --regexp 'container action completed successfully.*\*\/2 \* \* \* \* \*'
}

@test "swarm listen - add service" {
	# Arrange
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}
	
    # Act
    docker service create --name ${DOER1_SERVICE} --replicas=1 --restart-condition=none --container-label=mobycron.schedule='* * * * * *' --container-label=mobycron.action='start' busybox sleep 6

	# Assert
    retry 10 1 container_action_completed_successfully ${SERVICE_NAME} 1
}

@test "swarm listen - update service" {
	# Arrange
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}
    docker service create --name ${DOER1_SERVICE} --replicas=1 --restart-condition=none --container-label=mobycron.schedule='*/10 * * * * *' --container-label=mobycron.action='start' busybox sleep 6

    # Act
    docker service update --args "sleep 7" ${DOER1_SERVICE}

	# Assert
    retry 10 1 remove_container_job_from_cron ${SERVICE_NAME} 1
}

@test "swarm listen - remove service" {
	# Arrange
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=true -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}
    docker service create --name ${DOER1_SERVICE} --replicas=1 --restart-condition=none --container-label=mobycron.schedule='* * * * * *' --container-label=mobycron.action='start' busybox sleep 6
    
	# Act
    docker service rm ${DOER1_SERVICE}
  
	# Assert
    retry 10 1 remove_container_job_from_cron ${SERVICE_NAME} 1
}