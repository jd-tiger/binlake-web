package filters

import (
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/zssky/log"

	"github.com/jd-tiger/binlake-web/models"
)

//ssoResp sso登录后返回的结构
type ssoResp struct {
	Code int             `json:"REQ_CODE"`
	Data models.UserInfo `json:"REQ_DATA"`
	Flag bool            `json:"REQ_FLAG"`
	Msg  string          `json:"REQ_MSG"`
}

const (
	//session会话超时30分钟
	sessionTimeout = 1800
)

// redirect 重定向
func redirect(ctx *context.Context) {
	url := ctx.Request.Host + ctx.Request.URL.String()
	ssoLogin := beego.AppConfig.String("sso::url") + "login?ReturnUrl=" + "http://" + url

	//ajax response code 403
	if ctx.Request.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		log.Infof("%v invalid ajax cookie", ctx)
		ctx.ResponseWriter.WriteHeader(403)
		return
	}

	ctx.Redirect(302, ssoLogin)
}

// excludePaths 不使用单点认证的 请求url
var excludePaths = []string{
	"/flow/auth/callback",
	"/login",
	"/logout",
	"/instance/create/complete",
	"/config/list",
	"/config/create",
	"/config/update",
	"/config/delete",
}

// isExclude 判断是否需要单点登录
func isExclude(path string) bool {
	log.Infof("Request Keyword:%s", path)
	isExclude := false
	for _, s := range excludePaths {
		if strings.EqualFold(s, path) || strings.HasPrefix(path, s) {
			isExclude = true
			break
		}
		if strings.HasSuffix(s, "*") && strings.HasPrefix(path, strings.TrimSuffix(s, "*")) {
			isExclude = true
			break
		}
	}

	return isExclude
}

//LoginFilter 验证用户，如果没登录，跳转到登录.
func LoginFilter(ctx *context.Context) {

	path := ctx.Request.RequestURI
	log.Infof("Request Path:%s", path)
	if isExclude(path) {
		log.Infof("Is exclude:%s", path)
		return
	}

	loginUser := ctx.Input.Session("UserInfo")
	if loginUser != nil {
		ctx.Input.SetData("UserInfo", loginUser)
		ctx.Input.SetData("SSOURL", beego.AppConfig.String("sso::url"))
		log.Infof("Is logged in:%v", loginUser)
		return
	}

	log.Infof("Redirect to Login:%s", path)
	ctx.Redirect(302, "/login")
}

// LoginTest 登录测试
func LoginTest(ctx *context.Context) {
	var sr ssoResp
	var userInfo = models.UserInfo{
		Email:     "123@163.com",
		Mobile:    "12345678",
		HrmDeptID: "3765",
		PersonID:  "00035029",
		OrgID:     "00010200",
		OrgName:   "运维研发组",
		User:      "李亚迪",
		UserID:    41868,
		UserName:  "bjliyadi",
	}
	sr.Code = 1
	sr.Data = userInfo

	ctx.Input.SetData("UserInfo", &sr.Data)

	log.Infof("%v login success:%v", ctx, sr)
}
