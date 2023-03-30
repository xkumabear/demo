package dto

import (
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"time"
)
//登录结构体
type AdminSessionInfo struct {
	ID int `json:"id"`
	UserName string `json:"username"`
	LoginTime time.Time `json:"login_time"`
}
type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"姓名" example:"admin" validate:"required,is_valid_username"`
	PassWord string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`
}

//结构体校验
func (param *AdminLoginInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c,param)
}

type AdminLoginOutput struct {
	Token string `json:"token" form:"token" comment:"token" example:"token" validate:""`
}