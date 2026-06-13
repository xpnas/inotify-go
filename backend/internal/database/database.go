package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"inotify/backend/internal/auth"
	"inotify/backend/internal/config"
	"inotify/backend/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Store struct {
	DB      *gorm.DB
	Config  config.Config
	JWTInfo models.JwtInfo
}

func Open(cfg config.Config) (*Store, error) {
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	store := &Store{DB: db, Config: cfg}
	if err := store.loadJWT(); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.SystemInfo{}, &models.SendInfo{}, &models.SendUserInfo{}, &models.SendAuthInfo{}, &models.MessageHistory{}); err != nil {
		return nil, err
	}
	if err := store.seed(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) loadJWT() error {
	if _, err := os.Stat(s.Config.JWTPath); errors.Is(err, os.ErrNotExist) {
		s.JWTInfo = models.JwtInfo{
			Issuer:             "Inotify",
			Audience:           "Inotify",
			IssuerSigningKey:   randomHex(32),
			AccessTokenExpires: 60 * 24 * 30,
		}
		return s.SaveJWT(s.JWTInfo)
	}
	data, err := os.ReadFile(s.Config.JWTPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &s.JWTInfo); err != nil {
		return err
	}
	if s.JWTInfo.IssuerSigningKey == "" {
		s.JWTInfo.IssuerSigningKey = randomHex(32)
	}
	return nil
}

func (s *Store) SaveJWT(info models.JwtInfo) error {
	if info.Issuer == "" {
		info.Issuer = "Inotify"
	}
	if info.Audience == "" {
		info.Audience = "Inotify"
	}
	if info.IssuerSigningKey == "" {
		info.IssuerSigningKey = randomHex(32)
	}
	if info.AccessTokenExpires <= 0 {
		info.AccessTokenExpires = 60 * 24 * 30
	}
	s.JWTInfo = info
	if err := os.MkdirAll(filepath.Dir(s.Config.JWTPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Config.JWTPath, data, 0600)
}

func (s *Store) seed() error {
	var count int64
	if err := s.DB.Model(&models.SendUserInfo{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		admin := models.SendUserInfo{
			SystemUserInfo: models.SystemUserInfo{
				UserName:   "admin",
				Password:   auth.MD5Hex("123456"),
				Avatar:     "",
				Email:      "",
				Active:     true,
				CreateTime: time.Now(),
			},
			Token: randomHex(16),
		}
		if err := s.DB.Create(&admin).Error; err != nil {
			return err
		}
	}
	defaults := map[string]string{
		"githubClientId":     "",
		"githubClientSecret": "",
		"proxyAddress":       "",
		"administrators":     "admin",
		"adminUserName":      "admin",
	}
	for key, value := range defaults {
		var item models.SystemInfo
		err := s.DB.First(&item, "key = ?", key).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.DB.Create(&models.SystemInfo{Key: key, Value: value}).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetSystemValue(key string) string {
	var item models.SystemInfo
	if err := s.DB.First(&item, "key = ?", key).Error; err != nil {
		return ""
	}
	return item.Value
}

func (s *Store) SetSystemValue(key, value string) error {
	return s.DB.Save(&models.SystemInfo{Key: key, Value: value}).Error
}

func (s *Store) GetUser(username string) (*models.SendUserInfo, error) {
	var user models.SendUserInfo
	if err := s.DB.First(&user, "userName = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) Role(username string) string {
	if username == "admin" || containsCSV(s.GetSystemValue("administrators"), username) || containsCSV(s.GetSystemValue("adminUserName"), username) {
		return auth.RoleSystem
	}
	return auth.RoleUser
}

func containsCSV(csv, value string) bool {
	for _, item := range strings.Split(csv, ",") {
		if strings.TrimSpace(item) == value {
			return true
		}
	}
	return false
}

func randomHex(bytes int) string {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return stringsUpper(hex.EncodeToString(buf))
}

func stringsUpper(s string) string {
	out := []byte(s)
	for i, c := range out {
		if c >= 'a' && c <= 'f' {
			out[i] = c - 32
		}
	}
	return string(out)
}
