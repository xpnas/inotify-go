package models

import "time"

type SystemInfo struct {
	Key   string `gorm:"column:key;primaryKey" json:"key"`
	Value string `gorm:"column:value" json:"value"`
}

func (SystemInfo) TableName() string { return "systemInfo" }

type JwtInfo struct {
	Issuer             string `json:"issuer"`
	Audience           string `json:"audience"`
	IssuerSigningKey   string `json:"issuerSigningKey"`
	AccessTokenExpires int    `json:"accessTokenExpires"`
}

type SystemUserInfo struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserName   string    `gorm:"column:userName;uniqueIndex" json:"userName"`
	Password   string    `gorm:"column:password" json:"password,omitempty"`
	Avatar     string    `gorm:"column:avatar" json:"avatar"`
	Email      string    `gorm:"column:email" json:"email"`
	Active     bool      `gorm:"column:active" json:"active"`
	CreateTime time.Time `gorm:"column:createTime" json:"createTime"`
}

func (SystemUserInfo) TableName() string { return "systemUser" }

type SendUserInfo struct {
	SystemUserInfo
	Token      string `gorm:"column:token;uniqueIndex" json:"token"`
	SendAuthID int    `gorm:"column:sendAuthId" json:"sendAuthId"`
}

func (SendUserInfo) TableName() string { return "userInfo" }

type SendAuthInfo struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     int       `gorm:"column:userId;index" json:"userId"`
	TemplateID string    `gorm:"column:templateID;index" json:"templateID"`
	Key        string    `gorm:"column:key;uniqueIndex" json:"key"`
	Name       string    `gorm:"column:name" json:"name"`
	Config     string    `gorm:"column:config" json:"config"`
	Active     bool      `gorm:"column:active" json:"active"`
	CreateTime time.Time `gorm:"column:createTime" json:"createTime"`
}

func (SendAuthInfo) TableName() string { return "sendAuthInfo" }

type SendInfo struct {
	TemplateID string `gorm:"column:templateID;primaryKey" json:"templateID"`
	Date       string `gorm:"column:date;primaryKey" json:"date"`
	Count      int    `gorm:"column:count" json:"count"`
}

func (SendInfo) TableName() string { return "sendInfo" }

type MessageHistory struct {
	ID           int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       int       `gorm:"column:userId;index" json:"userId"`
	Title        string    `gorm:"column:title;index" json:"title"`
	Body         string    `gorm:"column:body" json:"body"`
	URL          string    `gorm:"column:url" json:"url"`
	Group        string    `gorm:"column:groupName" json:"group"`
	Sound        string    `gorm:"column:sound" json:"sound"`
	SendKey      string    `gorm:"column:sendKey;index" json:"sendKey"`
	Success      bool      `gorm:"column:success;index" json:"success"`
	ChannelCount int       `gorm:"column:channelCount" json:"channelCount"`
	CreateTime   time.Time `gorm:"column:createTime;index" json:"createTime"`
}

func (MessageHistory) TableName() string { return "messageHistory" }

type APIResult struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"msg,omitempty"`
}
