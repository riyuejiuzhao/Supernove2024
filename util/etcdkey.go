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
	return fmt.Sprintf("%s.%s.", serviceName, ServiceHashKey)
}

func InstanceKey2InstanceID(key string, service string) string {
	return strings.TrimPrefix(key, InstancePrefix(service))
}

func InstanceKey(serviceName string, id int64) string {
	return fmt.Sprintf("%s.%s.%v", serviceName, ServiceHashKey, id)
}

func RouterKVPrefix(serviceName string) string {
	return fmt.Sprintf("%s.%s.KV.", serviceName, RouterHashKey)
}

func TargetRouterKey2RouterID(key string, service string) (int64, error) {
	idstr := strings.TrimPrefix(key, RouterTargetPrefix(service))
	return strconv.ParseInt(idstr, 10, 64)
}

func KVRouterKey2RouterID(key string, service string) (int64, error) {
	idstr := strings.TrimPrefix(key, RouterKVPrefix(service))
	return strconv.ParseInt(idstr, 10, 64)
}

func RouterTargetPrefix(serviceName string) string {
	return fmt.Sprintf("%s.%s.Dst.", serviceName, RouterHashKey)
}

func RouterTableKey(serviceName string) string {
	return fmt.Sprintf("%s.Table", serviceName)
}

func KVRouterInfoKey(serviceName string, routerID int64) string {
	return fmt.Sprintf("%s.%s.KV.%v", serviceName, RouterHashKey, routerID)
}

func TargetRouterInfoKey(serviceName string, routerID int64) string {
	return fmt.Sprintf("%s.%s.Dst.%v", serviceName, RouterHashKey, routerID)
}

func IsKVRouter(serviceName string, Key string) bool {
	from := len(serviceName) + len(RouterHashKey) + 2
	return Key[from:from+2] == "KV"
}

func IsDstRouter(serviceName string, Key string) bool {
	from := len(serviceName) + len(RouterHashKey) + 2
	return Key[from:from+3] == "Dst"
}

func IsRouterTable(serviceName string, Key string) bool {
	from := len(serviceName) + 1
	return Key[from:from+5] == "Table"
}

func KvRouterHashSlotKey(dstService string, dic map[string]string) string {
	return fmt.Sprintf("%s-%s", dstService, dic)
}

func TargetRouterHashSlotKey(dstServiceName string, srcInstanceName string) string {
	return fmt.Sprintf("%s-%s", dstServiceName, srcInstanceName)
}
