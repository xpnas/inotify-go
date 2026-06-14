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
	g.GET("/GithubBind", s.Auth(auth.RoleSystem, auth.RoleUser), s.githubBind)
	g.GET("/githubbind", s.Auth(auth.RoleSystem, auth.RoleUser), s.githubBind)
	g.POST("/GithubUnbind", s.Auth(auth.RoleSystem, auth.RoleUser), s.githubUnbind)
	g.POST("/githubunbind", s.Auth(auth.RoleSystem, auth.RoleUser), s.githubUnbind)
	g.GET("/WeixinQrEnable", s.weixinQrEnable)
	g.GET("/weixinQrEnable", s.weixinQrEnable)
	g.GET("/WeixinQrLogin", s.weixinQrLogin)
	g.GET("/weixinQrLogin", s.weixinQrLogin)
	g.GET("/WeixinQrBind", s.Auth(auth.RoleSystem, auth.RoleUser), s.weixinQrBind)
	g.GET("/weixinQrBind", s.Auth(auth.RoleSystem, auth.RoleUser), s.weixinQrBind)
	g.POST("/WeixinQrUnbind", s.Auth(auth.RoleSystem, auth.RoleUser), s.weixinQrUnbind)
	g.POST("/weixinQrUnbind", s.Auth(auth.RoleSystem, auth.RoleUser), s.weixinQrUnbind)
	g.POST("/ResetPassword", s.Auth(auth.RoleSystem, auth.RoleUser), s.resetPassword)
	g.POST("/resetPassword", s.Auth(auth.RoleSystem, auth.RoleUser), s.resetPassword)
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

	user, err := s.upsertWeixinUser(weixinUserID)
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
	OK(c, gin.H{"name": user.UserName, "role": role, "token": token, "avatar": user.Avatar, "weixinId": user.WeixinID})
}

func (s *Server) weixinQrBind(c *gin.Context) {
	corpID := strings.TrimSpace(s.Store.GetSystemValue("weixinCorpId"))
	secret := strings.TrimSpace(s.Store.GetSystemValue("weixinCorpSecret"))
	agentID := strings.TrimSpace(s.Store.GetSystemValue("weixinAgentId"))
	if corpID == "" || agentID == "" || secret == "" {
		Error(c, "未配置企业微信 CorpID / CorpSecret / AgentID，请先前往系统管理配置")
		return
	}
	code := strings.TrimSpace(Param(c, "code"))
	redirectURI := strings.TrimSpace(Param(c, "redirectUri"))
	if code == "" {
		if redirectURI == "" {
			Error(c, "redirectUri is required")
			return
		}
		OK(c, s.weixinQrAuthorizeURL(corpID, agentID, redirectURI, "inotify_weixin_bind"))
		return
	}
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
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, "用户不存在")
		return
	}
	if err := s.bindWeixinUser(user, weixinUserID); err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, gin.H{"weixinId": weixinUserID})
}

func (s *Server) weixinQrUnbind(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, "用户不存在")
		return
	}
	if !canRemoveLoginBinding(user.Password, user.GithubLogin) {
		Error(c, "当前账号没有其他可用登录方式，不能解除绑定")
		return
	}
	user.WeixinID = ""
	if err := s.Store.DB.Save(user).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) weixinQrAuthorizeURL(corpID, agentID, redirectURI, state string) string {
	u := url.URL{Scheme: "https", Host: "open.work.weixin.qq.com", Path: "/wwopen/sso/qrConnect"}
	q := u.Query()
	q.Set("appid", corpID)
	q.Set("agentid", agentID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Server) upsertWeixinUser(weixinUserID string) (*models.SendUserInfo, error) {
	var user models.SendUserInfo
	if err := s.Store.DB.First(&user, "weixinId = ?", weixinUserID).Error; err == nil {
		return &user, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	userName := "weixin_" + weixinUserID
	if err := s.Store.DB.First(&user, "userName = ?", userName).Error; err == nil {
		user.WeixinID = weixinUserID
		return &user, s.Store.DB.Save(&user).Error
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h, _ := auth.HashPassword("123456")
	user = models.SendUserInfo{
		SystemUserInfo: models.SystemUserInfo{
			UserName:   userName,
			Password:   h,
			WeixinID:   weixinUserID,
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
	redirectURI := strings.TrimSpace(Param(c, "redirectUri"))
	if code == "" {
		OK(c, s.githubAuthorizeURL(clientID, redirectURI, "login"))
		return
	}
	ghUser, err := s.githubUser(clientID, clientSecret, code, redirectURI)
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
	OK(c, gin.H{"name": user.UserName, "role": role, "token": token, "avatar": user.Avatar, "email": user.Email, "githubLogin": user.GithubLogin, "githubId": user.GithubID})
}

func (s *Server) githubBind(c *gin.Context) {
	clientID := s.Store.GetSystemValue("githubClientId")
	clientSecret := s.Store.GetSystemValue("githubClientSecret")
	if clientID == "" || clientSecret == "" {
		Error(c, "未启用GITHUB登陆")
		return
	}
	redirectURI := strings.TrimSpace(Param(c, "redirectUri"))
	code := Param(c, "code")
	if code == "" {
		OK(c, s.githubAuthorizeURL(clientID, redirectURI, "bind"))
		return
	}
	ghUser, err := s.githubUser(clientID, clientSecret, code, redirectURI)
	if err != nil {
		Error(c, err.Error())
		return
	}
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, "用户不存在")
		return
	}
	if err := s.bindGithubUser(user, ghUser); err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, gin.H{"githubLogin": ghUser.Login, "githubId": ghUser.ID})
}

func (s *Server) githubUnbind(c *gin.Context) {
	user, err := s.Store.GetUser(auth.CurrentUser(c))
	if err != nil {
		Error(c, "用户不存在")
		return
	}
	if !canRemoveLoginBinding(user.Password, user.WeixinID) {
		Error(c, "当前账号没有其他可用登录方式，不能解除绑定")
		return
	}
	user.GithubID = 0
	user.GithubLogin = ""
	if err := s.Store.DB.Save(user).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) githubAuthorizeURL(clientID, redirectURI, state string) string {
	u := url.URL{Scheme: "https", Host: "github.com", Path: "/login/oauth/authorize"}
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("scope", "read:user user:email")
	if redirectURI != "" {
		q.Set("redirect_uri", redirectURI)
	}
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Server) resetPassword(c *gin.Context) {
	username := Param(c, "username")
	password := Param(c, "password")
	oldPassword := Param(c, "oldPassword")
	if username == "" {
		Error(c, "username is required")
		return
	}
	currentUser := auth.CurrentUser(c)
	currentRoleValue, _ := c.Get("role")
	currentRole, _ := currentRoleValue.(string)
	if username != currentUser && currentRole != auth.RoleSystem {
		Error(c, "无权限")
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
	if username == currentUser && !auth.CheckPassword(user.Password, oldPassword) {
		Error(c, "旧密码错误")
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
		"name":        user.UserName,
		"roles":       []string{s.Store.Role(user.UserName)},
		"role":        s.Store.Role(user.UserName),
		"avatar":      user.Avatar,
		"email":       user.Email,
		"token":       user.Token,
		"githubLogin": user.GithubLogin,
		"githubId":    user.GithubID,
		"weixinId":    user.WeixinID,
	})
}

func (s *Server) logout(c *gin.Context) {
	OK(c, true)
}

type githubUserInfo struct {
	ID     int64  `json:"id"`
	Login  string `json:"login"`
	Avatar string `json:"avatar_url"`
	Email  string `json:"email"`
}

func (s *Server) githubUser(clientID, clientSecret, code, redirectURI string) (githubUserInfo, error) {
	tokenPayload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	if redirectURI != "" {
		tokenPayload["redirect_uri"] = redirectURI
	}
	payload, _ := json.Marshal(tokenPayload)
	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewReader(payload))
	if err != nil {
		return githubUserInfo{}, err
	}
	client, err := s.githubHTTPClient()
	if err != nil {
		return githubUserInfo{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
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
	resp, err = client.Do(req)
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

func (s *Server) githubHTTPClient() (*http.Client, error) {
	proxyAddress := strings.TrimSpace(s.Store.GetSystemValue("proxyAddress"))
	if proxyAddress == "" {
		return &http.Client{Timeout: 15 * time.Second}, nil
	}
	proxyURL, err := url.Parse(proxyAddress)
	if err != nil {
		return nil, errGithub("GitHub proxy address is invalid: " + err.Error())
	}
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}, nil
}

func (s *Server) upsertGithubUser(gh githubUserInfo) (*models.SendUserInfo, error) {
	var user models.SendUserInfo
	if gh.ID > 0 {
		if err := s.Store.DB.First(&user, "githubId = ?", gh.ID).Error; err == nil {
			user.GithubLogin = gh.Login
			user.Avatar = gh.Avatar
			user.Email = gh.Email
			return &user, s.Store.DB.Save(&user).Error
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	if gh.Login != "" {
		if err := s.Store.DB.First(&user, "githubLogin = ?", gh.Login).Error; err == nil {
			user.GithubID = gh.ID
			user.Avatar = gh.Avatar
			user.Email = gh.Email
			return &user, s.Store.DB.Save(&user).Error
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	if err := s.Store.DB.First(&user, "userName = ?", gh.Login).Error; err == nil {
		user.GithubID = gh.ID
		user.GithubLogin = gh.Login
		user.Avatar = gh.Avatar
		user.Email = gh.Email
		return &user, s.Store.DB.Save(&user).Error
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h, _ := auth.HashPassword("123456")
	user = models.SendUserInfo{
		SystemUserInfo: models.SystemUserInfo{
			UserName:    gh.Login,
			Password:    h,
			Avatar:      gh.Avatar,
			Email:       gh.Email,
			GithubID:    gh.ID,
			GithubLogin: gh.Login,
			Active:      true,
			CreateTime:  time.Now(),
		},
		Token: randomHexString(16),
	}
	return &user, s.Store.DB.Create(&user).Error
}

func (s *Server) bindGithubUser(user *models.SendUserInfo, gh githubUserInfo) error {
	var existing models.SendUserInfo
	if gh.ID > 0 {
		err := s.Store.DB.First(&existing, "githubId = ? AND id <> ?", gh.ID, user.ID).Error
		if err == nil {
			return errGithub("该 GitHub 账号已绑定其他用户")
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}
	if gh.Login != "" {
		err := s.Store.DB.First(&existing, "githubLogin = ? AND id <> ?", gh.Login, user.ID).Error
		if err == nil {
			return errGithub("该 GitHub 账号已绑定其他用户")
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}
	user.GithubID = gh.ID
	user.GithubLogin = gh.Login
	if user.Avatar == "" {
		user.Avatar = gh.Avatar
	}
	if user.Email == "" {
		user.Email = gh.Email
	}
	return s.Store.DB.Save(user).Error
}

func (s *Server) bindWeixinUser(user *models.SendUserInfo, weixinUserID string) error {
	var existing models.SendUserInfo
	err := s.Store.DB.First(&existing, "weixinId = ? AND id <> ?", weixinUserID, user.ID).Error
	if err == nil {
		return errGithub("该企业微信账号已绑定其他用户")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	user.WeixinID = weixinUserID
	return s.Store.DB.Save(user).Error
}

func canRemoveLoginBinding(password, otherBinding string) bool {
	return strings.TrimSpace(password) != "" || strings.TrimSpace(otherBinding) != ""
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
