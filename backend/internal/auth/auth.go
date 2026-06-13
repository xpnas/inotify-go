package auth

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"inotify/backend/internal/models"
)

const (
	RoleSystem = "System"
	RoleUser   = "User"
)

func MD5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func GenerateToken(jwtInfo models.JwtInfo, username, role string) (string, error) {
	if jwtInfo.AccessTokenExpires <= 0 {
		jwtInfo.AccessTokenExpires = 60 * 24 * 30
	}
	claims := jwt.MapClaims{
		"name": username,
		"role": role,
		"iss":  jwtInfo.Issuer,
		"aud":  jwtInfo.Audience,
		"exp":  time.Now().Add(time.Duration(jwtInfo.AccessTokenExpires) * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtInfo.IssuerSigningKey))
}

func ParseToken(jwtInfo models.JwtInfo, tokenString string) (jwt.MapClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtInfo.IssuerSigningKey), nil
	}, jwt.WithAudience(jwtInfo.Audience), jwt.WithIssuer(jwtInfo.Issuer))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func CurrentUser(c *gin.Context) string {
	if value, ok := c.Get("userName"); ok {
		if s, ok := value.(string); ok {
			return s
		}
	}
	return ""
}

func CurrentRole(c *gin.Context) string {
	if value, ok := c.Get("role"); ok {
		if s, ok := value.(string); ok {
			return s
		}
	}
	return ""
}
