#!/usr/bin/env bash

function docker_image_name() {
	echo browsh/browsh:v"$BROWSH_VERSION"
}

function docker_build() {
	local og_xpi && og_xpi=$(versioned_xpi_file)
	[ ! -f "$og_xpi" ] && _panic "Can't find latest webextension build: $og_xpi"
	[ ! -f "$XPI_PATH" ] && _panic "Can't find bundleable browsh.xpi: $XPI_PATH"
	if [ "$(_md5 "$og_xpi")" != "$(_md5 "$XPI_PATH")" ]; then
		_panic "XPI file's MD5 does not match original XPI file's MD5"
	fi
	docker build -t "$(docker_image_name)" .
}

function is_docker_logged_in() {
	docker system info | grep -E 'Username|Registry'
}

function docker_login() {
	docker login docker.io \
		-u tombh \
		-p "$DOCKER_ACCESS_TOKEN"
}

function docker_release() {
	! is_docker_logged_in && try_docker_login
	docker_build
	docker push "$(docker_image_name)"
}
