package public

import (
	"golang.org/x/time/rate"
	"sync"
)

//  创建NewFlowCounter 统计器 单例模式
var FlowLimitHandler *FlowLimiter

type FlowLimiter struct {
	FlowLimitServiceMap map[string]*FlowLimiterItem //服务类型多，快
	FlowLimitServiceSlice []*FlowLimiterItem//服务类型少，便利减少锁的开销
	Locker sync.RWMutex
}

func NewFlowLimiter() *FlowLimiter{
	return  &FlowLimiter{
		FlowLimitServiceMap : map[string]*FlowLimiterItem{},
		FlowLimitServiceSlice : []*FlowLimiterItem{},
		Locker : sync.RWMutex{},
	}
}

type FlowLimiterItem struct {
	ServiceName string
	Limiter *rate.Limiter
}

func init() {
	FlowLimitHandler = NewFlowLimiter()
}

func (f *FlowLimiter) GetFlowLimiter(serviceName string,qps float64) (*rate.Limiter,error) {
	for _ ,item := range f.FlowLimitServiceSlice{
		if item.ServiceName == serviceName{
			return item.Limiter,nil
		}
	}

	newLimiter  := rate.NewLimiter(rate.Limit(qps),int(qps*3))
	item := &FlowLimiterItem{
		serviceName,
		newLimiter,
	}
	f.FlowLimitServiceSlice = append(f.FlowLimitServiceSlice,item)

	f.Locker.Lock()
	defer f.Locker.Unlock()
	f.FlowLimitServiceMap[serviceName] = item
	return newLimiter , nil
}