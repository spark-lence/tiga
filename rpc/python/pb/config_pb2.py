# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: config.proto
"""Generated protocol buffer code."""
from google.protobuf from . import descriptor as _descriptor
from google.protobuf from . import descriptor_pool as _descriptor_pool
from google.protobuf from . import symbol_database as _symbol_database
from google.protobuf.internal from . import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0c\x63onfig.proto\x12\x02pb\")\n\rConfigRequest\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\x0b\n\x03\x65nv\x18\x02 \x01(\t\"\x1f\n\x0e\x43onfigResponse\x12\r\n\x05value\x18\x01 \x01(\x0c\x32>\n\x06\x43onfig\x12\x34\n\tGetConfig\x12\x11.pb.ConfigRequest\x1a\x12.pb.ConfigResponse\"\x00\x42?\n\x11\x63om.sparklence.pbB\x06\x43onfigP\x01Z github.com/spark-lence/common/pbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'config_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\021com.sparklence.pbB\006ConfigP\001Z github.com/spark-lence/common/pb'
  _globals['_CONFIGREQUEST']._serialized_start=20
  _globals['_CONFIGREQUEST']._serialized_end=61
  _globals['_CONFIGRESPONSE']._serialized_start=63
  _globals['_CONFIGRESPONSE']._serialized_end=94
  _globals['_CONFIG']._serialized_start=96
  _globals['_CONFIG']._serialized_end=158
# @@protoc_insertion_point(module_scope)
