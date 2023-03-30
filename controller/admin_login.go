package controller

import (
	"encoding/json"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type  AdminLoginController struct {
	
}
//登录-路由注册
func RegisterAdminLogin(group *gin.RouterGroup)  {
	adminLogin := &AdminLoginController{}
	group.POST("/login",adminLogin.AdminLogin)
	group.GET("/login_out",adminLogin.AdminLoginOut)

}
// AdminLogin godoc
// @Summary 管理员登陆
// @Description 管理员登陆
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminlogin *AdminLoginController) AdminLogin(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(c);err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}
	//取得administertor信息 ，加密加盐 ，判断是否与db中pwd相等
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	admin:=&dao.Admin{}
	admin,err = admin.LoginCheck(c,tx,params)
	if err != nil {
		middleware.ResponseError(c,103,err)
		return
	}

	//设置session
	AdSeInfo := &dto.AdminSessionInfo{
		ID: admin.Id,
		UserName: admin.UserName,
		LoginTime: time.Now(),
	}
	seBts,err := json.Marshal(AdSeInfo)
	if err != nil {
		middleware.ResponseError(c,103,err)
		return
	}
	se := sessions.Default(c)
	se.Set( public.AdminSessionInfoKey , string(seBts) )
	se.Save()

	out := &dto.AdminLoginOutput{params.UserName}
	middleware.ResponseSuccess(c ,out)//成功
}

// AdminLoginOut godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @ID /admin_login/login_out
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/login_out [get]
func (adminlogin *AdminLoginController) AdminLoginOut(c *gin.Context) {
	se := sessions.Default(c)
	se.Delete(public.AdminSessionInfoKey)
	se.Save()
	middleware.ResponseSuccess(c ,"")//成功
}