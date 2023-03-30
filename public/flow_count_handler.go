package public

import (
	"sync"
	"time"
)

//  创建NewFlowCounter 统计器 单例模式
var FlowCounterHandler *FlowCounter

type FlowCounter struct {
	RedisFlowCountServiceMap map[string]*RedisFlowCountService //服务类型多，快
	RedisFlowCountServiceSlice []*RedisFlowCountService//服务类型少，便利减少锁的开销
	Locker sync.RWMutex
}

func NewFlowCounter() *FlowCounter{
	return &FlowCounter{
		RedisFlowCountServiceMap: map[string]*RedisFlowCountService{},
		RedisFlowCountServiceSlice: []*RedisFlowCountService{},
		Locker: sync.RWMutex{},
	}
}

func init() {
	FlowCounterHandler = NewFlowCounter()
}

func (c *FlowCounter) GetCounter(serviceName string) (*RedisFlowCountService,error) {
	for _ ,item := range c.RedisFlowCountServiceSlice{
		if item.AppID == serviceName{
			return item,nil
		}
	}
	newCount := NewRedisFlowCountService( serviceName , 1 * time.Second)
	c.RedisFlowCountServiceSlice = append(c.RedisFlowCountServiceSlice,newCount)

	c.Locker.Lock()
	defer c.Locker.Unlock()
	c.RedisFlowCountServiceMap[serviceName] = newCount
	return newCount , nil
}