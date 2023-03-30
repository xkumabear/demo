package controller

import (
	"errors"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"time"
)

type DashboardController struct {

}
func DashboardRegister(group *gin.RouterGroup){
	services := &DashboardController{}
	group.GET("/panel_GroupDate",services.PanelGroupDate)
	group.GET("/flow_statistics",services.FlowStatistics)
	group.GET("/service_statistics",services.ServiceStatistics)
}

// PanelGroupDate godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘数据
// @ID /dashboard/panel_GroupDate
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDateOutput} "success"
// @Router  /dashboard/panel_GroupDate [get]
func (service *DashboardController) PanelGroupDate(c *gin.Context){

	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	_,serviceNumber,err :=serviceInfo.PageList(c,tx,&dto.ServiceListInput{PageSize: 1,PageNo: 1})
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}

	app := &dao.App{}
	_,appNumber,err :=app.APPList(c,tx,&dto.AppListInput{PageSize: 1,PageNo: 1})
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}

	totalCounter ,err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(c,2004,err)
		return
	}

	out := &dto.PanelGroupDateOutput{
		ServiceNumber : serviceNumber,
		AppNumber : appNumber,
		TodayRequestNumber : totalCounter.TotalCount,
		CurrentQPS : totalCounter.QPS,
	}
	middleware.ResponseSuccess(c ,out)//成功
}


// FlowStatistics godoc
// @Summary flow统计
// @Description flow统计
// @Tags 首页大盘数据
// @ID /dashboard/flow_statistics
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatisticsOutput} "success"
// @Router  /dashboard/flow_statistics [get]
func (service *DashboardController) FlowStatistics(c *gin.Context){

	counter ,err:=  public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err!= nil{
		middleware.ResponseError(c,2001,err)
		return
	}
	todayList := []int64{}
	currentTime := time.Now()
	for i := 0 ;i<currentTime.Hour();i++{
		newTime := time.Date(currentTime.Year(),currentTime.Month(),currentTime.Day(),i,0,0,0,lib.TimeLocation)
		hourDate ,_ := counter.GetHourData(newTime)
		//fmt.Println(newTime,":",hourDate)
		todayList = append(todayList,hourDate)
	}

	yesterdayList := []int64{}
	yesterdayTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23 ;i++ {
		newTime := time.Date(yesterdayTime.Year(),yesterdayTime.Month(),yesterdayTime.Day(),i,0,0,0,lib.TimeLocation)
		yesterdayDate ,_ := counter.GetHourData(newTime)
		yesterdayList = append(yesterdayList,yesterdayDate)
	}

	middleware.ResponseSuccess(c ,&dto.ServiceStatisticsOutput{
		todayList,
		yesterdayList,
	})//成功
}


// ServiceStatistics godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘数据
// @ID /dashboard/service_statistics
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashServiceStatisticsOutput} "success"
// @Router  /dashboard/service_statistics [get]
func (service *DashboardController) ServiceStatistics(c *gin.Context){
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	list,err := serviceInfo.GroupByLoadType(c,tx)
	if err != nil {
		middleware.ResponseError(c,2001,err)
		return
	}
	legend := []string{}
	for index,item := range list {
		name ,ok := public.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(c,2003,errors.New("load_type不存在"))
			return
		}
		list[index].Name = name
		legend = append(legend,name)
	}
	out := &dto.DashServiceStatisticsOutput{
		Legend: legend,
		Data: list,

	}
	middleware.ResponseSuccess(c ,out)//成功
}


