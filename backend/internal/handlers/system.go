package handlers

import (
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"inotify/backend/internal/auth"
	"inotify/backend/internal/models"
)

func (s *Server) RegisterSystem(r gin.IRouter) {
	g := r.Group("/settingsys", s.Auth(auth.RoleSystem))
	g.GET("/GetGlobal", s.getGlobal)
	g.GET("/getGlobal", s.getGlobal)
	g.POST("/SetGlobal", s.setGlobal)
	g.POST("/setGlobal", s.setGlobal)
	g.GET("/GetJWT", s.getJWT)
	g.GET("/getJWT", s.getJWT)
	g.POST("/SetJWT", s.setJWT)
	g.POST("/setJWT", s.setJWT)
	g.POST("/DeleteUser", s.deleteUser)
	g.POST("/deleteUser", s.deleteUser)
	g.POST("/ActiveUser", s.activeUser)
	g.POST("/activeUser", s.activeUser)
	g.GET("/GetUsers", s.getUsers)
	g.GET("/getUsers", s.getUsers)
	g.GET("/GetSendInfos", s.getSendInfos)
	g.GET("/getSendInfos", s.getSendInfos)
	g.GET("/GetSendTypeInfos", s.getSendTypeInfos)
	g.GET("/getSendTypeInfos", s.getSendTypeInfos)
	g.GET("/getGithubEnable", s.getGithubEnableForSystem)
	g.GET("/Diagnostics", s.diagnostics)
	g.GET("/diagnostics", s.diagnostics)
	g.GET("/BackupDatabase", s.backupDatabase)
	g.GET("/backupDatabase", s.backupDatabase)
}

func (s *Server) getGlobal(c *gin.Context) {
	OK(c, gin.H{
		"githubClientId":     s.Store.GetSystemValue("githubClientId"),
		"githubClientSecret": s.Store.GetSystemValue("githubClientSecret"),
		"weixinCorpId":       s.Store.GetSystemValue("weixinCorpId"),
		"weixinCorpSecret":   s.Store.GetSystemValue("weixinCorpSecret"),
		"weixinAgentId":      s.Store.GetSystemValue("weixinAgentId"),
		"proxyAddress":       s.Store.GetSystemValue("proxyAddress"),
		"administrators":     s.Store.GetSystemValue("administrators"),
		"adminUserName":      s.Store.GetSystemValue("adminUserName"),
	})
}

func (s *Server) diagnostics(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		origin = scheme + "://" + c.Request.Host
	}
	proxyAddress := s.Store.GetSystemValue("proxyAddress")
	proxyOK := proxyAddress == ""
	if proxyAddress != "" {
		_, err := url.Parse(proxyAddress)
		proxyOK = err == nil
	}
	dataDirWritable := false
	if s.Store.Config.DataDir != "" {
		testFile := filepath.Join(s.Store.Config.DataDir, ".write-test")
		if err := os.WriteFile(testFile, []byte("ok"), 0600); err == nil {
			dataDirWritable = true
			_ = os.Remove(testFile)
		}
	}
	OK(c, gin.H{
		"githubConfigured": s.Store.GetSystemValue("githubClientId") != "" && s.Store.GetSystemValue("githubClientSecret") != "",
		"weixinConfigured": s.Store.GetSystemValue("weixinCorpId") != "" && s.Store.GetSystemValue("weixinCorpSecret") != "" && s.Store.GetSystemValue("weixinAgentId") != "",
		"githubCallback":   origin + "/oauth/github/callback",
		"weixinCallback":   origin + "/oauth/weixin/callback",
		"proxyConfigured":  proxyAddress != "",
		"proxyValid":       proxyOK,
		"dataDir":          s.Store.Config.DataDir,
		"dataDirWritable":  dataDirWritable,
		"databasePath":     s.Store.Config.DBPath,
	})
}

func (s *Server) backupDatabase(c *gin.Context) {
	if s.Store.Config.DBPath == "" {
		Error(c, "database path is empty")
		return
	}
	if _, err := os.Stat(s.Store.Config.DBPath); err != nil {
		Error(c, err.Error())
		return
	}
	c.FileAttachment(s.Store.Config.DBPath, "inotify-backup.db")
}

func (s *Server) setGlobal(c *gin.Context) {
	keys := []string{"githubClientId", "githubClientSecret", "weixinCorpId", "weixinCorpSecret", "weixinAgentId", "proxyAddress", "administrators", "adminUserName"}
	for _, key := range keys {
		if value, ok := ParamValue(c, key); ok {
			if err := s.Store.SetSystemValue(key, value); err != nil {
				Error(c, err.Error())
				return
			}
		}
	}
	OK(c, true)
}

func (s *Server) getJWT(c *gin.Context) {
	OK(c, s.Store.JWTInfo)
}

func (s *Server) setJWT(c *gin.Context) {
	var jwt models.JwtInfo
	if !Bind(c, &jwt) {
		Error(c, "invalid jwt payload")
		return
	}
	if err := s.Store.SaveJWT(jwt); err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) deleteUser(c *gin.Context) {
	username := Param(c, "userName")
	if username == "admin" {
		Error(c, "不能删除默认管理员")
		return
	}
	if err := s.Store.DB.Where("userName = ?", username).Delete(&models.SendUserInfo{}).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) activeUser(c *gin.Context) {
	username := Param(c, "userName")
	state := ParamBool(c, "state")
	var user models.SendUserInfo
	if err := s.Store.DB.First(&user, "userName = ?", username).Error; err != nil {
		Error(c, "用户不存在")
		return
	}
	user.Active = state
	if user.CreateTime.IsZero() {
		user.CreateTime = time.Now()
	}
	if err := s.Store.DB.Save(&user).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, true)
}

func (s *Server) getUsers(c *gin.Context) {
	var users []models.SendUserInfo
	if err := s.Store.DB.Find(&users).Error; err != nil {
		Error(c, err.Error())
		return
	}
	for i := range users {
		users[i].Password = ""
	}
	OK(c, users)
}

func (s *Server) getSendInfos(c *gin.Context) {
	var infos []models.SendInfo
	if err := s.Store.DB.Find(&infos).Error; err != nil {
		Error(c, err.Error())
		return
	}
	OK(c, infos)
}

func (s *Server) getSendTypeInfos(c *gin.Context) {
	OK(c, s.Sender.Templates())
}

func (s *Server) getGithubEnableForSystem(c *gin.Context) {
	OK(c, s.Store.GetSystemValue("githubClientId") != "" && s.Store.GetSystemValue("githubClientSecret") != "")
}
