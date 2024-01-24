package errors

import (
	"context"
	"runtime/debug"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetUidFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	uids := md.Get("x-uid")
	if len(uids) == 0 {
		return ""
	}
	return uids[0]
}
func GetRequestIdFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	reqids := md.Get("x-request-id")
	if len(reqids) == 0 {
		return ""
	}
	return reqids[0]
}
func GetOrSetRequestId(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
		reqid := tiga.GetUUID()
		md.Append("x-request-id", reqid)
		grpc.SetHeader(ctx, md)
		return reqid
	}
	reqids := md.Get("x-request-id")
	if len(reqids) == 0 {
		reqid := tiga.GetUUID()
		md.Append("x-request-id", reqid)
		grpc.SetHeader(ctx, md)
		return reqid
	}
	return reqids[0]
}

func InterceptorWithLogger(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 调用原始处理程序
		_, ok := metadata.FromIncomingContext(ctx)
		xRequestId := GetOrSetRequestId(ctx)
		xUid := ""
		if ok {
			xUid = GetUidFromContext(ctx)
		}
		log := logger.WithFields(logrus.Fields{
			"x-request-id": xRequestId,
			"x-uid":        xUid,
			"grpc_method":  info.FullMethod,
		},
		)
		defer func() {
			if r := recover(); r != nil {
				log.Error("Recovered in unaryInterceptor: %w", r)
				stackTrace := debug.Stack()
				log.Error(string(stackTrace))
				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()
		resp, err = handler(ctx, req)
		if err != nil {
			log := log.WithFields(logrus.Fields{
				"grpc_code": status.Code(err).String(),
				"stack":     errors.GetSafeDetails(err),
			})

			err := errors.Cause(err)
			log.Error(err.Error())
			// 可以在这里修改错误信息
			return nil, err
		}
		return resp, nil
	}
}
