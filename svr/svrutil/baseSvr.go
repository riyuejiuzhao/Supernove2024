package svrutil

import (
	"Supernove2024/miniRouterProto"
	"Supernove2024/util"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"google.golang.org/protobuf/proto"
)

const (
	HealthKey            = "Health"
	ServiceKey           = "Service"
	ServiceLockKey       = "ServiceLock"
	ServiceHealthLockKey = "ServiceHealthLockKey"
	SetKey               = "Set"

	ServiceInfoFiled         = "Info"
	ServiceRevisionFiled     = "Revision"
	HealthTtlFiled           = "TTL"
	HealthLastHeartBeatField = "LastHeartBeat"
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
	Mgr ServiceBuffer
}

func ServiceHash(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceKey, serviceName)
}

func HealthHash(serviceName string, instanceID string) string {
	return fmt.Sprintf("%s.%s.%s", HealthKey, serviceName, instanceID)
}

func ServiceSetKey(serviceName string) string {
	return fmt.Sprintf("%s.%s", SetKey, serviceName)
}

// ServiceInfoLockName 对一个资源的分布式锁给一个名字
func ServiceInfoLockName(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceLockKey, serviceName)
}

// ServiceHealthInfoLockName 对一个资源的分布式锁命名
func ServiceHealthInfoLockName(serviceName string) string {
	return fmt.Sprintf("%s.%s", ServiceHealthLockKey, serviceName)
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
		BaseServer: NewBaseSvr(redisAddress, redisPassword, redisDB),
		Mgr:        NewServiceBuffer(),
	}
}

type SvrContext interface {
	GetServiceName() string
	GetServiceHash() string
}

// FlushBuffer 如果Revision和Redis不同就刷新
func (r *BufferServer) FlushBuffer(ctx SvrContext) error {
	redisRevision, err := r.Rdb.HGet(ctx.GetServiceHash(), ServiceRevisionFiled).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	} else if err != nil {
		return err
	}
	serviceInfo, ok := r.Mgr.TryGetServiceInfo(ctx.GetServiceName())
	if !ok || serviceInfo.Revision != redisRevision {
		originRevision := int64(0)
		if ok {
			originRevision = serviceInfo.Revision
		}
		//需要更新本地缓存
		infoBytes, err := r.Rdb.HGet(ctx.GetServiceHash(), ServiceInfoFiled).Bytes()
		if err != nil {
			return err
		}
		serviceInfo = &miniRouterProto.ServiceInfo{}
		err = proto.Unmarshal(infoBytes, serviceInfo)
		if err != nil {
			return err
		}
		r.Mgr.FlushService(serviceInfo)
		util.Info("刷新服务器缓存: %s,%v->%v", serviceInfo.ServiceName, originRevision, redisRevision)
	}
	return nil
}

func (r *BaseServer) LockRedisService(serviceName string) (*redsync.Mutex, error) {
	mutex := r.RedMutex.NewMutex(ServiceInfoLockName(serviceName))
	err := mutex.Lock()
	if err != nil {
		util.Error("err: %v", err)
		return nil, err
	}
	return mutex, nil
}
