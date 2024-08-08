package svrutil

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
)

const (
	//健康信息 redis key
	HealthHashKey            = "Hash.Health"
	HealthTtlFiled           = "TTL"
	HealthLastHeartBeatField = "LastHeartBeat"

	//服务信息 redis key
	ServiceHashKey = "Hash.Service"

	//路由信息
	RouterHashKey = "Hash.Router"

	InfoField     = "Info"
	RevisionFiled = "Revision"
)

// BaseServer 基本服务器，保持和Redis链接的能力
type BaseServer struct {
	Rdb      *redis.Client
	RedMutex *redsync.Redsync
}

func ServiceHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceHashKey, serviceName)
}

func InstanceAddress(host string, port int32) string {
	return fmt.Sprintf("%v:%v", host, port)
}

func HealthHash(serviceName string, instanceID string) string {
	return fmt.Sprintf("%s.%s.%s", HealthHashKey, serviceName, instanceID)
}

func RouterIsKvField(field string) bool {
	startIndex := len(InfoField) + 1
	return field[startIndex:startIndex+2] == "KV"
}

func RouterIsDstField(field string) bool {
	startIndex := len(InfoField) + 1
	return field[startIndex:startIndex+3] == "Dst"
}

func RouterKVInfoField(key string) string {
	return fmt.Sprintf("%s.KV.%s", InfoField, key)
}

func RouterDstInfoField(key string) string {
	return fmt.Sprintf("%s.Dst.%s", InfoField, key)
}

func RouterHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", RouterHashKey, serviceName)
}

func NewBaseSvr(redisAddress string, redisPassword string, redisDB int) *BaseServer {
	//链接redis
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress, Password: redisPassword, DB: redisDB})
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	return &BaseServer{
		Rdb:      rdb,
		RedMutex: rs,
	}
}

/*
func (r *BufferServer) FlushRouterBufferLocked(hash string, service string) error {
	mutex, err := r.LockRedis(ServiceInfoLockName(service))
	if err != nil {
		return err
	}
	defer TryUnlock(mutex)
	err = r.FlushRouterBuffer(hash, service)
	if err != nil {
		return err
	}
	return nil
}

// FlushRouterBuffer 如果Revision和Redis不同就刷新
func (r *BufferServer) FlushRouterBuffer(hash string, service string) error {
	redisRevision, err := r.Rdb.HGet(hash, RevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	} else if err != nil {
		return err
	}
	routerInfo, ok := r.RouterBuffer.GetServiceRouter(service)
	if !ok || routerInfo.Revision != redisRevision {
		originRevision := int64(0)
		if ok {
			originRevision = routerInfo.Revision
		}
		//需要更新本地缓存
		infoBytes, err := r.Rdb.HGet(hash, InfoField).Bytes()
		if err != nil {
			return err
		}
		routerInfo = &pb.ServiceRouterInfo{}
		err = proto.Unmarshal(infoBytes, routerInfo)
		if err != nil {
			return err
		}
		r.RouterBuffer.FlushService(routerInfo)
		util.Info("刷新服务器路由缓存: %s,%v->%v", routerInfo.ServiceName, originRevision, redisRevision)
	}
	return nil
}

func (r *BaseServer) LockRedis(lockName string) (*redsync.Mutex, error) {
	mutex := r.RedMutex.NewMutex(lockName)
	err := mutex.Lock()
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	return mutex, nil
}
*/
