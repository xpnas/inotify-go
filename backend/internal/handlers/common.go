package handlers

import (
	"io/fs"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"inotify/backend/internal/auth"
	"inotify/backend/internal/database"
	"inotify/backend/internal/models"
)

type Server struct {
	Store  *database.Store
	Sender SenderService
	UIRoot fs.FS // embedded frontend, nil in dev mode
}

type SenderService interface {
	Templates() []map[string]interface{}
	SendAuthTemplates() map[string]interface{}
	Send(token, key, title, body, url, group, sound string) bool
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResult{Code: 20000, Data: data})
}

func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, models.APIResult{Code: 50000, Msg: msg})
}

func Bind(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBind(dst); err == nil {
		return true
	}
	if err := c.ShouldBindJSON(dst); err == nil {
		return true
	}
	return false
}

func Param(c *gin.Context, key string) string {
	if v := c.Query(key); v != "" {
		return v
	}
	return c.PostForm(key)
}

func ParamInt(c *gin.Context, key string) int {
	n, _ := strconv.Atoi(Param(c, key))
	return n
}

func ParamBool(c *gin.Context, key string) bool {
	v := Param(c, key)
	return v == "true" || v == "True" || v == "1"
}

func (s *Server) Auth(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Token")
		if token == "" {
			token = c.GetHeader("Authorization")
		}
		if token == "" {
			Error(c, "未登录")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(s.Store.JWTInfo, token)
		if err != nil {
			Error(c, "登录失效")
			c.Abort()
			return
		}
		username, _ := claims["name"].(string)
		role, _ := claims["role"].(string)
		if username == "" {
			Error(c, "登录失效")
			c.Abort()
			return
		}
		if len(requiredRoles) > 0 {
			allowed := false
			for _, r := range requiredRoles {
				if r == role {
					allowed = true
					break
				}
			}
			if !allowed {
				Error(c, "无权限")
				c.Abort()
				return
			}
		}
		c.Set("userName", username)
		c.Set("role", role)
		c.Next()
	}
}
