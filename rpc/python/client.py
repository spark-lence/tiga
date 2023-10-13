#!/usr/bin/env python
# -*- encoding: utf-8 -*-
'''
@File    :   config.py
@Time    :   2023/10/13 09:06:17
@Desc    :   
'''
import pickle
import grpc

from .pb import config_pb2_grpc, config_pb2

class RemoteConfigure:
    def __init__(self, env="dev", address="localhost", port=50051):
        self.channel = grpc.insecure_channel(f'{address}:{port}')
        self.stub = config_pb2_grpc.ConfigStub(self.channel)
        self.env = env

    def get(self, key) -> bytes:
        return self.stub.GetConfig(config_pb2.ConfigRequest(key=key, env=self.env)).value

    def get_string(self, key) -> str:
        return self.get(key).decode("utf-8")

    def get_int(self, key) -> int:
        return int(self.get(key))

    def get_bool(self, key) -> bool:
        return bool(self.get(key))

    def get_object(self, key) -> list:
        return pickle.loads(self.get(key))
