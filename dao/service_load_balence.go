package dao

import (
	"fmt"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/demo/reverse_proxy/load_balance"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoadBalance struct {
	ID            int64  `json:"id" gorm:"primary_key"`
	ServiceID     int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	CheckMethod   int    `json:"check_method" gorm:"column:check_method" description:"检查方法 tcpchk=检测端口是否握手成功	"`
	CheckTimeout  int    `json:"check_timeout" gorm:"column:check_timeout" description:"check超时时间	"`
	CheckInterval int    `json:"check_interval" gorm:"column:check_interval" description:"检查间隔, 单位s		"`
	RoundType     int    `json:"round_type" gorm:"column:round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList        string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
	WeightList    string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
	ForbidList    string `json:"forbid_list" gorm:"column:forbid_list" description:"禁用ip列表"`

	UpstreamConnectTimeout int `json:"upstream_connect_timeout" gorm:"column:upstream_connect_timeout" description:"下游建立连接超时, 单位s"`
	UpstreamHeaderTimeout  int `json:"upstream_header_timeout" gorm:"column:upstream_header_timeout" description:"下游获取header超时, 单位s	"`
	UpstreamIdleTimeout    int `json:"upstream_idle_timeout" gorm:"column:upstream_idle_timeout" description:"下游链接最大空闲时间, 单位s	"`
	UpstreamMaxIdle        int `json:"upstream_max_idle" gorm:"column:upstream_max_idle" description:"下游最大空闲链接数"`
}


func (t *LoadBalance) TableName() string {
	return "gateway_service_load_balance"
}

func (t *LoadBalance) Find(c *gin.Context, tx *gorm.DB, search *LoadBalance) (*LoadBalance, error) {
	model := &LoadBalance{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(model).Error
	return model, err
}

func (t *LoadBalance) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

//获取ip列表
func (t *LoadBalance) GetIpListByModel() []string{
	return strings.Split(t.IpList,",")
}

func (t *LoadBalance) GetWeightListByModel() []string{
	return strings.Split(t.WeightList,",")
}

var LoadBalancerHandler *LoadBalancer

type LoadBalancer struct {
	LoadBalanceMap map[string]*LoadBalancerItem //服务类型多，快
	LoadBalanceSlice []*LoadBalancerItem//服务类型少，便利减少锁的开销
	Locker sync.RWMutex
}

type LoadBalancerItem struct {
	LoadBalance load_balance.LoadBalance
	ServiceName string
}

func NewLoadBalancer() *LoadBalancer{
	return &LoadBalancer{
		LoadBalanceMap: map[string]*LoadBalancerItem{},
		LoadBalanceSlice: []*LoadBalancerItem{},
		Locker: sync.RWMutex{},
	}
}

func init() {
	LoadBalancerHandler = NewLoadBalancer()
}

func (l *LoadBalancer)GetLoadBalancer(service *ServiceDetail) (load_balance.LoadBalance,error){

	for _ ,LoadBalanceItem := range l.LoadBalanceSlice{
		if LoadBalanceItem.ServiceName == service.Info.ServiceName{
			return LoadBalanceItem.LoadBalance,nil
		}
	}

	schema := "http://"
	if service.HTTP.NeedHttps == public.HTTPNeedHttps{
		schema = "https://"
	}
	if service.Info.LoadType==public.LoadTypeTCP || service.Info.LoadType==public.LoadTypeGRPC{
		schema = ""
	}
	//prefix := ""
	//
	//if service.HTTP.RuleType == public.HTTPRuleTypePrefixURL {
	//	prefix = service.HTTP.Rule
	//}
	//组合
	ipList := service.LoadBalance.GetIpListByModel()
	ipWeightList := service.LoadBalance.GetWeightListByModel()
	ipConf := map[string]string{}
	for ipIndex ,ipItem := range ipList{
		ipConf[ipItem] = ipWeightList[ipIndex]
	}
	fmt.Println("ipConf",ipConf)
	//ipConf:=map[string]string{"127.0.0.1:2003": "20", "127.0.0.1:2004": "20"}
	mConf, err := load_balance.NewLoadBalanceCheckConf(
		//fmt.Sprintf(fmt.Sprintf("%s%s", schema, "%s")),
		fmt.Sprintf("%s%s", schema, "%s"),
		ipConf)
	if err != nil {
		return nil,err
	}
	//save to map and slice
	lb := load_balance.LoadBanlanceFactorWithConf(load_balance.LbType(service.LoadBalance.RoundType), mConf)
	LoadBalanceItem := &LoadBalancerItem{
		LoadBalance : lb,
		ServiceName : service.Info.ServiceName,
	}
	l.LoadBalanceSlice = append(l.LoadBalanceSlice,LoadBalanceItem)

	l.Locker.Lock()
	defer l.Locker.Unlock()
	l.LoadBalanceMap[service.Info.ServiceName] = LoadBalanceItem
	return lb , nil

}
//   连接池
var TransporterHandler *Transporter

type Transporter struct {
	TransporterMap  map[string]*TransporterItem //服务类型多，快
	TransporterSlice []*TransporterItem//服务类型少，便利减少锁的开销
	Locker sync.RWMutex
}
type TransporterItem struct {
	Trans *http.Transport
	ServiceName string
}

func NewTransporter() *Transporter{
	return &Transporter{
		TransporterMap: map[string]*TransporterItem{},
		TransporterSlice: []*TransporterItem{},
		Locker: sync.RWMutex{},
	}
}

func init() {
	TransporterHandler = NewTransporter()
}

func (T *Transporter)GetTransporter(service *ServiceDetail) (*http.Transport,error){
	for _ ,transportItem := range T.TransporterSlice{
		if transportItem.ServiceName == service.Info.ServiceName{
			return transportItem.Trans,nil
		}
	}
	//todo 优化点5
	if service.LoadBalance.UpstreamConnectTimeout==0{
		service.LoadBalance.UpstreamConnectTimeout = 30
	}
	if service.LoadBalance.UpstreamMaxIdle==0{
		service.LoadBalance.UpstreamMaxIdle = 100
	}
	if service.LoadBalance.UpstreamIdleTimeout==0{
		service.LoadBalance.UpstreamIdleTimeout = 90
	}
	if service.LoadBalance.UpstreamHeaderTimeout==0{
		service.LoadBalance.UpstreamHeaderTimeout = 30
	}
	//
	trans := &http.Transport{
		//
		Proxy: http.ProxyFromEnvironment,
		//
		DialContext: (&net.Dialer{
			Timeout:  time.Duration(service.LoadBalance.UpstreamConnectTimeout)*time.Second,
			//
			KeepAlive: 30 * time.Second,
			DualStack: true,
			//
		}).DialContext,
		MaxIdleConns:          service.LoadBalance.UpstreamMaxIdle,
		IdleConnTimeout:       time.Duration(service.LoadBalance.UpstreamIdleTimeout)*time.Second,
		ResponseHeaderTimeout:       time.Duration(service.LoadBalance.UpstreamHeaderTimeout)*time.Second,
		//
		TLSHandshakeTimeout:   10 * time.Second,
		//
	}
	//save to map and slice
	transItem := &TransporterItem{
		Trans : trans,
		ServiceName : service.Info.ServiceName,
	}
	T.TransporterSlice = append(T.TransporterSlice,transItem)

	T.Locker.Lock()
	defer T.Locker.Unlock()
	T.TransporterMap[service.Info.ServiceName] = transItem
	return trans , nil
}
