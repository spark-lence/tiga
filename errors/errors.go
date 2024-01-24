package errors

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type SrvError struct {
	// 给前端的状态码
	Code int32 `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	// 这个对应着`google.rpc.Status.message`。
	// 提供给用户的错误信息
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	// 这个对应着`google.rpc.Status.details`
	Details    []*anypb.Any `protobuf:"bytes,4,rep,name=details,proto3" json:"details,omitempty"`
	GrpcStatus int32        `protobuf:"varint,3,opt,name=grpc_status,json=grpcStatus,proto3" json:"grpc_status,omitempty"`
}
type Errors struct {
	srvErr  *SrvError
	errWrap error
}
type ErrorOption func(*Errors)

func UnwrapSvrErr(err error) *Errors {
	var se *Errors
	if errors.As(err, &se) {
		return se
	}
	return nil

}
func WithMessage(message string) ErrorOption {
	return func(e *Errors) {
		e.srvErr.Message = message
	}
}
func WithDetails(details proto.Message) ErrorOption {
	return func(e *Errors) {
		any, _ := anypb.New(details)
		e.srvErr.Details = append(e.srvErr.Details, any)
	}
}
func WithCode(code protoreflect.Enum) ErrorOption {
	return func(e *Errors) {
		e.srvErr.Code = int32(code.Number())
	}

}
func (s *Errors) Error() string {
	return s.errWrap.Error()
}

func New(srvErr error, svrMsg string, opts ...ErrorOption) *Errors {

	err := &Errors{
		errWrap: errors.Wrap(srvErr, svrMsg),
		srvErr:  &SrvError{

		},
	}
	fmt.Println(err.errWrap, srvErr)
	for _, opt := range opts {
		opt(err)
	}
	return err
}
func (s *Errors) ToGrpcStatus() *status.Status {
	return status.New(codes.Code(s.srvErr.GrpcStatus), errors.Cause(s.errWrap).Error())
}

//	func WithInternalError(message string, code int32) ErrorOption {
//		return func(e *Errors) {
//			e.srvErr.Code = code
//			e.srvErr.Message = message
//			// e.srvErr.ErrMessage = fmt.Sprintf(errMessage, args...)
//		}
//	}
func WithMsgAndCode(code int32, msg string, args ...interface{}) ErrorOption {
	return func(e *Errors) {
		e.srvErr.Code = code
		e.srvErr.Message = fmt.Sprintf(msg, args...)
	}
}

//	func WithParamsError(code int32, message string, args ...interface{}) ErrorOption {
//		return func(e *Errors) {
//			e.srvErr.Code = code
//			e.srvErr.Message = fmt.Sprintf(message, args...)
//		}
//	}
//
//	func WithNotFoundError(code int32, message string, args ...interface{}) ErrorOption {
//		return func(e *Errors) {
//			e.srvErr.Code = code
//			e.srvErr.Message = fmt.Sprintf(message, args...)
//		}
//	}
func Wrapf(err error, format string, args ...interface{}) error {
	return errutil.WrapWithDepthf(1+1, err, format, args...)
}
func Wrap(err error, msg string) error {
	return errutil.WrapWithDepth(1+1, err, msg)
}
func (s *Errors) Err() error {
	return s.errWrap
}
func (s *Errors) ClientMessage() string {
	return s.srvErr.Message
}
func (s *Errors) Code() int32 {
	return s.srvErr.Code
}
