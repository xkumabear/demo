package dao

import (
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

//数据库结构
  type Admin struct {
	  Id int `json:"id" gorm:"primary_key" description:"主键"`
	  UserName string `json:"user_name" gorm:"column:user_name" description:"admin用户名"`
	  Salt string `json:"salt" gorm:"column:salt" description:"盐"`
	  PassWord string `json:"password" gorm:"column:password" description:"密码"`
	  UpdatedAt time.Time `json:"updated_at" gorm:"column:update_at" description:"更新时间"`
	  CreateAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	  IsDelete int `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
  }

func (t *Admin) TableName() string {
	return "gateway_admin" //表名
}

func (t *Admin ) Find( c *gin.Context,tx *gorm.DB,search *Admin) (*Admin,error) {
	admins := &Admin{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(admins).Error//设置上下文，结构体查询
	if err != nil{
		return nil, err
	}
	return admins,err
}
//1、加盐验证登录接口
func (t *Admin) LoginCheck( c *gin.Context , tx *gorm.DB, param *dto.AdminLoginInput)  (*Admin,error) {
 	 administrator,err:=t.Find(c,tx,&Admin{UserName: param.UserName , IsDelete: 0})
	if err != nil {
		return nil,errors.New("用户名错误！")//打印堆栈
	}
	saltPassword := public.GenSaltPassword(administrator.Salt,param.PassWord)
	if administrator.PassWord != saltPassword {
		return nil,errors.New("密码错误！")
	}
	return administrator,nil
}
//
func (t *Admin) Save( c *gin.Context , tx *gorm.DB)  error {
	return tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error
}