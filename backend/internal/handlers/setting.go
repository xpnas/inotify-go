package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inotify/backend/internal/auth"
	"inotify/backend/internal/models"
)

func (s *Server) RegisterSetting(r gin.IRouter) {
	r.GET("/setting/BindWeixinCallback", s.bindWeixinCallback)
	r.GET("/setting/bindWeixinCallback", s.bindWeixinCallback)

	g := r.Group("/setting", s.Auth(auth.RoleSystem, auth.RoleUser))
	g.GET("", s.settingIndex)
	g.GET("/", s.settingIndex)
	g.GET("/GetSendTemplates", s.getSendTemplates)
	g.GET("/getSendTemplates", s.getSendTemplates)
	g.GET("/GetSendAuths", s.getSendAuths)
	g.GET("/getSendAuths", s.getSendAuths)
	g.GET("/GetMessageHistories", s.getMessageHistories)
	g.GET("/getMessageHistories", s.getMessageHistories)
	g.GET("/reSendKey", s.reSendKey)
	g.POST("/ActiveSendAuth", s.activeSendAuth)
	g.POST("/activeSendAuth", s.activeSendAuth)
	g.POST("/DeleteSendAuth", s.deleteSendAuth)
	g.POST("/deleteSendAuth", s.deleteSendAuth)
	g.POST("/AddSendAuth", s.addSendAuth)
	g.POST("/addSendAuth", s.addSendAuth)
	g.POST("/ModifySendAuth", s.modifySendAuth)
	g.POST("/modifySendAuth", s.modifySendAuth)
	g.POST("/TestSendAuth", s.testSendAuth)
	g.POST("/testSendAuth", s.testSendAuth)
	g.GET("/GetWeixinBindUrl", s.getWeixinBindURL)
	g.GET("/getWeixinBindUrl", s.getWeixinBindURL)
}

func (s *Server) getMessageHistories(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := s.Store.DB.Model(&models.MessageHistory{}).Where("userId = ?", user.ID)
	if title := strings.TrimSpace(c.Query("title")); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if content := strings.TrimSpace(c.Query("content")); content != "" {
		query = query.Where("(title LIKE ? OR body LIKE ?)", "%"+content+"%", "%"+content+"%")
	}
	if success := strings.TrimSpace(c.Query("success")); success != "" {
		if success == "true" || success == "1" {
			query = query.Where("success = ?", true)
		} else if success == "false" || success == "0" {
			query = query.Where("success = ?", false)
		}
	}
	if start := parseQueryTime(c.Query("startTime"), false); !start.IsZero() {
		query = query.Where("createTime >= ?", start)
	}
	if end := parseQueryTime(c.Query("endTime"), true); !end.IsZero() {
		query = query.Where("createTime <= ?", end)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		Error(c, err.Error())
		return
	}
	var rows []models.MessageHistory
	if err := query.Order("createTime DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, gin.H{"items": rows, "total": total, "page": page, "pageSize": pageSize})
}

func (s *Server) settingIndex(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, user)
}

func (s *Server) getSendTemplates(c *gin.Context) {
	OK(c, s.Sender.Templates())
}

func (s *Server) getSendAuths(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	var auths []models.SendAuthInfo
	if err := s.Store.DB.Where("userId = ?", user.ID).Find(&auths).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, auths)
}

type sendAuthRequest struct {
	SendAuthID int                    `json:"sendAuthId" form:"sendAuthId"`
	ID         int                    `json:"id" form:"id"`
	TemplateID string                 `json:"templateID" form:"templateID"`
	Name       string                 `json:"name" form:"name"`
	Config     map[string]interface{} `json:"config" form:"config"`
	Inputs     map[string]interface{} `json:"inputs" form:"inputs"`
	Active     bool                   `json:"active" form:"active"`
	HasActive  *bool                  `json:"-" form:"-"`
}

func (s *Server) addSendAuth(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	req := parseSendAuthRequest(c)
	cfg, _ := json.Marshal(req.config())
	active := true
	if req.HasActive != nil {
		active = *req.HasActive
	}
	item := models.SendAuthInfo{
		UserID:     user.ID,
		TemplateID: req.TemplateID,
		Name:       req.Name,
		Config:     string(cfg),
		Key:        randomKey(),
		Active:     active,
		CreateTime: time.Now(),
	}
	if item.Name == "" {
		item.Name = item.TemplateID
	}
	if err := s.Store.DB.Create(&item).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, item)
}

func (s *Server) modifySendAuth(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	req := parseSendAuthRequest(c)
	id := req.SendAuthID
	if id == 0 {
		id = req.ID
	}
	var item models.SendAuthInfo
	if err := s.Store.DB.First(&item, "id = ? AND userId = ?", id, user.ID).Error; err != nil {
		Error(c, "发送配置不存在")
		return
	}
	cfg, _ := json.Marshal(req.config())
	if req.TemplateID != "" {
		item.TemplateID = req.TemplateID
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if len(cfg) > 2 {
		item.Config = string(cfg)
	}
	if req.HasActive != nil {
		item.Active = *req.HasActive
	}
	if err := s.Store.DB.Save(&item).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, item)
}

func (s *Server) testSendAuth(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	req := parseSendAuthRequest(c)
	item := models.SendAuthInfo{
		UserID:     user.ID,
		TemplateID: req.TemplateID,
		Name:       req.Name,
		Active:     true,
	}
	if req.SendAuthID > 0 || req.ID > 0 {
		id := req.SendAuthID
		if id == 0 {
			id = req.ID
		}
		if err := s.Store.DB.First(&item, "id = ? AND userId = ?", id, user.ID).Error; err != nil {
			Error(c, "发送配置不存在")
			return
		}
		if req.TemplateID != "" {
			item.TemplateID = req.TemplateID
		}
		if req.Name != "" {
			item.Name = req.Name
		}
	}
	cfg, _ := json.Marshal(req.config())
	if len(cfg) > 2 {
		item.Config = string(cfg)
	}
	if item.TemplateID == "" {
		Error(c, "templateID is required")
		return
	}
	if item.Config == "" {
		item.Config = "{}"
	}
	ok := s.Sender.TestSendAuth(item, "Inotify 通道测试", "这是一条来自 Inotify 的测试消息")
	OK(c, gin.H{"success": ok})
}

func (s *Server) activeSendAuth(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	id := ParamInt(c, "sendAuthId")
	if id == 0 {
		id = ParamInt(c, "id")
	}
	state := ParamBool(c, "state")
	var item models.SendAuthInfo
	if err := s.Store.DB.First(&item, "id = ? AND userId = ?", id, user.ID).Error; err != nil {
		Error(c, "发送配置不存在")
		return
	}
	item.Active = state
	if err := s.Store.DB.Save(&item).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) deleteSendAuth(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	id := ParamInt(c, "sendAuthId")
	if id == 0 {
		id = ParamInt(c, "id")
	}
	if err := s.Store.DB.Where("id = ? AND userId = ?", id, user.ID).Delete(&models.SendAuthInfo{}).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) reSendKey(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	id := ParamInt(c, "sendAuthId")
	var item models.SendAuthInfo
	if err := s.Store.DB.First(&item, "id = ? AND userId = ?", id, user.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, "发送配置不存在")
			return
		}
		Error(c, err.Error())
		return
	}
	item.Key = randomKey()
	if err := s.Store.DB.Save(&item).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, item.Key)
}

func (s *Server) getWeixinBindURL(c *gin.Context) {
	const weixinScanTemplateID = "B1E7D9D4-2A9C-4B5A-8E53-65CC6D8C1F20"

	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, err.Error())
		return
	}
	sendAuthID := ParamInt(c, "sendAuthId")
	if sendAuthID <= 0 {
		Error(c, "sendAuthId is required")
		return
	}
	var item models.SendAuthInfo
	if err := s.Store.DB.First(&item, "id = ? AND userId = ?", sendAuthID, user.ID).Error; err != nil {
		Error(c, "发送配置不存在")
		return
	}
	if item.TemplateID != weixinScanTemplateID && item.TemplateID != "409A30D5-ABE8-4A28-BADD-D04B9908D763" {
		Error(c, "该通道不是企业微信扫码绑定模板")
		return
	}

	corpID := strings.TrimSpace(s.Store.GetSystemValue("weixinCorpId"))
	agentID := strings.TrimSpace(s.Store.GetSystemValue("weixinAgentId"))
	redirectURI := strings.TrimSpace(Param(c, "redirectUri"))
	if corpID == "" || agentID == "" {
		Error(c, "请先在系统管理-企业微信扫码绑定配置中填写 CorpID 和 AgentID")
		return
	}
	if redirectURI == "" {
		Error(c, "redirectUri is required")
		return
	}
	redirectURI = strings.TrimRight(redirectURI, "/") + "/api/setting/BindWeixinCallback"
	state := s.signWeixinBindState(sendAuthID, user.ID, time.Now().Add(10*time.Minute).Unix())

	u := url.URL{Scheme: "https", Host: "open.weixin.qq.com", Path: "/connect/oauth2/authorize"}
	q := u.Query()
	q.Set("appid", corpID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "snsapi_base")
	q.Set("agentid", agentID)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	OK(c, u.String()+"#wechat_redirect")
}

func (s *Server) bindWeixinCallback(c *gin.Context) {
	const (
		weixinTemplateID     = "409A30D5-ABE8-4A28-BADD-D04B9908D763"
		weixinScanTemplateID = "B1E7D9D4-2A9C-4B5A-8E53-65CC6D8C1F20"
	)

	code := strings.TrimSpace(Param(c, "code"))
	state := strings.TrimSpace(Param(c, "state"))
	if code == "" {
		s.renderWeixinBindPage(c, false, "缺少 code 参数，请重新扫码")
		return
	}
	sendAuthID, userID, ok := s.parseWeixinBindState(state)
	if !ok {
		s.renderWeixinBindPage(c, false, "绑定链接无效或已过期，请重新生成二维码")
		return
	}

	corpID := strings.TrimSpace(s.Store.GetSystemValue("weixinCorpId"))
	secret := strings.TrimSpace(s.Store.GetSystemValue("weixinCorpSecret"))
	agentID := strings.TrimSpace(s.Store.GetSystemValue("weixinAgentId"))
	if corpID == "" || secret == "" || agentID == "" {
		s.renderWeixinBindPage(c, false, "系统未配置企业微信参数，请联系管理员")
		return
	}

	var item models.SendAuthInfo
	if err := s.Store.DB.First(&item, "id = ? AND userId = ?", sendAuthID, userID).Error; err != nil {
		s.renderWeixinBindPage(c, false, "通道不存在或无权限")
		return
	}
	if item.TemplateID != weixinTemplateID && item.TemplateID != weixinScanTemplateID {
		s.renderWeixinBindPage(c, false, "该通道不是企业微信扫码绑定模板")
		return
	}

	accessToken, err := s.weixinAccessToken(corpID, secret)
	if err != nil {
		s.renderWeixinBindPage(c, false, err.Error())
		return
	}
	toUser, err := s.weixinUserIDByCode(accessToken, code)
	if err != nil {
		s.renderWeixinBindPage(c, false, err.Error())
		return
	}
	cfg, _ := json.Marshal(map[string]interface{}{
		"Corpid":     corpID,
		"Corpsecret": secret,
		"AgentID":    agentID,
		"OpengId":    toUser,
	})
	if strings.TrimSpace(item.Name) == "" || item.Name == "企业微信扫码绑定" {
		item.Name = "企业微信-" + toUser
	}
	item.TemplateID = weixinTemplateID
	item.Config = string(cfg)
	item.Active = true
	if err := s.Store.DB.Save(&item).Error; err != nil {
		s.renderWeixinBindPage(c, false, "保存绑定结果失败")
		return
	}

	s.renderWeixinBindPage(c, true, "绑定成功："+toUser+"，可返回 Inotify 刷新页面")
}

func (s *Server) weixinAccessToken(corpID, secret string) (string, error) {
	api := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", url.QueryEscape(corpID), url.QueryEscape(secret))
	resp, err := http.Get(api)
	if err != nil || resp == nil {
		return "", fmt.Errorf("获取企业微信 access token 失败")
	}
	defer resp.Body.Close()
	var body struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("解析企业微信 token 响应失败")
	}
	if body.ErrCode != 0 || body.AccessToken == "" {
		if body.ErrMsg == "" {
			body.ErrMsg = "unknown error"
		}
		return "", fmt.Errorf("企业微信 token 错误: %s", body.ErrMsg)
	}
	return body.AccessToken, nil
}

func (s *Server) weixinUserIDByCode(accessToken, code string) (string, error) {
	api := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token=%s&code=%s", url.QueryEscape(accessToken), url.QueryEscape(code))
	resp, err := http.Get(api)
	if err != nil || resp == nil {
		return "", fmt.Errorf("获取企业微信用户信息失败")
	}
	defer resp.Body.Close()
	var body struct {
		UserID  string `json:"UserId"`
		OpenID  string `json:"OpenId"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("解析企业微信用户响应失败")
	}
	if body.ErrCode != 0 {
		if body.ErrMsg == "" {
			body.ErrMsg = "unknown error"
		}
		return "", fmt.Errorf("企业微信换取用户失败: %s", body.ErrMsg)
	}
	if body.UserID != "" {
		return body.UserID, nil
	}
	if body.OpenID != "" {
		return body.OpenID, nil
	}
	return "", fmt.Errorf("企业微信返回空用户标识")
}

func (s *Server) signWeixinBindState(sendAuthID, userID int, expireAt int64) string {
	payload := fmt.Sprintf("%d:%d:%d", sendAuthID, userID, expireAt)
	key := []byte(s.Store.JWTInfo.IssuerSigningKey)
	if len(key) == 0 {
		key = []byte("inotify")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + ":" + sig
}

func (s *Server) parseWeixinBindState(raw string) (int, int, bool) {
	parts := strings.Split(raw, ":")
	if len(parts) != 4 {
		return 0, 0, false
	}
	sendAuthID, err1 := strconv.Atoi(parts[0])
	userID, err2 := strconv.Atoi(parts[1])
	expireAt, err3 := strconv.ParseInt(parts[2], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, false
	}
	if time.Now().Unix() > expireAt {
		return 0, 0, false
	}
	payload := strings.Join(parts[:3], ":")
	key := []byte(s.Store.JWTInfo.IssuerSigningKey)
	if len(key) == 0 {
		key = []byte("inotify")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	expect := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expect), []byte(parts[3])) {
		return 0, 0, false
	}
	return sendAuthID, userID, true
}

func (s *Server) renderWeixinBindPage(c *gin.Context, ok bool, msg string) {
	status := "失败"
	color := "#ef4444"
	if ok {
		status = "成功"
		color = "#16a34a"
	}
	html := fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>企业微信扫码绑定</title></head><body style="font-family:Arial,sans-serif;background:#f8fafc;margin:0;padding:24px;"><div style="max-width:560px;margin:40px auto;background:#fff;border:1px solid #e2e8f0;border-radius:12px;padding:20px;"><h2 style="margin:0 0 12px;color:%s;">企业微信扫码绑定%s</h2><p style="margin:0;color:#334155;line-height:1.7;">%s</p><p style="margin:14px 0 0;color:#64748b;">可关闭当前页面，返回 Inotify 刷新通道列表。</p></div></body></html>`, color, status, msg)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func parseSendAuthRequest(c *gin.Context) sendAuthRequest {
	var req sendAuthRequest
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err == nil && raw != nil {
		if value, ok := raw["active"].(bool); ok {
			req.HasActive = &value
			req.Active = value
		}
		data, _ := json.Marshal(raw)
		_ = json.Unmarshal(data, &req)
	} else {
		Bind(c, &req)
		if value := Param(c, "active"); value != "" {
			active := ParamBool(c, "active")
			req.HasActive = &active
			req.Active = active
		}
	}
	if req.TemplateID == "" {
		req.TemplateID = Param(c, "templateID")
	}
	if req.Name == "" {
		req.Name = Param(c, "name")
	}
	return req
}

func (r sendAuthRequest) config() map[string]interface{} {
	if len(r.Config) > 0 {
		return r.Config
	}
	if len(r.Inputs) > 0 {
		return r.Inputs
	}
	return map[string]interface{}{}
}

func randomKey() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return stringsUpper(hex.EncodeToString(buf))
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
}

func parseQueryTime(value string, endOfDay bool) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			if endOfDay && layout == "2006-01-02" {
				return t.Add(24*time.Hour - time.Nanosecond)
			}
			return t
		}
	}
	return time.Time{}
}

func stringsUpper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'f' {
			b[i] = c - 32
		}
	}
	return string(b)
}
