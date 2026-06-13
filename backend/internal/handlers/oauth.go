package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inotify/backend/internal/auth"
	"inotify/backend/internal/models"
)

func (s *Server) RegisterOAuth(r gin.IRouter) {
	g := r.Group("/oauth")
	g.POST("/login", LoginRateLimit(), s.login)
	g.POST("/Login", LoginRateLimit(), s.login)
	g.GET("/GithubEnable", s.githubEnable)
	g.GET("/githubenable", s.githubEnable)
	g.GET("/GithubLogin", s.githubLogin)
	g.GET("/githublogin", s.githubLogin)
	g.GET("/WeixinQrEnable", s.weixinQrEnable)
	g.GET("/weixinQrEnable", s.weixinQrEnable)
	g.GET("/WeixinQrLogin", s.weixinQrLogin)
	g.GET("/weixinQrLogin", s.weixinQrLogin)
	g.POST("/ResetPassword", s.resetPassword)
	g.POST("/resetPassword", s.resetPassword)
	g.GET("/Info", s.Auth(auth.RoleSystem, auth.RoleUser), s.info)
	g.GET("/info", s.Auth(auth.RoleSystem, auth.RoleUser), s.info)
	g.POST("/Logout", s.Auth(auth.RoleSystem, auth.RoleUser), s.logout)
	g.POST("/logout", s.Auth(auth.RoleSystem, auth.RoleUser), s.logout)
}

func (s *Server) login(c *gin.Context) {
	username := Param(c, "username")
	password := Param(c, "password")
	var req struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	if username == "" && Bind(c, &req) {
		username = req.Username
		password = req.Password
	}
	user, err := s.Store.GetUser(username)
	if err != nil || !user.Active || !auth.CheckPassword(user.Password, password) {
		RecordFailedLogin(c)
		Error(c, "用户名或密码错误")
		return
	}
	// Transparently upgrade MD5 hash to bcrypt on successful login
	if !strings.HasPrefix(user.Password, "$2") {
		if h, err := auth.HashPassword(password); err == nil {
			user.Password = h
			_ = s.Store.DB.Save(&user).Error
		}
	}
	role := s.Store.Role(username)
	token, err := auth.GenerateToken(s.Store.JWTInfo, username, role)
	if err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, gin.H{"name": username, "role": role, "token": token})
}

func (s *Server) githubEnable(c *gin.Context) {
	OK(c, s.Store.GetSystemValue("githubClientId") != "" && s.Store.GetSystemValue("githubClientSecret") != "")
}

func (s *Server) weixinQrEnable(c *gin.Context) {
	OK(c, s.Store.GetSystemValue("weixinCorpId") != "" && s.Store.GetSystemValue("weixinAgentId") != "" && s.Store.GetSystemValue("weixinCorpSecret") != "")
}

func (s *Server) weixinQrLogin(c *gin.Context) {
	corpID := strings.TrimSpace(s.Store.GetSystemValue("weixinCorpId"))
	secret := strings.TrimSpace(s.Store.GetSystemValue("weixinCorpSecret"))
	agentID := strings.TrimSpace(s.Store.GetSystemValue("weixinAgentId"))
	if corpID == "" || agentID == "" || secret == "" {
		Error(c, "未配置企业微信 CorpID / CorpSecret / AgentID，请先前往系统管理配置")
		return
	}

	code := strings.TrimSpace(Param(c, "code"))
	if code == "" {
		// 返回扫码 URL 供前端跳转
		redirectURI := strings.TrimSpace(Param(c, "redirectUri"))
		if redirectURI == "" {
			Error(c, "redirectUri is required")
			return
		}
		u := url.URL{Scheme: "https", Host: "open.work.weixin.qq.com", Path: "/wwopen/sso/qrConnect"}
		q := u.Query()
		q.Set("appid", corpID)
		q.Set("agentid", agentID)
		q.Set("redirect_uri", redirectURI)
		q.Set("state", "inotify_weixin_login")
		u.RawQuery = q.Encode()
		OK(c, u.String())
		return
	}

	// 用 code 换取企业微信用户身份
	accessToken, err := s.weixinAccessToken(corpID, secret)
	if err != nil {
		Error(c, "获取企业微信 AccessToken 失败："+err.Error())
		return
	}
	weixinUserID, err := s.weixinUserIDByCode(accessToken, code)
	if err != nil {
		Error(c, "获取企业微信用户信息失败："+err.Error())
		return
	}

	// 查找或创建本地用户，用户名以 weixin_ 为前缀
	userName := "weixin_" + weixinUserID
	user, err := s.upsertWeixinUser(userName)
	if err != nil {
		Error(c, err.Error())
		return
	}

	role := s.Store.Role(user.UserName)
	token, err := auth.GenerateToken(s.Store.JWTInfo, user.UserName, role)
	if err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, gin.H{"name": user.UserName, "role": role, "token": token, "avatar": user.Avatar})
}

func (s *Server) upsertWeixinUser(userName string) (*models.SendUserInfo, error) {
	var user models.SendUserInfo
	err := s.Store.DB.First(&user, "userName = ?", userName).Error
	if err == nil {
		return &user, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h, _ := auth.HashPassword("123456")
	user = models.SendUserInfo{
		SystemUserInfo: models.SystemUserInfo{
			UserName:   userName,
			Password:   h,
			Active:     true,
			CreateTime: time.Now(),
		},
		Token: randomHexString(16),
	}
	return &user, s.Store.DB.Create(&user).Error
}

func (s *Server) githubLogin(c *gin.Context) {
	clientID := s.Store.GetSystemValue("githubClientId")
	clientSecret := s.Store.GetSystemValue("githubClientSecret")
	if clientID == "" || clientSecret == "" {
		Error(c, "未启用GITHUB登陆")
		return
	}
	code := Param(c, "code")
	if code == "" {
		u := url.URL{Scheme: "https", Host: "github.com", Path: "/login/oauth/authorize"}
		q := u.Query()
		q.Set("client_id", clientID)
		q.Set("scope", "read:user user:email")
		u.RawQuery = q.Encode()
		OK(c, u.String())
		return
	}
	ghUser, err := s.githubUser(clientID, clientSecret, code)
	if err != nil {
		Error(c, err.Error())
		return
	}
	user, err := s.upsertGithubUser(ghUser)
	if err != nil {
		Error(c, err.Error())
		return
	}
	role := s.Store.Role(user.UserName)
	token, err := auth.GenerateToken(s.Store.JWTInfo, user.UserName, role)
	if err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, gin.H{"name": user.UserName, "role": role, "token": token, "avatar": user.Avatar, "email": user.Email})
}

func (s *Server) resetPassword(c *gin.Context) {
	username := Param(c, "username")
	password := Param(c, "password")
	if username == "" {
		Error(c, "username is required")
		return
	}
	var user models.SendUserInfo
	if err := s.Store.DB.First(&user, "userName = ?", username).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, "用户不存在")
			return
		}
		Error(c, err.Error())
		return
	}
	if password == "" {
		password = "123456"
	}
	h, err := auth.HashPassword(password)
	if err != nil {
		Error(c, err.Error())
		return
	}
	user.Password = h
	if user.CreateTime.IsZero() {
		user.CreateTime = time.Now()
	}
	if err := s.Store.DB.Save(&user).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) info(c *gin.Context) {
	username := auth.CurrentUser(c)
	user, err := s.Store.GetUser(username)
	if err != nil {
		Error(c, "用户不存在")
		return
	}
	OK(c, gin.H{
		"name":   user.UserName,
		"roles":  []string{s.Store.Role(user.UserName)},
		"role":   s.Store.Role(user.UserName),
		"avatar": user.Avatar,
		"email":  user.Email,
		"token":  user.Token,
	})
}

func (s *Server) logout(c *gin.Context) {
	OK(c, true)
}

type githubUserInfo struct {
	Login  string `json:"login"`
	Avatar string `json:"avatar_url"`
	Email  string `json:"email"`
}

func (s *Server) githubUser(clientID, clientSecret, code string) (githubUserInfo, error) {
	payload, _ := json.Marshal(map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	})
	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewReader(payload))
	if err != nil {
		return githubUserInfo{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubUserInfo{}, err
	}
	defer resp.Body.Close()
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return githubUserInfo{}, err
	}
	if tokenResp.AccessToken == "" {
		if tokenResp.Description != "" {
			return githubUserInfo{}, errGithub(tokenResp.Description)
		}
		return githubUserInfo{}, errGithub("GitHub token exchange failed")
	}
	req, err = http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return githubUserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return githubUserInfo{}, err
	}
	defer resp.Body.Close()
	var user githubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return githubUserInfo{}, err
	}
	if user.Login == "" {
		return githubUserInfo{}, errGithub("GitHub user login is empty")
	}
	return user, nil
}

func (s *Server) upsertGithubUser(gh githubUserInfo) (*models.SendUserInfo, error) {
	var user models.SendUserInfo
	if err := s.Store.DB.First(&user, "userName = ?", gh.Login).Error; err == nil {
		user.Avatar = gh.Avatar
		user.Email = gh.Email
		return &user, s.Store.DB.Save(&user).Error
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h, _ := auth.HashPassword("123456")
	user = models.SendUserInfo{
		SystemUserInfo: models.SystemUserInfo{
			UserName:   gh.Login,
			Password:   h,
			Avatar:     gh.Avatar,
			Email:      gh.Email,
			Active:     true,
			CreateTime: time.Now(),
		},
		Token: randomHexString(16),
	}
	return &user, s.Store.DB.Create(&user).Error
}

type errGithub string

func (e errGithub) Error() string { return string(e) }

func randomHexString(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return strings.ToUpper(hex.EncodeToString(buf))
}
