package util

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	//服务信息 redis key
	ServiceHashKey = "Service"

	//路由信息
	RouterHashKey = "Router"
)

func InstancePrefix(serviceName string) string {
	return fmt.Sprintf("%s.%s.", ServiceHashKey, serviceName)
}

func InstanceKey2InstanceID(key string, service string) string {
	return strings.TrimPrefix(key, InstancePrefix(service))
}

func InstanceKey(serviceName string, id int64) string {
	return fmt.Sprintf("%s.%s.%v", ServiceHashKey, serviceName, id)
}

func KVRouterKey2Key(key string, service string) string {
	return strings.TrimPrefix(key, RouterKVPrefix(service))
}

func RouterKVPrefix(serviceName string) string {
	return fmt.Sprintf("%s.KV.%s.", RouterHashKey, serviceName)
}

func TargetRouterKey2InstanceID(key string, service string) (int64, error) {
	idstr := strings.TrimPrefix(key, RouterTargetPrefix(service))
	rt, err := strconv.ParseInt(idstr, 10, 64)
	return rt, err
}

func RouterTargetPrefix(serviceName string) string {
	return fmt.Sprintf("%s.Dst.%s.", RouterHashKey, serviceName)
}

func RouterKVInfoKey(serviceName string, key string) string {
	return fmt.Sprintf("%s.KV.%s.%s", RouterHashKey, serviceName, key)
}

func RouterTargetInfoKey(serviceName string, srcInstanceID int64) string {
	return fmt.Sprintf("%s.Dst.%s.%v", RouterHashKey, serviceName, srcInstanceID)
}
