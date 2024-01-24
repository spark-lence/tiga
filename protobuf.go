package tiga

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

type ErrDetailsOption func(err error) *status.Status

// 基于message 名称反序列化
func ProtoMsgUnserializer(msgName string, data []byte) (interface{}, error) {
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgName))
	if err != nil {
		return nil, fmt.Errorf("Failed to find message type: %s,%s", err.Error(), msgName)
	}

	// 创建一个新的动态消息实例
	message := dynamicpb.NewMessage(msgType.Descriptor())
	// 反序列化 data 到动态消息实例
	if err := proto.Unmarshal(data, message); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data: %v", err)
	}
	return message.Interface(), nil
}
func RequestIdUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 生成一个唯一的 requestId
	requestId := uuid.New().String()

	// 将 requestId 添加到 Metadata 中
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	md = md.Copy()
	md.Set("x-request-id", requestId)
	ctx = metadata.NewIncomingContext(ctx, md)
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", requestId))
	// 调用实际的 RPC 方法
	return handler(ctx, req)
}

//	func WithErrDetailsUnwrap(unwrap func(err error)*status.Status) ErrDetailsOption {
//		return func(err error) *status.Status {
//			return unwrap(err)
//		}
//	}
func MakeErrWithDetails(rsp protoreflect.ProtoMessage, srcErr error) error {
	details, err := anypb.New(rsp)
	if err != nil {
		// Handle error
		st := status.New(codes.Internal, srcErr.Error())
		st, err = st.WithDetails(details)
		if err != nil {
			return srcErr
		}
		return st.Err()
	}
	st := status.New(codes.Internal, srcErr.Error())
	st, err = st.WithDetails(details)
	if err != nil {
		return srcErr
	}

	return st.Err()

}
func setFieldValueFromString(pbReflect protoreflect.Message, field protoreflect.FieldDescriptor, value string) error {
	// 根据字段类型进行转换
	switch field.Kind() {
	case protoreflect.BoolKind:
		val := value == "true"
		pbReflect.Set(field, protoreflect.ValueOfBool(val))

	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		val, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return fmt.Errorf("type mismatch for integer field")
		}
		pbReflect.Set(field, protoreflect.ValueOfInt64(val))

	case protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("type mismatch for unsigned integer field")
		}
		pbReflect.Set(field, protoreflect.ValueOfUint64(val))
	case protoreflect.StringKind:
		pbReflect.Set(field, protoreflect.ValueOfString(value))
	case protoreflect.BytesKind:

		pbReflect.Set(field, protoreflect.ValueOfBytes([]byte(value)))
	case protoreflect.MessageKind:
		val := dynamicpb.NewMessage(field.Message())
		err := proto.Unmarshal([]byte(value), val)
		if err != nil {
			return fmt.Errorf("type mismatch for message field")
		}
		pbReflect.Set(field, protoreflect.ValueOf(val))
	case protoreflect.DoubleKind:
		floatValue, err := strconv.ParseFloat(value, 64)

		if err != nil {
			return fmt.Errorf("type mismatch for double field")
		}
		pbReflect.Set(field, protoreflect.ValueOfFloat64(floatValue))
	case protoreflect.FloatKind:
		floatValue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return fmt.Errorf("type mismatch for float field")
		}
		pbReflect.Set(field, protoreflect.ValueOfFloat32(float32(floatValue)))
	case protoreflect.EnumKind:
		val, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return fmt.Errorf("type mismatch for enum field")
		}
		pbReflect.Set(field, protoreflect.ValueOfEnum(protoreflect.EnumNumber(val)))
	default:
		return fmt.Errorf("unsupported field type")
	}

	return nil
}
func MakeMapToProtobuf(data map[string]string, pb protoreflect.ProtoMessage) error {
	pbReflect := pb.ProtoReflect()
	fields := pbReflect.Descriptor().Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		jsonTag := field.JSONName()

		// 如果 map 中存在对应的键
		if value, ok := data[jsonTag]; ok {
			// 转换并设置值
			if err := setFieldValueFromString(pbReflect, field, value); err != nil {
				return fmt.Errorf("failed to set field '%s': %v", field.FullName(), err)
			}
		}
	}

	return nil
}
func MakeAPIResponse(code int32, data protoreflect.ProtoMessage, message string, API string) (protoreflect.ProtoMessage, error) {
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(API))
	if err != nil{
		return nil, fmt.Errorf("Failed to find message type: %s,%s", err.Error(), API)
	}
	rsp, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}
	dataType := ""
	if data != nil {
		dataType = string(data.ProtoReflect().Descriptor().FullName().Name())
	}
	// 创建一个新的动态消息实例
	msgPb := dynamicpb.NewMessage(msgType.Descriptor())
	msgPb.Set(msgPb.Descriptor().Fields().ByJSONName("code"), protoreflect.ValueOfFloat64(float64(code)))
	msgPb.Set(msgPb.Descriptor().Fields().ByJSONName("message"), protoreflect.ValueOfString(message))
	msgPb.Set(msgPb.Descriptor().Fields().ByJSONName("data"), protoreflect.ValueOfBytes(rsp))
	msgPb.Set(msgPb.Descriptor().Fields().ByJSONName("response_type"), protoreflect.ValueOfString(dataType))

	// msgPb.ProtoMessage()
	return msgPb, nil
}
func MakeResponse(code int32, data protoreflect.ProtoMessage, srcErr error, message string, API string) (protoreflect.ProtoMessage, error) {
	rsp, err := MakeAPIResponse(code, data, message, API)
	if err != nil {
		return rsp, fmt.Errorf("构建响应失败,%w", err)
	}
	if srcErr != nil {
		srcErr = MakeErrWithDetails(rsp, srcErr)

	}
	return rsp, srcErr
}

func GetFieldValueFromPb(message protoreflect.ProtoMessage, filedName string) (interface{}, error) {
	// 使用Protocol Buffer反射获取消息
	value := message.ProtoReflect()
	fieldDescriptor := value.Descriptor().Fields().ByName(protoreflect.Name(filedName))
	if fieldDescriptor == nil {
		return nil, fmt.Errorf("invalid field name: %s", filedName)
	}
	fieldValue := value.Get(fieldDescriptor)
	return fieldValue.Interface(), nil

}
func GetValueOfTaggedField(input interface{}, tagName, tagValue string) (interface{}, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 确保我们正在处理一个结构体
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct or a pointer to a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)
		if tagVal, ok := structField.Tag.Lookup(tagName); ok && tagVal == tagValue {
			// 找到匹配的tag，返回其值
			return field.Interface(), nil
		}
	}

	return nil, fmt.Errorf("no field found with tag %s:%s", tagName, tagValue)
}
