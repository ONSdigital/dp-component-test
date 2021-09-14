#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-component-test
  make test-component
popd