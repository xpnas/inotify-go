package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inotify/backend/internal/models"
)

func (s *Server) RegisterSend(r gin.IRouter) {
	r.GET("/send", s.send)
	r.POST("/send", s.send)
}

func (s *Server) send(c *gin.Context) {
	req := bindSendRequest(c)
	token := firstNonEmpty(Param(c, "token"), req.Token)
	key := firstNonEmpty(Param(c, "key"), req.Key)
	title := decodeMessageText(firstNonEmpty(Param(c, "title"), req.Title))
	body := decodeMessageText(firstNonEmpty(Param(c, "body"), req.Body))
	if body == "" {
		body = decodeMessageText(firstNonEmpty(Param(c, "data"), req.Data))
	}
	if body == "" {
		body = decodeMessageText(firstNonEmpty(Param(c, "content"), req.Content))
	}
	link := firstNonEmpty(Param(c, "url"), req.URL)
	group := firstNonEmpty(Param(c, "group"), req.Group)
	sound := firstNonEmpty(Param(c, "sound"), req.Sound)
	if token == "" && key == "" {
		Error(c, "token or key is required")
		return
	}
	if title == "" {
		title = "Inotify"
	}
	if s.Sender.Send(token, key, title, body, link, group, sound) {
		OK(c, true)
		return
	}
	Error(c, "send failed")
}

func (s *Server) RegisterBark(r *gin.Engine) {
	r.GET("/Ping", func(c *gin.Context) { c.String(200, "pong") })
	r.GET("/Healthz", func(c *gin.Context) { c.String(200, "ok") })
	r.GET("/Info", func(c *gin.Context) { OK(c, gin.H{"version": "inotify", "build": "go"}) })
	r.GET("/Register", s.barkRegister)
	r.POST("/Register", s.barkRegister)
	r.GET("/RegisterCheck", s.barkRegisterCheck)
	r.GET("/:key/:title", s.barkSendPath)
	r.GET("/:key/:title/:body", s.barkSendPath)
}

func (s *Server) barkRegister(c *gin.Context) {
	req := bindSendRequest(c)
	act := firstNonEmpty(Param(c, "act"), req.Act)
	deviceKey := firstNonEmpty(Param(c, "device_key"), Param(c, "deviceKey"), Param(c, "key"), req.DeviceKey, req.DeviceKeyCamel, req.Key)
	deviceToken := firstNonEmpty(Param(c, "devicetoken"), Param(c, "device_token"), Param(c, "deviceToken"), req.DeviceToken, req.DeviceTokenCamel)
	if act != "" {
		s.registerBarkDevice(c, act, deviceKey, deviceToken)
		return
	}
	if deviceToken == "" {
		Error(c, "deviceToken is required")
		return
	}
	OK(c, gin.H{"deviceToken": deviceToken})
}

func (s *Server) barkRegisterCheck(c *gin.Context) {
	deviceKey := firstNonEmpty(Param(c, "device_key"), Param(c, "deviceKey"), Param(c, "key"))
	if deviceKey == "" {
		Error(c, "device key is empty")
		return
	}
	var count int64
	if err := s.Store.DB.Model(&models.SendAuthInfo{}).Where("key = ?", deviceKey).Count(&count).Error; err != nil {
		Error(c, err.Error())
		return
	}
	if count == 0 {
		Error(c, "device not registered")
		return
	}
	OK(c, true)
}

func (s *Server) barkSendPath(c *gin.Context) {
	key := c.Param("key")
	title := decodeMessageText(c.Param("title"))
	body := decodeMessageText(c.Param("body"))
	if strings.EqualFold(key, "api") {
		return
	}
	if s.Sender.Send("", key, title, body, Param(c, "url"), Param(c, "group"), Param(c, "sound")) {
		OK(c, true)
		return
	}
	Error(c, "send failed")
}

type sendRequest struct {
	Token            string `json:"token" form:"token"`
	Key              string `json:"key" form:"key"`
	Title            string `json:"title" form:"title"`
	Body             string `json:"body" form:"body"`
	Data             string `json:"data" form:"data"`
	Content          string `json:"content" form:"content"`
	URL              string `json:"url" form:"url"`
	Group            string `json:"group" form:"group"`
	Sound            string `json:"sound" form:"sound"`
	Act              string `json:"act" form:"act"`
	DeviceKey        string `json:"device_key" form:"device_key"`
	DeviceToken      string `json:"device_token" form:"device_token"`
	DeviceKeyCamel   string `json:"deviceKey" form:"deviceKey"`
	DeviceTokenCamel string `json:"deviceToken" form:"deviceToken"`
}

func bindSendRequest(c *gin.Context) sendRequest {
	var req sendRequest
	_ = c.ShouldBind(&req)
	return req
}

func decodeMessageText(value string) string {
	return strings.NewReplacer(`\r\n`, "\n", `\n`, "\n", `\r`, "\n").Replace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *Server) registerBarkDevice(c *gin.Context, act, deviceKey, deviceToken string) {
	if act == "" {
		Error(c, "request bind failed : act is empty")
		return
	}
	if deviceToken == "" {
		Error(c, "request bind failed : device_token is empty")
		return
	}
	var user models.SendUserInfo
	if err := s.Store.DB.First(&user, "token = ?", act).Error; err != nil {
		Error(c, "request bind failed : act is not registered")
		return
	}
	if deviceKey == "" {
		deviceKey = randomKey()
	}
	cfg := map[string]string{
		"DeviceKey":         deviceKey,
		"DeviceToken":       deviceToken,
		"IsArchive":         "1",
		"AutoMaticallyCopy": "1",
		"Sound":             "1107",
	}
	data, _ := json.Marshal(cfg)
	var authInfo models.SendAuthInfo
	err := s.Store.DB.First(&authInfo, "key = ?", deviceKey).Error
	if err == nil {
		authInfo.Config = string(data)
		authInfo.Active = true
		authInfo.UserID = user.ID
		authInfo.TemplateID = "3B6DE04D-A9EF-4C91-A151-60B7425C5AB2"
		authInfo.Name = "Bark"
		s.Store.DB.Save(&authInfo)
	} else if err == gorm.ErrRecordNotFound {
		authInfo = models.SendAuthInfo{
			UserID:     user.ID,
			TemplateID: "3B6DE04D-A9EF-4C91-A151-60B7425C5AB2",
			Key:        deviceKey,
			Name:       "Bark",
			Config:     string(data),
			Active:     true,
			CreateTime: time.Now(),
		}
		s.Store.DB.Create(&authInfo)
	} else {
		Error(c, err.Error())
		return
	}
	OK(c, gin.H{"key": deviceKey, "device_key": deviceKey, "device_token": deviceToken})
}
