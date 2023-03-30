package dao

import (
	"fmt"
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTP          *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

var ServiceManagerHandler *ServiceManager

func init(){
	ServiceManagerHandler = NewServiceManager()
}

type ServiceManager struct {
	ServiceMap map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker  sync.RWMutex
	init sync.Once
	err error
}

func NewServiceManager() *ServiceManager{
	return &ServiceManager{
		ServiceMap: map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker: sync.RWMutex{},
		init: sync.Once{},
	}
}


func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	list := []*ServiceDetail{}
	for _ , Item := range s.ServiceSlice {
		tempItem := Item
		if tempItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list,tempItem)
		}
	}
	return list
}

func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetail {
	list := []*ServiceDetail{}
	for _ , Item := range s.ServiceSlice {
		tempItem := Item
		if tempItem.Info.LoadType == public.LoadTypeGRPC {
			list = append(list,tempItem)
		}
	}
	return list
}

//接入
func (s *ServiceManager) HttpAccessMode( c *gin.Context) (*ServiceDetail,error)  {
	//前缀匹配  ==> serviceSlice.rule
	//域名匹配 ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.path
	host := c.Request.Host

	//截取域名
	host = host[0:strings.Index(host,":")]
	fmt.Println("host",host)
	paths := c.Request.URL.Path

	for _ , serviceItem := range s.ServiceSlice {
		//item := serviceItem
		if serviceItem.Info.LoadType != public.LoadTypeHTTP{
			continue
		}
		if serviceItem.HTTP.RuleType == public.HTTPRuleTypeDomain{
			if serviceItem.HTTP.Rule == host{
				fmt.Println("Domain_item",serviceItem)
				return serviceItem, nil

			}
		}
		if serviceItem.HTTP.RuleType == public.HTTPRuleTypePrefixURL{
			if 	strings.HasPrefix( paths , serviceItem.HTTP.Rule ){
				fmt.Println("paths",serviceItem.HTTP.Rule)
				return serviceItem, nil

			}
		}
	}
	fmt.Println("errors not matched")
	return nil, errors.New("not matched service")
}


func (s *ServiceManager) LoadOnce() error  {
	s.init.Do(func() {
		//从db中分页读取基本信息

		serviceInfo := &ServiceInfo{}
		params := &dto.ServiceListInput{
			PageNo: 1,
			PageSize: 99999}
		c , _ := gin.CreateTestContext(httptest.NewRecorder())
		tx,err := lib.GetGormPool("default")//数据连接池
		if err != nil {
			s.err = err
			return
		}
		list , _ , err := serviceInfo.PageList(c,tx,params)
		if err!= nil{
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _ , listItem := range list {
			tmpItem := listItem
			serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
			if err != nil {
				s.err = err
				return
			}
			s.ServiceMap[listItem.ServiceName]  =  serviceDetail
			s.ServiceSlice =  append( s.ServiceSlice , serviceDetail)
		}
	})
	return  s.err
}