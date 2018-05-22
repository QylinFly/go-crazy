/*
 * Etcd 封装
 * File: EtcdRegistry.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 3, 2018-5-16 8:34:28 pm
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 4, 2018-5-17 1:47:33 pm
 * -----
 * Copyright 2017 - 2027
 */

package Etcd

import (
	"net"
	"time"
	"log"
	"fmt"
	"strings"
	"strconv"
	"github.com/xoxo/crm-x/util/logger"
	. "github.com/xoxo/crm-x/Config"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Etcd/store"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Util"
)
type EtcdRegistry struct {
	addrs []string
	kv *EtcdV3
}

func NewClient(addrs []string) *EtcdRegistry {
	etcd := &EtcdRegistry{}
	etcd.addrs =	addrs
	
	kv, err := New(
		addrs,
		&store.Config{
			ConnectionTimeout: 3 * time.Second,
			Username:          "",
			Password:          "",
		},
	)
	
	if err != nil {
		logger.Info("----------------cannot create store---------------------")
	}else{
		logger.Info("---------------create store success!------------------")
	}
	etcd.kv = kv

	return etcd
}

func (e *EtcdRegistry)	TryNewEtcd()  *EtcdRegistry{

	kv, err := New(
		e.addrs,
		&store.Config{
			ConnectionTimeout: 3 * time.Second,
			Username:          "",
			Password:          "",
		},
	)

	if err != nil {
		e.kv = nil
		logger.Info("-------------------cannot create store-------------------")
	}else{
		e.kv = kv
		logger.Info("------------------create store success!------------------")
		// resp, err := e.kv.Grant(context.TODO(), 5)
		// if err != nil {
		// 	log.Fatal(err)
		// }
	}
	return e
}

func (e *EtcdRegistry)	WatchTree(key string)  (<-chan []*store.KVPair, error){

	stopCh := make(<-chan struct{})
	events, err := e.kv.WatchTree(key, stopCh, nil)
	return events, err
}
func (e *EtcdRegistry)	Find(key string)  []*Util.Endpoint{

	endpoints := []*Util.Endpoint{}
	
	if e.kv == nil	{
		e = e.TryNewEtcd()
	}
	if e.kv == nil	{
		return endpoints
	}

	k ,err :=e.kv.Get(key,nil,true)
	if err != nil {
		log.Print(err)
		return endpoints
	}

	logger.Info("发现服务节点："+key)
	endpoints = e.KVPairToEndpoint(k)

	return endpoints
}
// 数据类型转换
func (e *EtcdRegistry)KVPairToEndpoint( KVS []*store.KVPair) []*Util.Endpoint{

	endpoints := []*Util.Endpoint{}
	// var idx int = 1
	for _, ev := range KVS {
		start := strings.LastIndex(ev.Key,"/")
		if start >=0{
			endpointStr := string(ev.Key[start+1:strings.Count(ev.Key,"")-1 ])
			host := strings.Split(endpointStr,":")[0]
			portStr := strings.Split(endpointStr,":")[1];
			port,_	:=	strconv.Atoi(portStr)
			s_w := string(ev.Value[:])
			w,_	:=	strconv.Atoi(s_w) 
			ep	:=	Util.NewEndpoint(host,port,w)
			// idx++

			logger.Info("发现服务节点："+host +":"+ portStr + " >>weight = " + s_w)
			endpoints = append(endpoints, ep)
		}
	}
	return endpoints
}




func  getHostIp() string{
	addrs, err := net.InterfaceAddrs()
    if err != nil {
        logger.Info("获取本机IP地址失败！")
    }
    for _, address := range addrs {
        // 检查ip地址判断是否回环地址
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
            }
        }
	}
	return ""
}
// 向ETCD中注册服务
func (e *EtcdRegistry) Register(rootPath string,serviceName string, port int)  {
	strKey := fmt.Sprintf("/%s/%s/%s:%d",rootPath,serviceName, getHostIp(),port)

	channels := Config.Channels
	value := []byte(strconv.Itoa(channels))

	 ops := &store.WriteOptions{
		IsDir :false,
		TTL   :time.Second*10,
	 }
	e.kv.Put(strKey,value,ops,true)
}

func (e *EtcdRegistry)tests() *EtcdRegistry{

	value := []byte("bar232")
	e.kv.Put("dubbomesh--",value,nil,false)
	k ,_ :=	e.kv.Get("/dubbomesh",nil,true)
	logger.Info(k[0].Key+"  "+string(k[0].Value[:]))
	// fmt.Printf(k[0].Key+"  "+string(k[0].Value[:])+"\n")
	return e
}
