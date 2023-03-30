package dao

import (
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"time"
)

type ServiceInfo struct {
	ID int64 `json:"id" gorm:"primary_key"`//ID
	ServiceName string `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	LoadType int `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`//类型
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete int8 `json:"is_delete" gorm:"column:is_delete" description:"是否删除;0否,1是"`
}

func (t *ServiceInfo) TableName() string {
	return "gateway_service_info" //表名
}

func (t *ServiceInfo ) PageList (c *gin.Context,tx *gorm.DB, param *dto.ServiceListInput) ( []ServiceInfo, int64 ,error ){
	total := int64(0)
	list :=  []ServiceInfo{}
	offset := ( param.PageNo - 1 )*param.PageSize
 	query := tx.SetCtx(public.GetGinTraceContext(c))
	query =query.Table(t.TableName()).Where("is_delete=0")

	if param.Info != "" {
		query = query.Where("(service_name like ? or service_desc like ?)" ,"%" + param.Info + "%" ,"%" + param.Info + "%")
	}
	if err := query.Limit(param.PageSize).Offset(offset).Order("id desc").Find(&list).Error;err != nil && err!=gorm.ErrRecordNotFound {
		return nil ,0 ,err
	}
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return list,total,nil
}

func (t *ServiceInfo ) ServiceDetail (c *gin.Context,tx *gorm.DB, search *ServiceInfo) ( *ServiceDetail ,error ) {
	if search.ServiceName == "" {
		info ,err := t.Find(c,tx,search)
		if err != nil{
			return nil,err
		}
		search = info
	}
	http :=  &HttpRule{ServiceID: search.ID}
	http, err :=  http.Find( c , tx , http )
	if err != nil && err!=gorm.ErrRecordNotFound{
		return nil, err
	}

	tcp :=  &TcpRule{ServiceID: search.ID}
	tcp, err =  tcp.Find( c , tx , tcp )
	if err != nil && err!=gorm.ErrRecordNotFound{
		return nil, err
	}

	grpc :=  &GrpcRule{ServiceID: search.ID}
	grpc , err =  grpc.Find( c , tx , grpc )
	if err != nil && err!=gorm.ErrRecordNotFound{
		return nil, err
	}

	accessControl :=  &AccessControl{ServiceID: search.ID}
	accessControl , err =  accessControl.Find( c , tx , accessControl )
	if err != nil && err!=gorm.ErrRecordNotFound{
		return nil, err
	}

	loadBalance :=  &LoadBalance{ServiceID: search.ID}
	loadBalance , err =  loadBalance.Find( c , tx , loadBalance )
	if err != nil && err!=gorm.ErrRecordNotFound{
		return nil, err
	}
	//组装
	detail := &ServiceDetail{
		Info: search,
		HTTP: http,
		TCPRule: tcp,
		GRPCRule: grpc,
		LoadBalance: loadBalance,
		AccessControl: accessControl,
	}
	return detail,nil
}

func (t *ServiceInfo ) GroupByLoadType (c *gin.Context,tx *gorm.DB) ( []dto.DashServiceStatisticsItemOutput,error ){
	list :=  []dto.DashServiceStatisticsItemOutput{}

	query := tx.SetCtx(public.GetGinTraceContext(c))
	err :=query.Table(t.TableName()).Where("is_delete=0").Select("load_type,count(*) as value").Group("load_type").Scan(&list).Error
	if err!= nil{
		return nil ,err
	}

	return list,nil
}




func (t *ServiceInfo ) Find( c *gin.Context,tx *gorm.DB,search *ServiceInfo) (*ServiceInfo,error) {
	out := &ServiceInfo{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(out).Error//设置上下文，结构体查询
	if err != nil{
		return nil, err
	}
	return out,err
}

func (t *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error
}
