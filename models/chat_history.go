package models

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// ChatHistory 聊天历史记录模型
type ChatHistory struct {
	gorm.Model
	HistoryID string `gorm:"size:255;not null;unique" json:"history_id"` // 历史记录唯一标识
	UserID    uint   `gorm:"not null" json:"user_id"`                   // 用户ID，外键关联到User表
	ModelName     string `gorm:"size:255;not null" json:"model"`            // 使用的模型名称
	Messages  string `gorm:"type:text;not null" json:"messages"`         // 聊天消息内容，JSON格式存储
	User      User   `gorm:"foreignKey:UserID" json:"-"`                // 关联的用户
}

// SetMessages 将消息数组转换为JSON字符串并保存
func (ch *ChatHistory) SetMessages(messages []interface{}) error {
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	ch.Messages = string(messagesJSON)
	return nil
}

// GetMessages 将JSON字符串转换为消息数组
func (ch *ChatHistory) GetMessages() ([]interface{}, error) {
	var messages []interface{}
	err := json.Unmarshal([]byte(ch.Messages), &messages)
	return messages, err
}

// CreateChatHistory 创建新的聊天历史记录
func CreateChatHistory(history *ChatHistory) (*ChatHistory, error) {
	result := DB.Create(history)
	if result.Error != nil {
		return nil, result.Error
	}
	return history, nil
}

// GetChatHistoryByID 通过ID获取聊天历史记录
func GetChatHistoryByID(id uint) (*ChatHistory, error) {
	var history ChatHistory
	result := DB.First(&history, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("聊天历史记录不存在")
		}
		return nil, result.Error
	}
	return &history, nil
}

// GetChatHistoriesByUserID 获取用户的所有聊天历史记录
func GetChatHistoriesByUserID(userID uint) ([]ChatHistory, error) {
	var histories []ChatHistory
	result := DB.Where("user_id = ?", userID).Order("created_at desc").Find(&histories)
	if result.Error != nil {
		return nil, result.Error
	}
	return histories, nil
}

// GetChatHistoryByHistoryID 通过历史记录ID获取聊天历史
func GetChatHistoryByHistoryID(historyID string) (*ChatHistory, error) {
	var history ChatHistory
	result := DB.Where("history_id = ?", historyID).First(&history)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("聊天历史记录不存在")
		}
		return nil, result.Error
	}
	return &history, nil
}

// DeleteChatHistory 删除聊天历史记录
func DeleteChatHistory(id uint) error {
	result := DB.Delete(&ChatHistory{}, id)
	return result.Error
}