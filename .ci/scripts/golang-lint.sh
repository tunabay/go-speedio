#!/bin/bash

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
  | sh -s v1.25.1

./bin/golangci-lint run \
  --issues-exit-code=0 \
  --timeout=5m0s \
  --tests=false \
  --no-config \
  --max-issues-per-linter=4096 \
  --max-same-issues=1024 \
  --disable-all \
  --enable=deadcode \
  --enable=errcheck \
  --enable=gosimple \
  --enable=govet \
  --enable=ineffassign \
  --enable=staticcheck \
  --enable=structcheck \
  --enable=typecheck \
  --enable=unused \
  --enable=varcheck \
  --enable=bodyclose \
  --enable=depguard \
  --enable=dogsled \
  --enable=funlen \
  --enable=gocognit \
  --enable=goconst \
  --enable=gocritic \
  --enable=gocyclo \
  --enable=gofmt \
  --enable=goimports \
  --enable=golint \
  --enable=goprintffuncname \
  --enable=gosec \
  --enable=interfacer \
  --enable=misspell \
  --enable=nakedret \
  --enable=prealloc \
  --enable=rowserrcheck \
  --enable=scopelint \
  --enable=stylecheck \
  --enable=unconvert \
  --enable=unparam \
  --enable=godot \
  --enable=gomodguard \
  --enable=asciicheck \
  --enable=nestif \
  --enable=testpackage \
  --fix=false

# ./bin/golangci-lint linters
