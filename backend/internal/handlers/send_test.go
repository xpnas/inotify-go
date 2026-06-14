package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"inotify/backend/internal/models"
)

type captureSender struct {
	token string
	key   string
	title string
	body  string
}

func (s *captureSender) Templates() []map[string]interface{}       { return nil }
func (s *captureSender) SendAuthTemplates() map[string]interface{} { return nil }
func (s *captureSender) Send(token, key, title, body, _, _, _ string) bool {
	s.token, s.key, s.title, s.body = token, key, title, body
	return true
}
func (s *captureSender) TestSendAuth(_ models.SendAuthInfo, title, body string) models.SendResult {
	s.title, s.body = title, body
	return models.SendResult{Success: true}
}

func TestSendAcceptsJSONPostAndEscapedNewline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sender := &captureSender{}
	server := &Server{Sender: sender}
	r := gin.New()
	server.RegisterSend(r.Group("/api"))

	req := httptest.NewRequest(http.MethodPost, "/api/send", strings.NewReader(`{"token":"tok","title":"hello","data":"line1\\nline2"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if sender.token != "tok" || sender.title != "hello" || sender.body != "line1\nline2" {
		t.Fatalf("captured = %#v", sender)
	}
}

func TestRegisterBarkDeviceCreatesAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := openTestStore(t)
	user := models.SendUserInfo{
		SystemUserInfo: models.SystemUserInfo{UserName: "u", Active: true, CreateTime: time.Now()},
		Token:          "ACT",
	}
	if err := store.DB.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	server := &Server{Store: store, Sender: &captureSender{}}
	r := gin.New()
	server.RegisterBark(r)

	req := httptest.NewRequest(http.MethodPost, "/Register", strings.NewReader(`{"act":"ACT","device_key":"DEVKEY","device_token":"DEVTOKEN"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var authInfo models.SendAuthInfo
	if err := store.DB.First(&authInfo, "key = ?", "DEVKEY").Error; err != nil {
		t.Fatal(err)
	}
	if authInfo.TemplateID != "3B6DE04D-A9EF-4C91-A151-60B7425C5AB2" || !authInfo.Active {
		t.Fatalf("bad auth info: %#v", authInfo)
	}
	if !strings.Contains(authInfo.Config, `"IsArchive":"1"`) {
		t.Fatalf("missing archive default: %s", authInfo.Config)
	}
}
