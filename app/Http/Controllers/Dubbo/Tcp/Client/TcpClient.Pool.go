/*
 * 链接池管理
 * File: TcpClient.Pool.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 2, 2018-5-22 9:48:46 am
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 2, 2018-5-22 9:49:25 am
 * -----
 * Copyright 2017 - 2027
 * https://www.cnblogs.com/jkko123/p/7235257.html
 */

package TcpClient

import (
    "time"
    "sync"
    "errors"
    "net"
	"fmt"
	"strconv"
	"github.com/xoxo/crm-x/util/logger"
)

//频繁的创建和关闭连接，对系统会造成很大负担
//所以我们需要一个池子，里面事先创建好固定数量的连接资源，需要时就取，不需要就放回池中。
//但是连接资源有一个特点，我们无法保证连接长时间会有效。

//只要类型实现了ConnRes接口中的方法，就认为是一个连接资源类型
type ConnRes interface {
    Close() error;
}
 
//工厂方法，用于创建连接资源
type Factory func(address string) (ConnRes, error)
 
 
//连接池
type ConnPool struct {
    //工厂方法，创建连接资源
    factory Factory;
    //判断池是否关闭
    closed bool;
	
	// 链接池扩展
	//                     					---conn
	//                     	  ---chan *Conn
	//                     					---conn
	// address(IP:Port) <
	//                             			---conn
	//                        ---chan *Conn
	//                            			---conn
	mapNode   *sync.Map // address(IP:Port) -- conns chan *Conn;
}
 
//创建一个连接资源池
func NewConnPool(factory Factory) (*ConnPool, error) {

    cp := &ConnPool{
        factory:     factory,
        closed:      false,
	};
	cp.mapNode = new(sync.Map)
    return cp, nil;
}
func (cp *ConnPool) GetOrCreateNode(address string) (  conns *sync.Map ) {
	mapConns,ok := cp.mapNode.Load(address)
    if !ok {
		mapConns = new(sync.Map)
		cp.mapNode.Store(address,mapConns)
	}

	node2 := mapConns.(*sync.Map)
	return node2
} 

//获取连接资源
func (cp *ConnPool) Get(address string) (ConnRes, error) {
    if cp.closed {
        return nil, errors.New("连接池已关闭");
	}

	connsMap := cp.GetOrCreateNode(address)

	var connRes ConnRes = nil
	connsMap.Range(func(k, v interface{}) bool {
		connRes = v.(ConnRes)
        return false // 终止循环
    })

	if connRes != nil {
		return connRes, nil;
	}else{
		//如果无法从通道中获取资源，则重新创建一个资源返回
		connRes, err := cp.factory(address);
		if err != nil {
			return nil, err;
		}
		// 节点放入链接池
		connsMap.Store(connRes,connRes)
		logger.Info("ConnPool Get Creater A Node:"+ address)
		return connRes, nil;
	}
}
 
//连接资源放回池中
func (cp *ConnPool) RemoveConn(address string,conn ConnRes)  {
	mapConns,ok := cp.mapNode.Load(address)
    if !ok {
		conns := mapConns.(*sync.Map)
		conns.Delete(conn)
	}
	logger.Info("RemoveConn :"+ address)
}
 
func (cp *ConnPool) RemoveNode(address string)  {
	mapConns,ok := cp.mapNode.Load(address)
    if !ok {
		conns := mapConns.(*sync.Map)
		conns.Range(func(k, v interface{}) bool {
			connRes := v.(*ConnRes)
			(*connRes).Close()
			conns.Delete(k)
			return true 
		})
	}
	cp.mapNode.Delete(address)
	logger.Info("RemoveNode :"+ address)
}
//关闭连接池
func (cp *ConnPool) Close() {
    if cp.closed {
        return;
	}
	cp.closed = true
	// 关闭所有链接
	cp.mapNode.Range(func(kk, vv interface{}) bool {
		conns := vv.(*sync.Map)
		conns.Range(func(k, v interface{}) bool {
			connRes := v.(*ConnRes)
			(*connRes).Close()
			conns.Delete(k)
			return true 
		})
		cp.mapNode.Delete(kk)
		return true 
	})
}
 
//返回池中通道的长度
func (cp *ConnPool) len() (node int, conn int) {

	var lenNode int =  0
	var lenConn int =  0
	
	cp.mapNode.Range(func(kk, vv interface{}) bool {
		conns := vv.(*sync.Map)
		conns.Range(func(k, v interface{}) bool {
			connRes := v.(*ConnRes)
			(*connRes).Close()
			conns.Delete(k)
			lenConn ++
			return true 
		})
		lenNode ++
		cp.mapNode.Delete(kk)
		return true 
	})

	logger.Debug("ConnPool len node="+ strconv.Itoa(lenNode)+"conn="+ strconv.Itoa(lenNode))
	
    return lenNode,lenConn
}
 
func tests() {
 
    cp, _ := NewConnPool(func(address string) (ConnRes, error) {
        return net.Dial("tcp", ":8080");
    });
 
    //获取资源
    conn1, _ := cp.Get("127.0.0.1:8080");
    conn2, _ := cp.Get("127.0.0.1:8081");
 
	//这里连接池中资源大小为8
	node,_ :=cp.len()
    fmt.Println("cp len : ", node);
    conn1.(net.Conn).Write([]byte("hello"));
    conn2.(net.Conn).Write([]byte("world"));
    buf := make([]byte, 1024);
    n, _ := conn1.(net.Conn).Read(buf);
    fmt.Println("conn1 read : ", string(buf[:n]));
    n, _ = conn2.(net.Conn).Read(buf);
    fmt.Println("conn2 read : ", string(buf[:n]));
 
    //等待15秒
    time.Sleep(time.Second * 15);
    //我们再从池中获取资源
    conn3, _ := cp.Get("127.0.0.1:8080");
	//这里显示为0，因为池中的连接资源都超时了
	node,_ =cp.len()
    fmt.Println("cp len : ", node);
    conn3.(net.Conn).Write([]byte("test"));
    n, _ = conn3.(net.Conn).Read(buf);
    fmt.Println("conn3 read : ", string(buf[:n]));

    cp.Close();
}