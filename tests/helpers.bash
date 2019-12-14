#!/bin/bash

# Retry a command $1 times until it succeeds. Wait $2 seconds between retries.
function retry() {
	local attempts=$1
	shift
	local delay=$1
	shift
	local i

	for ((i=0; i < attempts; i++)); do
		run "$@"
		if [[ "$status" -eq 0 ]] ; then
			return 0
		fi
		sleep $delay
	done

	echo "Command "$@" failed $attempts times. Output: $output"
	false
}

function init_swarm() {
	if [ $(docker system info | grep -c 'Swarm: active') -eq 0 ]; then
		 docker swarm init
	fi
}