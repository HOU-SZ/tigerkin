package tnet

import (
	"errors"
	"fmt"
	"sync"

	"github.com/HOU-SZ/tigerkin/tiface"
)

/*
	连接管理模块
*/
type ConnManager struct {
	// 管理的连接信息
	connections map[uint32]tiface.IConnection
	// 读写连接的读写锁
	connLock sync.RWMutex
}

/*
	创建一个链接管理
*/
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]tiface.IConnection),
	}
}

// 添加链接
func (connMgr *ConnManager) Add(conn tiface.IConnection) {
	// 保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 将conn连接添加到ConnManager中
	connMgr.connections[conn.GetConnID()] = conn

	fmt.Println("connection with connID = ", conn.GetConnID(), "has been added to ConnManager successfully: conn num = ", connMgr.Len())
}

// 删除连接
func (connMgr *ConnManager) Remove(conn tiface.IConnection) {
	// 保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 删除连接信息
	delete(connMgr.connections, conn.GetConnID())

	fmt.Println("connection withcConnID = ", conn.GetConnID(), "has been removed from ConnManager successfully: conn num = ", connMgr.Len())
}

// 利用ConnID获取链接
func (connMgr *ConnManager) Get(connID uint32) (tiface.IConnection, error) {
	// 保护共享资源Map 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

// 获取当前连接
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

// 清除并停止所有连接
func (connMgr *ConnManager) ClearConn() {
	// 保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 停止全部链接并删除全部的连接信息
	for connID, conn := range connMgr.connections {
		// 停止
		conn.Stop()
		// 删除
		delete(connMgr.connections, connID)
	}

	fmt.Println("Clear All Connections successfully: conn num = ", connMgr.Len())
}
