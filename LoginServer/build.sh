#!/bin/bash
go build -o LoginServer *.go
mv LoginServer $GOPATH/bin
