package sender

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"inotify/backend/internal/database"
	"inotify/backend/internal/models"

	"gopkg.in/gomail.v2"
)

type Message struct {
	Title string
	Body  string
	URL   string
	Group string
	Sound string
}

type Service struct {
	store        *database.Store
	client       *http.Client
	telegramBase string
	weixinBase   string
}

type Template struct {
	ID     string
	Name   string
	Order  int
	Inputs []Input `json:"inputs"`
}

type Input struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Placeholder string `json:"placeholder"`
}

func New(store *database.Store) *Service {
	return &Service{
		store:        store,
		client:       &http.Client{Timeout: 15 * time.Second},
		telegramBase: "https://api.telegram.org",
		weixinBase:   "https://qyapi.weixin.qq.com",
	}
}

func (s *Service) SetHTTPClient(client *http.Client) {
	if client != nil {
		s.client = client
	}
}

func (s *Service) SetProviderBases(telegramBase, weixinBase string) {
	if telegramBase != "" {
		s.telegramBase = strings.TrimRight(telegramBase, "/")
	}
	if weixinBase != "" {
		s.weixinBase = strings.TrimRight(weixinBase, "/")
	}
}

func (s *Service) Templates() []map[string]interface{} {
	items := []map[string]interface{}{}
	for _, t := range templates() {
		items = append(items, map[string]interface{}{
			"key":    t.ID,
			"name":   t.Name,
			"order":  t.Order,
			"inputs": t.Inputs,
		})
	}
	return items
}

func (s *Service) SendAuthTemplates() map[string]interface{} {
	out := map[string]interface{}{}
	for _, t := range templates() {
		out[t.ID] = t
	}
	return out
}

func (s *Service) Send(token, key, title, body, link, group, sound string) bool {
	var auths []models.SendAuthInfo
	userID := 0
	query := s.store.DB.Where("active = ?", true)
	if key != "" {
		query = query.Where("key = ?", key)
	} else {
		var user models.SendUserInfo
		if err := s.store.DB.First(&user, "token = ?", token).Error; err != nil {
			s.recordHistory(0, key, Message{Title: title, Body: body, URL: link, Group: group, Sound: sound}, false, 0)
			return false
		}
		userID = user.ID
		query = query.Where("userId = ?", user.ID)
	}
	if err := query.Find(&auths).Error; err != nil || len(auths) == 0 {
		s.recordHistory(userID, key, Message{Title: title, Body: body, URL: link, Group: group, Sound: sound}, false, 0)
		return false
	}
	if userID == 0 {
		userID = auths[0].UserID
	}
	msg := Message{Title: title, Body: body, URL: link, Group: group, Sound: sound}
	ok := false
	count := 0
	for _, authInfo := range auths {
		if s.sendOne(authInfo, msg) {
			ok = true
			count++
			s.increment(authInfo.TemplateID)
		}
	}
	s.recordHistory(userID, key, msg, ok, count)
	return ok
}

func (s *Service) recordHistory(userID int, key string, msg Message, success bool, channelCount int) {
	if s == nil || s.store == nil || s.store.DB == nil || userID == 0 {
		return
	}
	_ = s.store.DB.Create(&models.MessageHistory{
		UserID:       userID,
		Title:        msg.Title,
		Body:         msg.Body,
		URL:          msg.URL,
		Group:        msg.Group,
		Sound:        msg.Sound,
		SendKey:      key,
		Success:      success,
		ChannelCount: channelCount,
		CreateTime:   time.Now(),
	}).Error
}

func (s *Service) sendOne(authInfo models.SendAuthInfo, msg Message) bool {
	cfg, err := decodeConfig(authInfo.Config)
	if err != nil {
		return false
	}
	switch strings.ToLower(authInfo.TemplateID) {
	case strings.ToLower("ADB11045-F2C8-457E-BF7E-1698AD37ED53"), strings.ToLower("HTTP-GET"):
		return s.sendHTTPGet(cfg, msg)
	case strings.ToLower("A3C1E614-717E-4CF1-BA9B-7242717FC037"), strings.ToLower("HTTP-POST"):
		return s.sendHTTPPost(cfg, msg)
	case strings.ToLower("E9669473-FF0B-4474-92BB-E939D92045BB"):
		return s.sendTelegram(cfg, msg)
	case strings.ToLower("048297D4-D975-48F6-9A91-8B4EF75805C1"), strings.ToLower("DINGTALK"):
		return s.sendDingtalk(cfg, msg)
	case strings.ToLower("C01A08B4-3A71-452B-9D4B-D8EC7EF1D68F"), strings.ToLower("FEISHU"):
		return s.sendFeishu(cfg, msg)
	case strings.ToLower("409A30D5-ABE8-4A28-BADD-D04B9908D763"), strings.ToLower("WEIXIN"):
		return s.sendWeixin(cfg, msg)
	case strings.ToLower("EA2B43F7-956C-4C01-B583-0C943ABB36C3"), strings.ToLower("EMAIL"):
		return s.sendEmail(cfg, msg)
	case strings.ToLower("3B6DE04D-A9EF-4C91-A151-60B7425C5AB2"), strings.ToLower("BARK"):
		return s.sendBark(cfg, msg)
	default:
		return s.sendByKnownFields(cfg, msg)
	}
}

func (s *Service) sendHTTPGet(cfg map[string]string, msg Message) bool {
	raw := first(cfg, "URL", "Url", "url")
	if raw == "" {
		return false
	}
	resp, err := s.client.Get(applyTemplate(raw, msg))
	return closeOK(resp, err)
}

func (s *Service) sendHTTPPost(cfg map[string]string, msg Message) bool {
	raw := first(cfg, "URL", "Url", "url", "WebHook")
	if raw == "" {
		return false
	}
	contentType := first(cfg, "ContentType")
	if contentType == "" {
		contentType = "application/json"
	}
	data := applyTemplate(first(cfg, "Data", "data"), msg)
	if data == "" {
		data = fmt.Sprintf(`{"title":%q,"data":%q,"body":%q,"url":%q}`, msg.Title, msg.Body, msg.Body, msg.URL)
	}
	resp, err := s.client.Post(applyTemplate(raw, msg), contentType, strings.NewReader(data))
	return closeOK(resp, err)
}

func (s *Service) sendTelegram(cfg map[string]string, msg Message) bool {
	token := first(cfg, "BotToken", "botToken")
	chatID := first(cfg, "ChatId", "Chat_id", "chat_id")
	if token == "" || chatID == "" {
		return false
	}
	api := fmt.Sprintf("%s/bot%s/sendMessage", s.telegramBase, token)
	return s.postForm(api, url.Values{"chat_id": {chatID}, "text": {msg.Title + "\n" + msg.Body}})
}

func (s *Service) sendDingtalk(cfg map[string]string, msg Message) bool {
	webhook := first(cfg, "WebHook", "Webhook", "url")
	if webhook == "" {
		return false
	}
	if secret := first(cfg, "Secret"); secret != "" {
		timestamp := time.Now().UTC().UnixMilli()
		sign := dingtalkSign(timestamp, secret)
		sep := "&"
		if !strings.Contains(webhook, "?") {
			sep = "?"
		}
		webhook = fmt.Sprintf("%s%stimestamp=%d&sign=%s", webhook, sep, timestamp, url.QueryEscape(sign))
	}
	payload := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": msg.Title + "\n" + msg.Body},
	}
	return s.sendWebhook(webhook, payload)
}

func (s *Service) sendFeishu(cfg map[string]string, msg Message) bool {
	webhook := first(cfg, "WebHook", "Webhook", "url")
	if webhook == "" {
		return false
	}
	payload := map[string]interface{}{
		"msg_type": "text",
		"content":  map[string]string{"text": msg.Title + "\n" + msg.Body},
	}
	if secret := first(cfg, "Secret"); secret != "" {
		timestamp := time.Now().Unix() - 10
		payload["timestamp"] = strconv.FormatInt(timestamp, 10)
		payload["sign"] = feishuSign(timestamp, secret)
	}
	return s.sendWebhook(webhook, payload)
}

func (s *Service) sendWeixin(cfg map[string]string, msg Message) bool {
	webhook := first(cfg, "WebHook", "Webhook", "url")
	if webhook != "" {
		return s.sendWebhook(webhook, map[string]interface{}{"msgtype": "text", "text": map[string]string{"content": msg.Title + "\n" + msg.Body}})
	}
	corpID := first(cfg, "Corpid", "CorpId", "corpId")
	secret := first(cfg, "Corpsecret", "CorpSecret", "Secret")
	agentID := first(cfg, "AgentID", "AgentId", "agentId")
	toUser := first(cfg, "OpengId", "ToUser", "touser")
	if toUser == "" {
		toUser = "@all"
	}
	if corpID == "" || secret == "" || agentID == "" {
		return false
	}
	token, ok := s.weixinToken(corpID, secret)
	if !ok {
		return false
	}
	payload := map[string]interface{}{
		"touser":  toUser,
		"msgtype": "text",
		"agentid": agentID,
		"text":    map[string]string{"content": msg.Title + "\n" + msg.Body},
		"safe":    0,
	}
	api := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", s.weixinBase, url.QueryEscape(token))
	return s.postJSON(api, payload)
}

func (s *Service) sendEmail(cfg map[string]string, msg Message) bool {
	host := first(cfg, "Host", "SmtpHost")
	user := first(cfg, "From", "User", "UserName", "Username")
	pass := first(cfg, "Password")
	to := first(cfg, "To", "Email")
	if host == "" || user == "" || to == "" {
		return false
	}
	port := 25
	fmt.Sscanf(first(cfg, "Port"), "%d", &port)
	m := gomail.NewMessage()
	fromName := first(cfg, "FromName")
	if fromName != "" {
		m.SetAddressHeader("From", user, fromName)
	} else {
		m.SetHeader("From", user)
	}
	m.SetHeader("To", to)
	m.SetHeader("Subject", msg.Title)
	m.SetBody("text/plain", msg.Body)
	return gomail.NewDialer(host, port, user, pass).DialAndSend(m) == nil
}

func (s *Service) sendBark(cfg map[string]string, msg Message) bool {
	if sendURL := first(cfg, "SendUrl", "URL", "Url", "url"); sendURL != "" {
		resp, err := s.client.Get(applyTemplate(sendURL, msg))
		return closeOK(resp, err)
	}
	deviceKey := first(cfg, "DeviceKey", "DeviceToken", "deviceToken", "token")
	if deviceKey == "" {
		return false
	}
	api := "https://api.day.app/" + url.PathEscape(deviceKey) + "/" + url.PathEscape(msg.Title) + "/" + url.PathEscape(msg.Body)
	q := url.Values{}
	if msg.URL != "" {
		q.Set("url", msg.URL)
	}
	if sound := first(cfg, "Sound"); sound != "" {
		q.Set("sound", sound)
	}
	if archive := first(cfg, "IsArchive"); archive != "" {
		q.Set("isArchive", archive)
	}
	if copyValue := first(cfg, "AutoMaticallyCopy", "AutomaticallyCopy"); copyValue != "" {
		q.Set("automaticallyCopy", copyValue)
	}
	if encoded := q.Encode(); encoded != "" {
		api += "?" + encoded
	}
	resp, err := s.client.Get(api)
	return closeOK(resp, err)
}

func (s *Service) sendByKnownFields(cfg map[string]string, msg Message) bool {
	if webhook := first(cfg, "WebHook", "Webhook", "Url", "URL", "url"); webhook != "" {
		return s.sendWebhook(webhook, map[string]interface{}{"title": msg.Title, "body": msg.Body, "url": msg.URL})
	}
	return false
}

func (s *Service) sendWebhook(webhook string, payload map[string]interface{}) bool {
	if webhook == "" {
		return false
	}
	return s.postJSON(webhook, payload)
}

func (s *Service) weixinToken(corpID, secret string) (string, bool) {
	api := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", s.weixinBase, url.QueryEscape(corpID), url.QueryEscape(secret))
	resp, err := s.client.Get(api)
	if err != nil || resp == nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.Copy(io.Discard, resp.Body)
		return "", false
	}
	var body struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", false
	}
	return body.AccessToken, body.AccessToken != "" && body.ErrCode == 0
}

func (s *Service) postJSON(raw string, payload interface{}) bool {
	data, _ := json.Marshal(payload)
	resp, err := s.client.Post(raw, "application/json", bytes.NewReader(data))
	return closeOK(resp, err)
}

func (s *Service) postForm(raw string, values url.Values) bool {
	resp, err := s.client.PostForm(raw, values)
	return closeOK(resp, err)
}

func (s *Service) increment(templateID string) {
	date := time.Now().Format("2006-01-02")
	var item models.SendInfo
	if err := s.store.DB.First(&item, "templateID = ? AND date = ?", templateID, date).Error; err != nil {
		s.store.DB.Create(&models.SendInfo{TemplateID: templateID, Date: date, Count: 1})
		return
	}
	item.Count++
	s.store.DB.Save(&item)
}

func closeOK(resp *http.Response, err error) bool {
	if err != nil || resp == nil {
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func first(cfg map[string]string, keys ...string) string {
	for _, key := range keys {
		if v := cfg[key]; v != "" {
			return v
		}
	}
	return ""
}

func decodeConfig(raw string) (map[string]string, error) {
	var values map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for key, value := range values {
		switch v := value.(type) {
		case string:
			out[key] = v
		case float64:
			out[key] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			out[key] = strconv.FormatBool(v)
		case nil:
			out[key] = ""
		default:
			data, _ := json.Marshal(v)
			out[key] = string(data)
		}
	}
	return out, nil
}

func templates() []Template {
	return []Template{
		{ID: "409A30D5-ABE8-4A28-BADD-D04B9908D763", Name: "企业微信", Order: 0, Inputs: []Input{{0, "Corpid", "Corpid", "企业ID"}, {1, "Corpsecret", "Corpsecret", "密钥"}, {2, "AgentID", "AgentID", "应用ID"}, {3, "OpengId", "OpengId", "@all"}}},
		{ID: "EA2B43F7-956C-4C01-B583-0C943ABB36C3", Name: "邮件推送", Order: 1, Inputs: []Input{{0, "FromName", "FromName", "管理员"}, {1, "From", "From", "abc@qq.com"}, {2, "Password", "Password", "123456"}, {3, "Host", "Host", "smtp.qq.com"}, {4, "Port", "Port", "587"}, {5, "EnableSSL", "EnableSSL", "true|false"}, {6, "To", "To", "abcd@qq.com"}}},
		{ID: "E9669473-FF0B-4474-92BB-E939D92045BB", Name: "电报机器人", Order: 2, Inputs: []Input{{0, "BotToken", "BotToken", "ID:Token"}, {1, "Chat_id", "ChatId", "ChatId"}}},
		{ID: "ADB11045-F2C8-457E-BF7E-1698AD37ED53", Name: "自定义GET", Order: 4, Inputs: []Input{{0, "URL", "URL", "https://api.day.app/token/{title}/{data}"}}},
		{ID: "A3C1E614-717E-4CF1-BA9B-7242717FC037", Name: "自定义POST", Order: 5, Inputs: []Input{{0, "URL", "URL", "https://api.day.app/token/{title}/{data}"}, {1, "Encoding", "Encoding", "utf-8"}, {1, "ContentType", "ContentType", "application/json"}, {2, "Data", "Data", `{"msgid":"123456","title":"{title}","data":"{data}"}`}}},
		{ID: "048297D4-D975-48F6-9A91-8B4EF75805C1", Name: "钉钉群机器人", Order: 21, Inputs: []Input{{0, "WebHook", "WebHook", "https://oapi.dingtalk.com/robot/send?access_token=xxxxx"}, {0, "Secret", "Secret", "SEC77xxxx"}}},
		{ID: "C01A08B4-3A71-452B-9D4B-D8EC7EF1D68F", Name: "飞书群机器人", Order: 22, Inputs: []Input{{0, "WebHook", "WebHook", "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxx"}, {0, "Secret", "Secret", "VcgAbeuZOhTZPSP0zxxxx"}}},
		{ID: "3B6DE04D-A9EF-4C91-A151-60B7425C5AB2", Name: "Bark", Order: 2999, Inputs: []Input{{1, "Sound", "Sound", "1107"}, {2, "IsArchive", "IsArchive", "1或0"}, {3, "AutoMaticallyCopy", "AutoMaticallyCopy", "1或0"}, {4, "DeviceKey", "DeviceKey", "DeviceKey"}, {5, "DeviceToken", "DeviceToken", "DeviceToken"}, {6, "SendUrl", "SendUrl", "SendUrl"}}},
	}
}

func applyTemplate(raw string, msg Message) string {
	out := strings.ReplaceAll(raw, "{title}", msg.Title)
	out = strings.ReplaceAll(out, "{data}", msg.Body)
	out = strings.ReplaceAll(out, "{body}", msg.Body)
	out = strings.ReplaceAll(out, "{url}", msg.URL)
	return out
}

func dingtalkSign(timestamp int64, secret string) string {
	return hmacSHA256Base64(fmt.Sprintf("%d\n%s", timestamp, secret), secret)
}

func feishuSign(timestamp int64, secret string) string {
	return hmacSHA256Base64("", fmt.Sprintf("%d\n%s", timestamp, secret))
}

func hmacSHA256Base64(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
