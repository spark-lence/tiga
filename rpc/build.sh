#!/usr/bin/bash
protoc -I=proto \
    --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto
python3.9 -m grpc_tools.protoc -I=proto --python_out=python/pb --grpc_python_out=python/pb  proto/*.proto
sed -i 's/import \(.*\) as/from . import \1 as/g' ./python/pb/*.py*