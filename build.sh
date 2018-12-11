#!/bin/bash

echo "getting dependancies"
go get -v -d ./...

echo "Building"
CGO_ENABLED=0 GOOS=linux go build -v -o ${1}
