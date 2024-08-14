package grpc_sdk

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClientConn struct {
	*grpc.ClientConn

	option *clientOptions
}

type clientOptions struct {
	DefaultRouterKey  string
	DefaultRouterType string

	DialOpt []grpc.DialOption
}

type ClientOption func(*clientOptions)

func WithGrpcDialOption(opts ...grpc.DialOption) ClientOption {
	return func(options *clientOptions) {
		options.DialOpt = opts
	}
}

func WithDefaultRouterType(rt string) ClientOption {
	return func(options *clientOptions) {
		options.DefaultRouterType = rt
	}
}

func WithDefaultRouterKey(key string) ClientOption {
	return func(options *clientOptions) {
		options.DefaultRouterKey = key
	}
}

// injectCallerInfo 将主调方信息写入 grpc 请求的 header 中
func injectCallerInfo(options *clientOptions) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var md metadata.MD
		var ok bool
		if md, ok = metadata.FromOutgoingContext(ctx); !ok {
			md = metadata.MD{}
		}

		routerTypes := md.Get(RouterTypeHeader)
		var routerType string
		if options.DefaultRouterType != "" {
			if routerTypes == nil || len(routerTypes) == 0 {
				routerTypes = []string{options.DefaultRouterType}
				md.Set(RouterTypeHeader, routerTypes...)
				routerType = routerTypes[0]
			} else {
				routerType = routerTypes[0]
			}
		}

		if routerType == KVRouterType {
			routerKeys := md.Get(RouterKeyHeader)
			if options.DefaultRouterKey != "" {
				if routerKeys == nil || len(routerTypes) == 0 {
					routerKeys = []string{options.DefaultRouterKey}
					md.Set(RouterKeyHeader, routerKeys...)
				}
			}
		}

		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func NewClient(address string, opts ...ClientOption) (conn *ClientConn, err error) {
	address = fmt.Sprintf("%s://%s", Scheme, address)
	option := &clientOptions{}
	for _, opt := range opts {
		opt(option)
	}

	defaultServiceConfig := fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, Scheme)
	option.DialOpt = append(option.DialOpt, grpc.WithDefaultServiceConfig(defaultServiceConfig))
	option.DialOpt = append(option.DialOpt, grpc.WithChainUnaryInterceptor(injectCallerInfo(option)))

	grpcConn, err := grpc.NewClient(address, option.DialOpt...)
	if err != nil {
		return
	}

	conn = &ClientConn{
		ClientConn: grpcConn,
		option:     option,
	}

	return
}
