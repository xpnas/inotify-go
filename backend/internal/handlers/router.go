package handlers

import (
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"inotify/backend/internal/auth"
)

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	api := r.Group("/api")
	s.RegisterOAuth(api)
	s.RegisterSetting(api)
	s.RegisterSystem(api)
	s.RegisterSend(api)

	r.GET("/Ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	r.GET("/Healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/Info", func(c *gin.Context) { OK(c, gin.H{"version": "inotify", "build": "go"}) })
	r.GET("/Register", s.barkRegister)
	r.POST("/Register", s.barkRegister)
	r.GET("/RegisterCheck", func(c *gin.Context) { OK(c, true) })
	r.NoRoute(s.noRoute)
	return r
}

func (s *Server) noRoute(c *gin.Context) {
	path := strings.Trim(c.Request.URL.Path, "/")
	if strings.HasPrefix(path, "api/") || path == "api" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found"})
		return
	}

	// 32-hex send key shortcut (e.g. /KEY/title/body)
	parts := strings.Split(path, "/")
	if c.Request.Method == http.MethodGet && len(parts) >= 2 {
		key := parts[0]
		if len(key) == 32 && isHexString(key) {
			title := parts[1]
			body := ""
			if len(parts) >= 3 {
				body = parts[2]
			}
			if s.Sender.Send("", key, title, body, Param(c, "url"), Param(c, "group"), Param(c, "sound")) {
				OK(c, true)
				return
			}
			Error(c, "发送失败")
			return
		}
	}

	// Serve embedded frontend (SPA fallback to index.html)
	if s.UIRoot != nil {
		serveUI(c, s.UIRoot)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found", "role": auth.RoleUser})
}

func serveUI(c *gin.Context, root fs.FS) {
	urlPath := strings.TrimLeft(c.Request.URL.Path, "/")
	f, err := root.Open(urlPath)
	if err == nil {
		f.Close()
		http.FileServer(http.FS(root)).ServeHTTP(c.Writer, c.Request)
		return
	}
	// SPA fallback
	index, err := root.Open("index.html")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer index.Close()
	stat, _ := index.Stat()
	c.Header("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(c.Writer, c.Request, "index.html", stat.ModTime(), index.(io.ReadSeeker))
}

func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
