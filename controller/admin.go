package controller

import (
	"encoding/json"
	"fmt"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type  AdminController struct {

}
//路由注册
func RegisterAdmin(group *gin.RouterGroup)  {
	adminLogin := &AdminController{}
	group.GET("/admin_info",adminLogin.AdminInfo)
	group.POST("/change_pwd",adminLogin.ChangePwd)

}
// AdminInfo godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=&dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (adminlogin *AdminController) AdminInfo(c *gin.Context) {

	//1。读取se_key对应的json 转换为结构体
	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminsessInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)),adminsessInfo);err!=nil{
		middleware.ResponseError(c,104,err)
		return
	}
	//2.读取数据然后封装
	out := &dto.AdminInfoOutput{
		ID: adminsessInfo.ID,
		Name: adminsessInfo.UserName,
		LoginTime:adminsessInfo.LoginTime,
		Avatar:"https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",  //头像
		Introduction : "A ADMINISTRATOR",
		Roles :[]string{"admin"},
	}
	middleware.ResponseSuccess(c ,out)//成功
}

// ChangePwd godoc
// @Summary 修改密码
// @Description 修改密码
// @Tags 管理员接口
// @ID /admin/change_pwd
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router  /admin/change_pwd [post]
func (adminlogin *AdminController) ChangePwd(c *gin.Context) {
	//1。读取se用户信息
	params := &dto.ChangePwdInput{}
	if err := params.BindValidParam(c);err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}

	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminsessInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)),adminsessInfo);err!=nil{
		middleware.ResponseError(c,104,err)
		return
	}
	//2.seinfo.id查询数据库 admininfo
	tx,err := lib.GetGormPool("default")//数据连接池
	if err != nil {
		middleware.ResponseError(c,102,err)
		return
	}
	adminInfo := &dao.Admin{}
	adminInfo ,err = adminInfo.Find(c,tx,(&dao.Admin{UserName: adminsessInfo.UserName}))
	if err != nil {
		middleware.ResponseError(c,103,err)
		return
	}
	//3.params.password + adminInfo,salt sha256 saltPassWord
	saltPassword := public.GenSaltPassword(adminInfo.Salt,params.PassWord)
	//4.saltPassword ==? adminInfo.password  save()
	adminInfo.PassWord = saltPassword
	if err:=adminInfo.Save(c,tx);err!=nil{
		middleware.ResponseError(c,102,err)
		return
	}
	middleware.ResponseSuccess(c ,"")//成功
}