package grpc_sdk

const Scheme = "mini-router"

const (
	RouterTypeHeader = "router-type"
	RouterKeyHeader  = "router-key"
)

const (
	// ConsistentRouterType 一致性哈希
	ConsistentRouterType = "Consistent"
	// RandomRouterType 随机路由
	RandomRouterType = "Random"
	// WeightedRouterType 基于权重
	WeightedRouterType = "Weighted"
	// TargetRouterType 特定路由
	TargetRouterType = "Target"
	// KVRouterType 键值对路由
	KVRouterType = "KV"
)
