package controller

import (
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/e421083458/demo/dao"
	"github.com/e421083458/demo/dto"
	"github.com/e421083458/demo/middleware"
	"github.com/e421083458/demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type  OauthController struct {

}
//登录-路由注册
func RegisterOauth(group *gin.RouterGroup)  {
	Oauth := &OauthController{}
	group.POST("/tokens",Oauth.Tokens)

}
// Tokens godoc
// @Summary 获取TOKEN
// @Description 获取TOKEN
// @Tags OAUTH 接口
// @ID /oauth/tokens
// @Accept  json
// @Produce  json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=dto.TokensOutput} "success"
// @Router /oauth/tokens [post]
func (o *OauthController) Tokens(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(c);err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}
	str := strings.Split(c.GetHeader("Authorization"),"")
	if len(str) != 2 {
		middleware.ResponseError(c,101,errors.New("用户名或密码错误"))//失败
		return
	}
	appSecret , err := base64.StdEncoding.DecodeString(str[1])//解码
	if err != nil {
		middleware.ResponseError(c,101,err)//失败
		return
	}
	fmt.Println("appSecret",string(appSecret))
	//取出appid&secret
	//生成app_list缓存
	//匹配app_id
	//基于jwt生成token
	//生成out
	splits := strings.Split(string(appSecret),":")
	if len(splits) != 2 {
		middleware.ResponseError(c,101,errors.New("appID,err"))//失败
		return
	}
	appList := dao.AppManagerHandler.GetAppList()
	for _ , appInfo := range appList{
		if appInfo.AppID == splits[0] && appInfo.Secret == splits[1]{
			claims := jwt.StandardClaims{
				Issuer: appInfo.AppID,
				ExpiresAt: time.Now().Add(public.JwtExpiresAt*time.Second).In(lib.TimeLocation).Unix(),
			}
			token , err := public.JwtEncode(claims)
			if err != nil {
				middleware.ResponseError(c,101,errors.New("JwtEncode err"))//失败
				return
			}
			out := &dto.TokensOutput{
				ExpiresIn :  public.JwtExpiresAt,
				TokenType : "Bearer",
				AccessToken : token,
				Scope : "read_write",
			}
			middleware.ResponseSuccess(c ,out)//成功
			return
		}
	}
	middleware.ResponseError(c,2005,errors.New("未匹配到app信息"))
}