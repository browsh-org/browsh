#!/bin/bash
set -e

go test browsh.go browsh_test.go webextension.go -args -use-existing-ff -debug
