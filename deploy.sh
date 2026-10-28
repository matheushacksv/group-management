#!/bin/sh
# Deploy on the server: pull main, rebuild the images and update the stack.
# Called by GitHub Actions through a SSH key that can only run this script
# (forced command in /root/.ssh/authorized_keys), can also be run by hand.
set -eu

# everything inside a function: the shell reads it whole before running,
# so the git pull below can safely replace this very file
main() {
	cd /opt/groups

	# one deploy at a time, even with two pushes in a row
	exec 9>/tmp/groups-deploy.lock
	flock 9

	echo "==> git pull"
	git pull --ff-only

	echo "==> build"
	docker build -t group-api:latest .
	docker build -t group-web:latest ./web

	echo "==> deploy"
	docker stack deploy -c docker-compose.prod.yml groups

	# the tag is always "latest", --force makes swarm restart with the new image;
	# it waits the service to converge (api healthcheck) and fails the deploy if it doesn't
	docker service update --force --image group-api:latest groups_api
	docker service update --force --image group-web:latest groups_web

	echo "==> done: $(git log --oneline -1)"
}

main
