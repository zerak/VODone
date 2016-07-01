#!/bin/bash
#find . -name "*.go" | xargs goimports -w
find . -name "*.go" | xargs gofmt -w
