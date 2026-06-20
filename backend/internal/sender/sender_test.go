package sender

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"inotify/backend/internal/models"
)

func TestTemplatesUseOriginalIDs(t *testing.T) {
	want := map[string]string{
		"409A30D5-ABE8-4A28-BADD-D04B9908D763": "企业微信",
		"EA2B43F7-956C-4C01-B583-0C943ABB36C3": "邮件推送",
		"E9669473-FF0B-4474-92BB-E939D92045BB": "电报机器人",
		"ADB11045-F2C8-457E-BF7E-1698AD37ED53": "自定义GET",
		"A3C1E614-717E-4CF1-BA9B-7242717FC037": "自定义POST",
		"048297D4-D975-48F6-9A91-8B4EF75805C1": "钉钉群机器人",
		"C01A08B4-3A71-452B-9D4B-D8EC7EF1D68F": "飞书群机器人",
		"3B6DE04D-A9EF-4C91-A151-60B7425C5AB2": "Bark",
	}
	got := map[string]string{}
	for _, tpl := range templates() {
		got[tpl.ID] = tpl.Name
	}
	for id, name := range want {
		if got[id] != name {
			t.Fatalf("template %s = %q, want %q", id, got[id], name)
		}
	}
}

func TestHTTPTemplatesReplaceTitleAndData(t *testing.T) {
	var gotPath, gotBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := New(nil)
	auth := models.SendAuthInfo{
		TemplateID: "A3C1E614-717E-4CF1-BA9B-7242717FC037",
		Config:     `{"URL":"` + ts.URL + `/{title}/{data}","ContentType":"application/json","Data":"{\"title\":\"{title}\",\"data\":\"{data}\"}"}`,
	}
	ok := s.sendOne(auth, Message{Title: "hello", Body: "world"})
	if !ok {
		t.Fatal("sendOne returned false")
	}
	if gotPath != "/hello/world" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody != `{"title":"hello","data":"world"}` {
		t.Fatalf("body = %q", gotBody)
	}
}

func TestDingtalkAddsSignature(t *testing.T) {
	var rawQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := New(nil)
	auth := models.SendAuthInfo{
		TemplateID: "048297D4-D975-48F6-9A91-8B4EF75805C1",
		Config:     `{"WebHook":"` + ts.URL + `/robot/send?access_token=abc","Secret":"SECxxx"}`,
	}
	if !s.sendOne(auth, Message{Title: "t", Body: "b"}) {
		t.Fatal("dingtalk send failed")
	}
	if !strings.Contains(rawQuery, "access_token=abc") || !strings.Contains(rawQuery, "timestamp=") || !strings.Contains(rawQuery, "sign=") {
		t.Fatalf("missing dingtalk signature query: %s", rawQuery)
	}
}

func TestFeishuAddsSignaturePayload(t *testing.T) {
	var payload map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := New(nil)
	auth := models.SendAuthInfo{
		TemplateID: "C01A08B4-3A71-452B-9D4B-D8EC7EF1D68F",
		Config:     `{"WebHook":"` + ts.URL + `","Secret":"secret"}`,
	}
	if !s.sendOne(auth, Message{Title: "t", Body: "b"}) {
		t.Fatal("feishu send failed")
	}
	if payload["timestamp"] == "" || payload["sign"] == "" || payload["msg_type"] != "text" {
		t.Fatalf("bad feishu payload: %#v", payload)
	}
}

func TestWeixinTokenThenSend(t *testing.T) {
	var sentPath string
	var sentPayload map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/gettoken":
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"token123"}`))
		case "/cgi-bin/message/send":
			sentPath = r.URL.String()
			_ = json.NewDecoder(r.Body).Decode(&sentPayload)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	s := New(nil)
	s.SetProviderBases("", ts.URL)
	auth := models.SendAuthInfo{
		TemplateID: "409A30D5-ABE8-4A28-BADD-D04B9908D763",
		Config:     `{"Corpid":"corp","Corpsecret":"secret","AgentID":"100","OpengId":"@all"}`,
	}
	if !s.sendOne(auth, Message{Title: "title", Body: "body"}) {
		t.Fatal("weixin send failed")
	}
	if !strings.Contains(sentPath, "access_token=token123") {
		t.Fatalf("send path = %q", sentPath)
	}
	if sentPayload["touser"] != "@all" || sentPayload["msgtype"] != "text" {
		t.Fatalf("bad weixin payload: %#v", sentPayload)
	}
}

func TestWeixinUploadsImageThenSendsMedia(t *testing.T) {
	var sentPayloads []map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dynamic":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte{0xff, 0xd8, 0xff, 0xdb, 0, 0, 0, 0, 0xff, 0xd9})
		case "/cgi-bin/gettoken":
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"token123"}`))
		case "/cgi-bin/media/upload":
			if r.URL.Query().Get("type") != "image" {
				t.Fatalf("upload type = %q", r.URL.RawQuery)
			}
			if err := r.ParseMultipartForm(3 << 20); err != nil {
				t.Fatalf("multipart: %v", err)
			}
			if _, _, err := r.FormFile("media"); err != nil {
				t.Fatalf("media file: %v", err)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"media_id":"media123"}`))
		case "/cgi-bin/message/send":
			var payload map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			sentPayloads = append(sentPayloads, payload)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	s := New(nil)
	s.SetProviderBases("", ts.URL)
	auth := models.SendAuthInfo{
		TemplateID: "409A30D5-ABE8-4A28-BADD-D04B9908D763",
		Config:     `{"Corpid":"corp","Corpsecret":"secret","AgentID":"100","OpengId":"@all","ImageMode":"upload"}`,
	}
	if !s.sendOne(auth, Message{Title: "title", Body: "body", URL: ts.URL + "/dynamic?f=JPEG&w=1024"}) {
		t.Fatal("weixin image send failed")
	}
	if len(sentPayloads) != 2 {
		t.Fatalf("sent payload count = %d", len(sentPayloads))
	}
	if sentPayloads[0]["msgtype"] != "text" || sentPayloads[1]["msgtype"] != "image" {
		t.Fatalf("bad payloads: %#v", sentPayloads)
	}
	image, _ := sentPayloads[1]["image"].(map[string]interface{})
	if image["media_id"] != "media123" {
		t.Fatalf("bad image payload: %#v", sentPayloads[1])
	}
}

func TestDecodeConfigAcceptsNumbersAndBools(t *testing.T) {
	cfg, err := decodeConfig(`{"Port":587,"EnableSSL":true,"Host":"smtp.example.com"}`)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["Port"] != "587" || cfg["EnableSSL"] != "true" || cfg["Host"] != "smtp.example.com" {
		t.Fatalf("bad config: %#v", cfg)
	}
}
