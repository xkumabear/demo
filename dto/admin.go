package dto

import (
	"github.com/e421083458/demo/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminInfoOutput struct {
	ID int `json:"id"`
	Name string `json:"name"`
	LoginTime time.Time `json:"login_time"`
	Avatar string `json:"avatar"`  //头像
	Introduction string `json:"introduction"`
	Roles []string `json:"roles"`
}

type ChangePwdInput struct {
	PassWord string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`
}
func (param *ChangePwdInput) BindValidParam(c *gin. Context) error {
	return public.DefaultGetValidParams(c,param)
}