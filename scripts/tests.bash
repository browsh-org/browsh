# For the webextension: in `webext/` folder, `npm test`
# For CLI unit tests: in `/interfacer` run `go test src/browsh/*.go`
# For CLI E2E tests: in `/interfacer` run `go test test/tty/*.go`
# For HTTP Service tests: in `/interfacer` run `go test test/http-server/*.go`

function test_all {
	test_webextension
	interfacer_test_setup
	test_interfacer_units
	test_http_server
	test_tty
}

function test_webextension {
	pushd $PROJECT_ROOT/webext
	npm test
}

function interfacer_test_setup {
	pushd $PROJECT_ROOT/webext
	touch "$PROJECT_ROOT/interfacer/src/browsh/browsh.xpi"
	npm run build:dev
}

function test_interfacer_units {
	pushd $PROJECT_ROOT/interfacer
	go test -v $(find src/browsh -name '*.go' | grep -v windows)
}

function test_tty {
	pushd $PROJECT_ROOT/interfacer
	go test test/tty/*.go -v -ginkgo.slowSpecThreshold=30 -ginkgo.flakeAttempts=3
}

function test_http_server {
	pushd $PROJECT_ROOT/interfacer
	go test test/http-server/*.go -v -ginkgo.slowSpecThreshold=30 -ginkgo.flakeAttempts=3
}
