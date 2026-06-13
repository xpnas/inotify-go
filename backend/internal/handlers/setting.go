package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"inotify/backend/internal/auth"
	"inotify/backend/internal/models"
)

func (s *Server) RegisterSetting(r gin.IRouter) {
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
