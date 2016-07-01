#!/bin/bash
go build -o QueueServer *.go
mv QueueServer $GOPATH/bin
