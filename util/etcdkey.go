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

/*
func KVRouterKey2Key(key string, service string) string {
	return strings.TrimPrefix(key, RouterKVPrefix(service))
}
*/

func RouterKVPrefix(serviceName string) string {
	return fmt.Sprintf("%s.KV.%s.", RouterHashKey, serviceName)
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
	return fmt.Sprintf("%s.Dst.%s.", RouterHashKey, serviceName)
}

func RouterTableKey(serviceName string) string {
	return fmt.Sprintf("Table.%s", serviceName)
}

func KVRouterInfoKey(serviceName string, routerID int64) string {
	return fmt.Sprintf("%s.KV.%s.%v", RouterHashKey, serviceName, routerID)
}

func TargetRouterInfoKey(serviceName string, routerID int64) string {
	return fmt.Sprintf("%s.Dst.%s.%v", RouterHashKey, serviceName, routerID)
}

func KvRouterHashSlotKey(dstService string, dic map[string]string) string {
	return fmt.Sprintf("%s-%s", dstService, dic)
}

func TargetRouterHashSlotKey(dstServiceName string, srcInstanceName string) string {
	return fmt.Sprintf("%s-%s", dstServiceName, srcInstanceName)
}
