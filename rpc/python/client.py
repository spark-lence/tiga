#!/usr/bin/env python
# -*- encoding: utf-8 -*-
'''
@File    :   config.py
@Time    :   2023/10/13 09:06:17
@Desc    :   
'''
import pickle
import grpc
import cachetools
import msgpack
from .pb import config_pb2_grpc, config_pb2

class SingletonMeta(type):
    _instances = {}

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super(SingletonMeta, cls).__call__(*args, **kwargs)
        return cls._instances[cls]

class RemoteConfigure(metaclass=SingletonMeta):
    def __init__(self, env="dev", address="localhost", port=50051):
        self.channel = grpc.insecure_channel(f'{address}:{port}')
        self.stub = config_pb2_grpc.ConfigStub(self.channel)
        self.env = env

    @cachetools.cached(cache=cachetools.TTLCache(maxsize=1024, ttl=60))
    def get(self, key) -> bytes:
        return self.stub.GetConfig(config_pb2.ConfigRequest(key=key, env=self.env)).value

    def get_string(self, key) -> str:
        val = self.get(key)
        return msgpack.unpackb(val, raw=False)

    def get_int(self, key) -> int:
        return int(self.get(key))

    def get_bool(self, key) -> bool:
        return bool(self.get(key))

    def get_object(self, key) -> list:
        return pickle.loads(self.get(key))

    def set(self, key, value):
        return self.stub.SetConfig(config_pb2.ConfigRequest(key=key, value=value, env=self.env))
