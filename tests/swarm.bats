#!/usr/bin/env bats

NS=${NS:-pfillion}
IMAGE_NAME=${IMAGE_NAME:-mobycron}
VERSION=${VERSION:-latest}
# Swarm don't use local image, need a specific digest
IMAGE_DIGEST=$(docker image inspect ${NS}/${IMAGE_NAME}:${VERSION} --format={{.ID}})
SERVICE_NAME="mobycron-${VERSION}"
DOER1_SERVICE="mobycron-doer1-${VERSION}"

load helpers
load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'

function setup(){
   init_swarm
}

function teardown(){
    docker service rm ${SERVICE_NAME} || true
    docker service rm ${DOER1_SERVICE} || true
}

function service_action_completed_successfully() {
	[ $(docker service logs $1 | grep -c 'service action completed successfully') -ge $2 ]
}

function remove_service_job_from_cron() {
	[ $(docker service logs $1 | grep -c 'remove service job from cron') -ge $2 ]
}

@test "swarm scan - update all replicas from service" {
    # Arrange
    docker service create --name ${DOER1_SERVICE} --replicas=2 --restart-condition=none --label=mobycron.schedule='* * * * * *' --label=mobycron.action='update' busybox sh -c 'echo ''job1'' && sleep 6'

	# Act
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=swarm -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}

	# Assert
	retry 10 1 service_action_completed_successfully ${SERVICE_NAME} 4
    run docker service logs ${DOER1_SERVICE}
    assert_output --regexp ${DOER1_SERVICE}'\.1.*job1'
    assert_output --regexp ${DOER1_SERVICE}'\.2.*job1'
}

@test "swarm listen - add service" {
	# Arrange
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=swarm -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}
	
    # Act
    docker service create --name ${DOER1_SERVICE} --replicas=1 --restart-condition=none --label=mobycron.schedule='*/10 * * * * *' --label=mobycron.action='update' busybox sh -c 'echo ''job1'' && sleep 6'

	# Assert
    retry 15 1 service_action_completed_successfully ${SERVICE_NAME} 2
    run docker service logs ${DOER1_SERVICE}
    assert_output --regexp 'job1.*job1'
}

@test "swarm listen - update service" {
	# Arrange
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=swarm -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}
    docker service create --name ${DOER1_SERVICE} --replicas=1 --restart-condition=none --label=mobycron.schedule='*/10 * * * * *' --label=mobycron.action='update' busybox sh -c 'echo ''job1'' && sleep 6'

    # Act
    docker service update --args "sh -c 'echo ''job2'' && sleep 6'" ${DOER1_SERVICE}

	# Assert
    retry 10 1 remove_service_job_from_cron ${SERVICE_NAME} 1
    run docker service logs ${DOER1_SERVICE}
    assert_output --regexp 'job1.*job2'
}

@test "swarm listen - remove service" {
	# Arrange
    docker service create --name ${SERVICE_NAME} -e MOBYCRON_DOCKER_MODE=swarm -e MOBYCRON_PARSE_SECOND=true --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock ${IMAGE_DIGEST}
    docker service create --name ${DOER1_SERVICE} --replicas=1 --restart-condition=none --label=mobycron.schedule='* */5 * * * *' --label=mobycron.action='update' busybox sleep 6
    
	# Act
    docker service rm ${DOER1_SERVICE}
  
	# Assert
    retry 10 1 remove_service_job_from_cron ${SERVICE_NAME} 1
}