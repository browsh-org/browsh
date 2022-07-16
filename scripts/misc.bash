#!/bin/env bash

function golang_lint_check() {
	pushd "$PROJECT_ROOT"/interfacer || _panic
	diff -u <(echo -n) <(gofmt -d ./)
	popd || _panic
}

function golang_lint_fix() {
	gofmt -w ./interfacer
}
