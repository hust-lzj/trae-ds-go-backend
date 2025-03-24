package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trae-ds-go-backend/config"
	"github.com/trae-ds-go-backend/models"
)

// SaveChatHistoryInput 保存聊天历史的请求结构
type SaveChatHistoryInput struct {
	Model    string           `json:"model" binding:"required"`
	Messages []config.Message `json:"messages" binding:"required"`
}

// SaveChatHistory 保存聊天历史记录
func SaveChatHistory(c *gin.Context) {
	// 获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 绑定请求数据
	var input SaveChatHistoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 创建聊天历史记录
	history := models.ChatHistory{
		HistoryID: uuid.New().String(), // 生成唯一的历史记录ID
		UserID:    userID,
		ModelName: input.Model,
	}

	// 将消息转换为JSON字符串
	messagesJSON, err := json.Marshal(input.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "消息序列化失败"})
		return
	}
	history.Messages = string(messagesJSON)

	// 保存到数据库
	result := models.DB.Create(&history)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存聊天历史失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "聊天历史保存成功",
		"history_id": history.HistoryID,
	})
}

// GetUserChatHistories 获取用户的聊天历史记录列表
func GetUserChatHistories(c *gin.Context) {
	// 获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 获取历史记录
	histories, err := models.GetChatHistoriesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取聊天历史失败: %v", err)})
		return
	}

	// 构建响应数据
	var responseHistories []gin.H
	for _, history := range histories {
		// 解析消息内容
		var messages []config.Message
		err := json.Unmarshal([]byte(history.Messages), &messages)
		if err != nil {
			continue // 跳过无法解析的记录
		}

		// 提取第一条消息作为标题（如果存在）
		title := "新对话"
		if len(messages) > 0 {
			// 截取内容的前30个字符作为标题
			content := messages[0].Content
			if len(content) > 30 {
				title = content[:30] + "..."
			} else {
				title = content
			}
		}

		responseHistories = append(responseHistories, gin.H{
			"id":         history.ID,
			"history_id": history.HistoryID,
			"model":      history.ModelName,
			"title":      title,
			"created_at": history.CreatedAt,
			"updated_at": history.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"histories": responseHistories})
}

// GetChatHistoryDetail 获取聊天历史详情
func GetChatHistoryDetail(c *gin.Context) {
	// 获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 获取历史记录ID
	historyID := c.Param("history_id")
	if historyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "历史记录ID不能为空"})
		return
	}

	// 查询历史记录
	history, err := models.GetChatHistoryByHistoryID(historyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "聊天历史记录不存在"})
		return
	}

	// 验证是否属于当前用户
	if history.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该聊天历史"})
		return
	}

	// 解析消息内容
	var messages []config.Message
	err = json.Unmarshal([]byte(history.Messages), &messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析聊天消息失败"})
		return
	}

	// 返回历史记录详情
	c.JSON(http.StatusOK, gin.H{
		"history_id": history.HistoryID,
		"model":      history.ModelName,
		"messages":   messages,
		"created_at": history.CreatedAt,
		"updated_at": history.UpdatedAt,
	})
}

// DeleteChatHistory 删除聊天历史记录
func DeleteChatHistory(c *gin.Context) {
	// 获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 获取历史记录ID
	historyIDStr := c.Param("id")
	historyID, err := strconv.ParseUint(historyIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的历史记录ID"})
		return
	}

	// 查询历史记录
	history, err := models.GetChatHistoryByID(uint(historyID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "聊天历史记录不存在"})
		return
	}

	// 验证是否属于当前用户
	if history.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除该聊天历史"})
		return
	}

	// 删除历史记录
	err = models.DeleteChatHistory(uint(historyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除聊天历史失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "聊天历史删除成功"})
}