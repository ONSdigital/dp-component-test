#!/bin/bash -eux

go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1

pushd dp-component-test
  make lint
popd
