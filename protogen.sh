#!/bin/bash

# python3 -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. --pyi_out=. ./pb/svc/audio/audio.proto

protoc --go_out=../../.. --go-grpc_out=../../.. ./pb/**/**/*.proto

protoc -I . --grpc-gateway_out . \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true \
    ./pb/svc/fixer/fixer.proto