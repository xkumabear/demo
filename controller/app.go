package controller

import (
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

func APPRegister(router *gin.RouterGroup) {
	admin := APPController{}
	router.GET("/app_list", admin.APPList)
	router.GET("/app_detail", admin.APPDetail)
	router.GET("/app_stat", admin.AppStatistics)
	router.GET("/app_delete", admin.APPDelete)
	router.POST("/app_add", admin.AppAdd)
	router.POST("/app_update", admin.AppUpdate)
}

type APPController struct {

}

// APPList godoc
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /app/app_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_no query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.APPListOutput} "success"
// @Router /app/app_list [get]
func (admin *APPController) APPList(c *gin.Context) {
	params := &dto.AppListInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	info := &dao.App{}
	list, total, err := info.APPList(c, lib.GORMDefaultPool, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	outputList := []dto.AppListItemOutput{}
	for _, item := range list {
		appCount ,err := public.FlowCounterHandler.GetCounter(public.FlowCountAppPrefix+item.AppID)
		if err != nil {
			middleware.ResponseError(c,4004,err)
			c.Abort()
			return
		}
		//appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + item.AppID)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		outputList = append(outputList, dto.AppListItemOutput{
			ID: item.ID,
			AppID: item.AppID,
			Name: item.Name,
			Secret: item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd: item.Qpd,
			Qps: item.Qps,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			RealQpd:  appCount.TotalCount,
			RealQps:  appCount.QPS,
		})
	}
	output := dto.AppListOutput{
		List:  outputList,
		Total: total,
	}
	middleware.ResponseSuccess(c, output)
	return
}

// APPDetail godoc
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /app/app_detail
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.App} "success"
// @Router /app/app_detail [get]
func (admin *APPController) APPDetail(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	detail := &dao.App{
		ID : params.ID,
		IsDelete : 0,
	}
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	detail, err = detail.Find(c, tx, detail)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	middleware.ResponseSuccess(c, detail)
	return
}

// APPDelete godoc
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /app/app_delete
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_delete [get]
func (admin *APPController) APPDelete(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	info := &dao.App{
		ID: params.ID,
	}
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	info, err = info.Find(c, tx, info)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	info.IsDelete = 1
	if err := info.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AppAdd godoc
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /app/app_add
// @Accept  json
// @Produce  json
// @Param body body dto.AppAddHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_add [post]
func (admin *APPController) AppAdd(c *gin.Context) {
	params := &dto.AppAddHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.App{
		AppID: params.AppID,
	}
	if _, err := search.Find(c, lib.GORMDefaultPool, search); err == nil {
		middleware.ResponseError(c, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	tx := lib.GORMDefaultPool
	info := &dao.App{
		AppID: params.AppID,
		Name: params.Name,
		Secret: params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps: params.Qps,
		Qpd: params.Qpd,
	}
	if err := info.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AppUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /app/app_update
// @Accept  json
// @Produce  json
// @Param body body dto.AppUpdateHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_update [post]
func (admin *APPController) AppUpdate(c *gin.Context) {
	params := &dto.AppUpdateHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	info, err := search.Find(c, tx, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	info.Name = params.Name
	info.Secret = params.Secret
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(c, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AppStatistics godoc
// @Summary 租户统计
// @Description 租户统计
// @Tags 租户管理
// @ID /app/app_stat
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /app/app_stat [get]
func (admin *APPController) AppStatistics(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	search := &dao.App{
		ID: params.ID,
	}
	detail, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := []int64{}
	appCount ,err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix+detail.AppID)
	if err != nil {
		middleware.ResponseError(c,4004,err)
		c.Abort()
		return
	}
	currentTime:= time.Now()
	for i := 0; i <= time.Now().In(lib.TimeLocation).Hour(); i++ {
		newTime := time.Date( currentTime.Year() , currentTime.Month() , currentTime.Day() , i , 0 , 0 , 0 , lib.TimeLocation )
		hourDate ,_ := appCount.GetHourData( newTime )
		fmt.Println( newTime , ":" , hourDate )
		todayStat = append( todayStat , hourDate )
	}

	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterdayTime:= currentTime.Add(-1*time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		newYesTime:=time.Date(yesterdayTime.Year(),yesterdayTime.Month(),yesterdayTime.Day(),i,0,0,0,lib.TimeLocation)
		hourData,_:=appCount.GetHourData(newYesTime)
		yesterdayStat = append(yesterdayStat, hourData)
	}
	stat := dto.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}
	middleware.ResponseSuccess(c, stat)
	return
}
