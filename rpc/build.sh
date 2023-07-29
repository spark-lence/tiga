#!/usr/bin/bash
protoc -I=proto \
    --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto
python3.11 -m grpc_tools.protoc -I=proto --python_out=pb --grpc_python_out=pb --pyi_out=pb proto/*.proto