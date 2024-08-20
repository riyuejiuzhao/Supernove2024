package grpc_sdk

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&miniRouterResolverBuilder{})
	balancer.Register(&miniRouterBalancerBuilder{})
}
