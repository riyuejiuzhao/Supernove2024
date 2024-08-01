package svrutil

import (
	"Supernove2024/pb"
	"Supernove2024/util"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"google.golang.org/protobuf/proto"
	"strings"
)

const (
	//健康信息 redis key
	HealthHashKey            = "Hash.Health"
	HealthLockKey            = "HealthLock"
	HealthTtlFiled           = "TTL"
	HealthLastHeartBeatField = "LastHeartBeat"
	//服务信息 redis key
	ServiceHashKey = "Hash.Service"
	ServiceLockKey = "ServiceLock"
	//路由信息
	RouterHashKey = "Hash.Router"
	RouterLockKey = "RouterLock"

	InfoFiled     = "Info"
	RevisionFiled = "Revision"
)

func TryUnlock(mutex *redsync.Mutex) {
	_, err := mutex.Unlock()
	if err != nil {
		util.Error("释放redis锁失败, Name:%s, err:%v", mutex.Name(), err)
	}
}

// BaseServer 基本服务器，保持和Redis链接的能力
type BaseServer struct {
	Rdb      *redis.Client
	RedMutex *redsync.Redsync
}

// BufferServer 带数据缓存的服务器
type BufferServer struct {
	*BaseServer
	InstanceBuffer ServiceBuffer
	RouterBuffer   RouterBuffer
}

func ServiceHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceHashKey, serviceName)
}

func ServiceHashToServiceName(hash string) string {
	return strings.TrimPrefix(hash, ServiceHashKey)
}

func HealthHash(serviceName string, instanceID string) string {
	return fmt.Sprintf("%s.%s.%s", HealthHashKey, serviceName, instanceID)
}

// ServiceInfoLockName 对一个资源的分布式锁给一个名字
func ServiceInfoLockName(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceLockKey, serviceName)
}

func RouterHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", RouterHashKey, serviceName)
}

func RouterInfoLockName(serviceName string) string {
	return fmt.Sprintf("%s.%s", RouterLockKey, serviceName)
}

// HealthInfoLockName 对一个资源的分布式锁命名
func HealthInfoLockName(serviceName string) string {
	return fmt.Sprintf("%s.%s", HealthLockKey, serviceName)
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

func NewBufferSvr(redisAddress string, redisPassword string, redisDB int) *BufferServer {
	return &BufferServer{
		BaseServer:     NewBaseSvr(redisAddress, redisPassword, redisDB),
		InstanceBuffer: NewServiceBuffer(),
		RouterBuffer:   NewRouterBuffer(),
	}
}

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
		infoBytes, err := r.Rdb.HGet(hash, InfoFiled).Bytes()
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

func (r *BufferServer) FlushServiceBufferLocked(hash string, service string) error {
	mutex, err := r.LockRedis(ServiceInfoLockName(service))
	if err != nil {
		return err
	}
	defer TryUnlock(mutex)
	err = r.FlushServiceBuffer(hash, service)
	if err != nil {
		return err
	}
	return nil
}

// FlushServiceBuffer 如果Revision和Redis不同就刷新
func (r *BufferServer) FlushServiceBuffer(hash string, service string) error {
	redisRevision, err := r.Rdb.HGet(hash, RevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	} else if err != nil {
		return err
	}
	serviceInfo, ok := r.InstanceBuffer.GetServiceInfo(service)
	if !ok || serviceInfo.Revision != redisRevision {
		originRevision := int64(0)
		if ok {
			originRevision = serviceInfo.Revision
		}
		//需要更新本地缓存
		infoBytes, err := r.Rdb.HGet(hash, InfoFiled).Bytes()
		if err != nil {
			return err
		}
		serviceInfo = &pb.ServiceInfo{}
		err = proto.Unmarshal(infoBytes, serviceInfo)
		if err != nil {
			return err
		}
		r.InstanceBuffer.FlushService(serviceInfo)
		util.Info("刷新服务器缓存: %s,%v->%v", serviceInfo.ServiceName, originRevision, redisRevision)
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
