/*
 * 
 * File: Endpoint.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 3, 2018-5-16 8:40:33 pm
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 4, 2018-5-17 1:48:01 pm
 * -----
 * Copyright 2017 - 2027
 */

package Etcd

import (
	"strconv"
)

type Endpoint struct {
	host	string
	port	int
}

func NewEndpoint(host string,port int) *Endpoint {
	ep := &Endpoint{
		host:host,
		port:port,
	}
	return ep
}


func (ep * Endpoint) ToString() string{

	port:=strconv.Itoa(ep.port) 
	return "http://"+ep.host + ":" + port;
}

// 比较
func (ep *Endpoint)equals(other *Endpoint) bool{
	return other.host == ep.host && other.port == ep.port;
}