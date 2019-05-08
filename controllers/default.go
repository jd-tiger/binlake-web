package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["appname"] = "tower_v4"
	c.TplName = "group/index.html"
}

func (c *MainController) Logout() {
	c.Ctx.Input.CruSession.Delete("UserInfo")
	c.Redirect(beego.AppConfig.String("sso::url")+"logout?ReturnUrl="+c.Ctx.Request.Referer()+"&rd="+strconv.FormatInt(time.Now().Unix(), 12), 302)
}

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

func (c *MainController) Login() {
	ctx := c.Ctx

	url := ctx.Request.Host + ctx.Request.URL.String()
	ticket := ctx.GetCookie("sso.jd.com")
	if ticket == "" {
		log.Infof("%v ticket cookie not found", url)
		redirect(ctx)
		return
	}
	ctx.Input.SetData("SSOURL", beego.AppConfig.String("sso::url"))

	ssoVerify := fmt.Sprintf("%sticket/verifyTicket?ticket=%s&url=%s&ip=%s", beego.AppConfig.String("sso::url"), ticket, url, ctx.Input.IP())
	resp, err := http.Get(ssoVerify)
	if err != nil {
		log.Infof("%v verifyTicket error:%v", ctx, err)
		redirect(ctx)
		return
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Infof("%v read resp error:%v", ctx, err)
		redirect(ctx)
		return
	}

	var sr ssoResp
	if err = json.Unmarshal(buf, &sr); err != nil {
		log.Infof("%v Unmarshal resp error:%v", ctx, err)
		redirect(ctx)
		return
	}

	if sr.Code != 1 {
		log.Infof("%v invalid resp:%+v", ctx, sr)
		redirect(ctx)
		return
	}

	userInfo := sr.Data
	//userInfo := models.UserInfo{
	//	UserName: "pengan3",
	//}

	// 从这里判断是否是超级管理员
	userInfo.IsSuperAdmin = models.IsSuperAdmin(userInfo.UserName)

	ctx.Output.Session("UserInfo", &userInfo)
	ctx.Input.SetData("UserInfo", &userInfo)

	log.Infof("login success:%v", userInfo)
	c.Redirect("/", 302)
}
