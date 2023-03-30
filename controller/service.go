package controller

import (
	"errors"
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type ServiceController struct {

}
func ServiceRegister(group *gin.RouterGroup){
	services := &ServiceController{}
	group.GET("/service_list",services.ServiceList)
	group.GET("/service_delete",services.ServiceDelete)
	group.GET("/service_detail",services.ServiceDetail)
	group.GET("/service_statistics",services.ServiceStatistics)
	group.POST("/service_add_http",services.ServiceAddHttp)
	group.POST("/service_update_http",services.ServiceUpdateHttp)
	group.POST("/service_add_tcp", services.ServiceAddTcp)
	group.POST("/service_update_tcp", services.ServiceUpdateTcp)
	group.POST("/service_add_grpc", services.ServiceAddGrpc)
	group.POST("/service_update_grpc", services.ServiceUpdateGrpc)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router  /service/service_list [get]
func (service *ServiceController) ServiceList(c *gin.Context){

	params := &dto.ServiceListInput{}
	if err := params.BindValidParam(c) ; err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	//从db中分页读取基本信息
	serviceInfo := &dao.ServiceInfo{}
	list , total , err := serviceInfo.PageList(c,tx,params)
	if err!= nil{
		middleware.ResponseError(c,2001,err)
		return
	}
	//格式化基本信息
	outputList := []dto.ServiceListItemOutput{}
	for _ , listItem := range list {
		serviceDetail , err := listItem.ServiceDetail(c,tx,&listItem)
		if err != nil {
			middleware.ResponseError(c,2003,err)
			return
		}
		//1、http后缀接入 ip＋port+path
		//2、http域名接入 domain
		//3、tcp grpc ip+port
		serviceAddr := "unknown"
		clusterIp := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSslPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		if serviceDetail.Info.LoadType==public.LoadTypeHTTP && serviceDetail.HTTP.RuleType == public.HTTPRuleTypePrefixURL && serviceDetail.HTTP.NeedHttps == public.HTTPNeedHttps{
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp, clusterSslPort , serviceDetail.HTTP.Rule)
		}
		if serviceDetail.Info.LoadType==public.LoadTypeHTTP && serviceDetail.HTTP.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTP.Rule
		}
		if serviceDetail.Info.LoadType==public.LoadTypeTCP {
			serviceAddr =fmt.Sprintf("%s:%d", clusterIp, serviceDetail.TCPRule.Port)
		}
		if serviceDetail.Info.LoadType==public.LoadTypeGRPC{
			serviceAddr =fmt.Sprintf("%s:%d", clusterIp, serviceDetail.GRPCRule.Port)
		}
		if serviceDetail.Info.LoadType==public.LoadTypeHTTP && serviceDetail.HTTP.RuleType == public.HTTPRuleTypePrefixURL && serviceDetail.HTTP.NeedHttps == public.HTTPNotNeedHttps{
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp , clusterPort ,serviceDetail.HTTP.Rule)
		}

		ipList := serviceDetail.LoadBalance.GetIpListByModel()
		ServiceCounter ,err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix+listItem.ServiceName)
		if err != nil {
			middleware.ResponseError(c,2004,err)
			return
		}
		outItem := dto.ServiceListItemOutput{
			ID: listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qpd: ServiceCounter.TotalCount,
			Qps: ServiceCounter.QPS,
			TotalNode: len(ipList),
		}
		outputList = append(outputList ,outItem)
	}
	out := &dto.ServiceListOutput{
		Total: total,
		List : outputList,
	}
	middleware.ResponseSuccess(c ,out)//成功


}


// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /service/service_delete [get]
func (service *ServiceController) ServiceDelete(c *gin.Context){

	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c) ; err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	//读取service基本信息
	serviceInfo := &dao.ServiceInfo{ ID : params.ID }
	serviceInfo , err  = serviceInfo.Find(c,tx,serviceInfo)
	if err!= nil{
		middleware.ResponseError(c,2001,err)
		return
	}
	//格式化基本信息
	serviceInfo.IsDelete = 1
	if err := serviceInfo.Save( c , tx ); err != nil {
		middleware.ResponseError(c,2001,err)
		return
	}
	middleware.ResponseSuccess(c ,"ok")//成功
}

// ServiceAddHttp godoc
// @Summary 添加http服务
// @Description 添加http服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /service/service_add_http [post]
func (service *ServiceController)ServiceAddHttp(c *gin.Context) {
	//1。读取se用户信息
	params := &dto.ServiceAddHttpInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 101, err) //失败
		return
	}

	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	tx = tx.Begin()//开启事务
	//校验是否被占用
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if  _, err = serviceInfo.Find(c,tx,serviceInfo) ;err == nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("服务已存在"))//pkg/errors携带错误堆栈
		return
	}

	httpUrl := &dao.HttpRule{RuleType: params.RuleType,Rule: params.Rule}
	if _,err=httpUrl.Find(c,tx,httpUrl); err == nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2004,errors.New("前缀/域名 已存在"))//pkg/errors携带错误堆栈
		return
	}
	if len(strings.Split(params.IpList,",")) != len(strings.Split(params.WeightList,",")){
		tx.Rollback()//回滚
		middleware.ResponseError(c,2005,errors.New("IP列表与权重列表数量不一致"))//pkg/errors携带错误堆栈
		return
	}
	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2006,err)
		return
	}
	httpRule := &dao.HttpRule{
		ServiceID: serviceModel.ID,
		RuleType: params.RuleType,
		Rule: params.Rule,
		NeedHttps: params.NeedHttps,
		NeedStripUri: params.NeedStripUri,
		NeedWebsocket: params.NeedWebsocket,
		UrlRewrite:params.UrlRewrite ,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID: serviceModel.ID,
		OpenAuth: params.OpenAuth,
		BlackList: params.BlackList,
		WhiteList: params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit: params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2001,err)
		return
	}

	loadBalance :=  &dao.LoadBalance{
		ServiceID: serviceModel.ID,
		RoundType: params.RoundType,
		IpList: params.IpList,
		WeightList: params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout: params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout: params.UpstreamIdleTimeout,
		UpstreamMaxIdle: params.UpstreamMaxIdle,
	}
	if err := loadBalance.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2008,err)
		return
	}
	tx.Commit()//入库

	middleware.ResponseSuccess(c,"")
}

// ServiceUpdateHttp godoc
// @Summary 更新http服务
// @Description 更新http服务
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /service/service_update_http [post]
func (service *ServiceController)ServiceUpdateHttp(c *gin.Context) {
	//1。读取se用户信息
	params := &dto.ServiceUpdateHttpInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 101, err) //失败
		return
	}

	if len(strings.Split(params.IpList,",")) != len(strings.Split(params.WeightList,",")){
		middleware.ResponseError(c,2005,errors.New("IP列表与权重列表数量不一致"))//pkg/errors携带错误堆栈
		return
	}

	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	tx = tx.Begin()//开启事务
	//校验是否被占用
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo , err = serviceInfo.Find(c,tx,serviceInfo)
	if  err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("服务不存在"))//pkg/errors携带错误堆栈
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(c,tx,serviceInfo)
	if  err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("服务不存在"))//pkg/errors携带错误堆栈
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc =params.ServiceDesc
	if err := info.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}

	httpRule := serviceDetail.HTTP
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err := httpRule.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}


	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2001,err)
		return
	}


	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadBalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadBalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadBalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadBalance.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2008,err)
		return
	}
	tx.Commit()//入库

	middleware.ResponseSuccess(c,"")
}

// ServiceAddTcp godoc
// @Summary 添加tcp服务
// @Description 添加tcp服务
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /service/service_add_tcp [post]
func (service *ServiceController)ServiceAddTcp(c *gin.Context) {
	//1。读取se用户信息
	params := &dto.ServiceAddTcpInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 101, err) //失败
		return
	}
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	//tx = tx.Begin()//开启事务
	//servicename是否占用
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName,IsDelete: 0}
	if  _, err := serviceInfo.Find(c,tx,serviceInfo) ;err == nil {
		//tx.Rollback()
		middleware.ResponseError(c,2003,errors.New("服务名已存在"))//pkg/errors携带错误堆栈
		return
	}
	//端口是否占用
	tcpUrl := &dao.TcpRule{Port: params.Port}
	if _,err=tcpUrl.Find(c,tx,tcpUrl); err == nil {
		//tx.Rollback()//回滚
		middleware.ResponseError(c,2004,errors.New("端口被另一tcpService占用"))//pkg/errors携带错误堆栈
		return
	}
	grpcUrl := &dao.GrpcRule{Port: params.Port}
	if _,err=grpcUrl.Find(c,tx,grpcUrl); err == nil {
		//tx.Rollback()//回滚
		middleware.ResponseError(c,2004,errors.New("端口被另一grpcService占用"))//pkg/errors携带错误堆栈
		return
	}

	if len(strings.Split(params.IpList,",")) != len(strings.Split(params.WeightList,",")){
		//tx.Rollback()//回滚
		middleware.ResponseError(c,2005,errors.New("IP列表与权重列表数量不一致"))//pkg/errors携带错误堆栈
		return
	}

	tx = tx.Begin()//开启事务

	tcpServiceModel := &dao.ServiceInfo{
		LoadType: public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}

	if err := tcpServiceModel.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2006,err)
		return
	}

	//loadBalance := &dao.LoadBalance{
	//	ServiceID: tcpServiceModel.ID,
	//	RoundType:  params.RoundType,
	//	IpList:     params.IpList,
	//	WeightList: params.WeightList,
	//	ForbidList: params.ForbidList,
	//}
	tcpRule := &dao.TcpRule{
		ServiceID: tcpServiceModel.ID,
		Port:      params.Port,
	}
	if err := tcpRule.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID: tcpServiceModel.ID,
		OpenAuth: params.OpenAuth,
		BlackList: params.BlackList,
		WhiteList: params.WhiteList,
		WhiteHostName: params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit: params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2001,err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID: tcpServiceModel.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2008,err)
		return
	}
	tx.Commit()//入库

	middleware.ResponseSuccess(c,"")
	return
}

// ServiceUpdateTcp godoc
// @Summary 更新tcp服务
// @Description 更新tcp服务
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /service/service_update_tcp [post]
func (service *ServiceController)ServiceUpdateTcp(c *gin.Context) {
	//1。读取se用户信息
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 101, err) //失败
		return
	}

	if len(strings.Split(params.IpList,",")) != len(strings.Split(params.WeightList,",")){
		middleware.ResponseError(c,2005,errors.New("IP列表与权重列表数量不一致"))//pkg/errors携带错误堆栈
		return
	}

	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	tx = tx.Begin()//开启事务
	//校验是否被占用
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo , err = serviceInfo.Find(c,tx,serviceInfo)
	if  err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("服务不存在"))//pkg/errors携带错误堆栈
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(c,tx,serviceInfo)
	if  err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("服务不存在"))//pkg/errors携带错误堆栈
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc =params.ServiceDesc
	if err := info.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}
	tcpRule := &dao.TcpRule{}
	if serviceDetail.TCPRule != nil {
		tcpRule = serviceDetail.TCPRule
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = params.Port
	if err := tcpRule.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}

	accessControl := &dao.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControl = serviceDetail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2001,err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadBalance = serviceDetail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2008,err)
		return
	}
	tx.Commit()//入库

	middleware.ResponseSuccess(c,"")
	return
}

// ServiceAddGrpc godoc
// @Summary 添加grpc服务
// @Description 添加grpc服务
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /service/service_add_grpc [post]
func (service *ServiceController)ServiceAddGrpc(c *gin.Context) {
	//1。读取se用户信息
	params := &dto.ServiceAddGrpcInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 101, err) //失败
		return
	}

	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	//tx = tx.Begin()//开启事务
	//servicename是否占用
	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete: 0,
	}
	if  _, err := serviceInfo.Find(c,tx,serviceInfo) ;err == nil {
		//tx.Rollback()
		middleware.ResponseError(c,2001,errors.New("服务名已存在"))//pkg/errors携带错误堆栈
		return
	}
	//端口是否占用
	tcpUrl := &dao.TcpRule{Port: params.Port}
	if _,err=tcpUrl.Find(c,tx,tcpUrl); err == nil {
		//tx.Rollback()//回滚
		middleware.ResponseError(c,2002,errors.New("端口被另一tcpService占用"))//pkg/errors携带错误堆栈
		return
	}
	grpcUrl := &dao.GrpcRule{Port: params.Port}
	if _,err=grpcUrl.Find(c,tx,grpcUrl); err == nil {
		//tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("端口被另一grpcService占用"))//pkg/errors携带错误堆栈
		return
	}

	if len(strings.Split(params.IpList,",")) != len(strings.Split(params.WeightList,",")){
		//tx.Rollback()//回滚
		middleware.ResponseError(c,2004,errors.New("IP列表与权重列表数量不一致"))//pkg/errors携带错误堆栈
		return
	}
	tx = tx.Begin()
	serviceModel := &dao.ServiceInfo{
		LoadType: public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(c,tx) ; err != nil {
		fmt.Println(serviceModel.LoadType)
		tx.Rollback()//回滚
		middleware.ResponseError(c,2005,err)
		return
	}
	grpcRule := &dao.GrpcRule{
		ServiceID: serviceModel.ID,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2006,err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID: serviceModel.ID,
		OpenAuth: params.OpenAuth,
		BlackList: params.BlackList,
		WhiteList: params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit: params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}

	loadBalance :=  &dao.LoadBalance{
		ServiceID: serviceModel.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2008,err)
		return
	}
	tx.Commit()//入库
	middleware.ResponseSuccess(c,"")
	return
}

// ServiceUpdateGrpc godoc
// @Summary 更新grpc服务
// @Description 更新grpc服务
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /service/service_update_grpc [post]
func (service *ServiceController)ServiceUpdateGrpc(c *gin.Context) {
	//1。读取se用户信息
	params := &dto.ServiceUpdateGrpcInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 101, err) //失败
		return
	}

	if len(strings.Split(params.IpList,",")) != len(strings.Split(params.WeightList,",")){
		middleware.ResponseError(c,2005,errors.New("IP列表与权重列表数量不一致"))//pkg/errors携带错误堆栈
		return
	}

	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	tx = tx.Begin()//开启事务
	//校验是否被占用

	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo , err = serviceInfo.Find(c,tx,serviceInfo)
	if  err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("服务不存在"))//pkg/errors携带错误堆栈
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(c,tx,serviceInfo)
	if  err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2003,errors.New("服务不存在"))//pkg/errors携带错误堆栈
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc =params.ServiceDesc
	if err := info.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if serviceDetail.GRPCRule != nil {
		grpcRule = serviceDetail.GRPCRule
	}
	grpcRule.ServiceID = info.ID
	//grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err := grpcRule.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2007,err)
		return
	}

	accessControl := &dao.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControl = serviceDetail.AccessControl
	}
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2001,err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadBalance = serviceDetail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c,tx) ; err != nil {
		tx.Rollback()//回滚
		middleware.ResponseError(c,2008,err)
		return
	}
	tx.Commit()//入库

	middleware.ResponseSuccess(c,"")
	return
}


// ServiceDetail godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/service_detail
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router  /service/service_detail [get]
func (service *ServiceController) ServiceDetail(c *gin.Context){

	params := &dto.ServiceDetailInput{}
	if err := params.BindValidParam(c) ; err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	//读取service基本信息
	serviceInfo := &dao.ServiceInfo{ ID : params.ID }
	serviceInfo , err  = serviceInfo.Find(c,tx,serviceInfo)
	if err!= nil{
		middleware.ResponseError(c,2001,err)
		return
	}
	serviceDetail , err := serviceInfo.ServiceDetail(c,tx,serviceInfo)
	if err!= nil{
		middleware.ResponseError(c,2001,err)
		return
	}

	middleware.ResponseSuccess(c ,serviceDetail)//成功
}

// ServiceStatistics godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/service_statistics
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatisticsOutput} "success"
// @Router  /service/service_statistics [get]
func (service *ServiceController) ServiceStatistics(c *gin.Context){

	params := &dto.ServiceDetailInput{}
	if err := params.BindValidParam(c) ; err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	//读取service基本信息
	serviceInfo := &dao.ServiceInfo{ ID : params.ID }
	//serviceInfo , err = serviceInfo.Find(c,tx,serviceInfo)
	//if err!= nil{
	//	middleware.ResponseError(c,2001,err)
	//	return
	//}
	serviceDetails , err := serviceInfo.ServiceDetail(c,tx,serviceInfo)
	if err != nil {
		middleware.ResponseError(c,2001,err)
		return
	}
	//初始化数组
	counter , err :=  public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix+serviceDetails.Info.ServiceName)
	if err != nil {
		middleware.ResponseError(c,2001,err)
		return
	}
	todayList := []int64{}
	currentTime := time.Now()
	for i := 0 ;i<currentTime.Hour();i++{
		newTime := time.Date(currentTime.Year(),currentTime.Month(),currentTime.Day(),i,0,0,0,lib.TimeLocation)
		hourDate ,_ := counter.GetHourData(newTime)
		fmt.Println(newTime,":",hourDate)
		todayList = append(todayList,hourDate)
	}

	yesterdayList := []int64{}
	yesterdayTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23 ;i++ {
		dataTime := time.Date(yesterdayTime.Year(),yesterdayTime.Month(),yesterdayTime.Day(),i,0,0,0,lib.TimeLocation)
		yesterdayDate ,_ := counter.GetHourData(dataTime)
		yesterdayList = append(yesterdayList,yesterdayDate)
	}

	middleware.ResponseSuccess(c ,&dto.ServiceStatisticsOutput{
		todayList,
		yesterdayList,
	})//成功
}