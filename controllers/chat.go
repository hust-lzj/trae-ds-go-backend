package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trae-ds-go-backend/config"
	"github.com/trae-ds-go-backend/models"
)

// ChatInput 聊天请求结构
type ChatInput struct {
	Messages  []config.Message       `json:"messages" binding:"required"`
	Model     string                 `json:"model" binding:"required"`
	Options   map[string]interface{} `json:"options"`
	HistoryID string                 `json:"history_id"` // 聊天历史ID，可选参数
}

// StreamChat 处理流式聊天请求
func StreamChat(c *gin.Context) {
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
	var input ChatInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 打印请求信息
	fmt.Println("发送流式聊天请求，模型:", input.Model)
	fmt.Println("消息数量:", len(input.Messages))
    fmt.Println("历史记录ID:", input.HistoryID)

	// 创建LLM客户端
	client := config.NewLLMClient()

	// 设置响应头，通知前端这是一个流式响应
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	
	// 创建一个响应收集器，用于收集AI的响应内容
	var aiResponseContent string
	
	// 先创建ResponseCollector实例
	responseCollector := &ResponseCollector{
		Writer: c.Writer,
		UserID: userID,
		ModelName: input.Model,
		UserMessages: input.Messages,
		HistoryID: input.HistoryID,
		ResponseContent: "",
	}
	
	// 然后设置CollectContent函数
	responseCollector.CollectContent = func(content string) {
		aiResponseContent += content
		responseCollector.ResponseContent += content
	}

	// 发送流式请求到模型并直接将响应流式传输给客户端
	err := client.StreamChat(responseCollector, input.Messages, input.Options, input.Model)
	if err != nil {
		fmt.Println("流式模型请求失败:", err)
		// 注意：此时可能已经发送了部分响应，无法再发送JSON错误响应
		// 可以考虑发送一个特殊的事件消息表示错误
		c.Writer.Write([]byte(fmt.Sprintf("event: error\ndata: {\"error\": \"模型请求失败: %v\"}\n\n", err)))
		c.Writer.Flush()
		return
	}
	
	// 注意：历史记录的创建已经移到ResponseCollector的Write方法中处理
	// 这里不再需要单独创建历史记录
}

// ResponseCollector 用于同时收集AI响应内容并转发给客户端
type ResponseCollector struct {
	http.ResponseWriter
	Writer         http.ResponseWriter
	CollectContent func(string)
	// 新增字段
	UserID         uint
	ModelName      string
	UserMessages   []config.Message
	HistoryID      string
	LastDoneData   []byte // 存储最后一条done=true的数据
	ResponseContent string // 存储收集到的AI响应内容
}

// Write 实现http.ResponseWriter接口
func (rc *ResponseCollector) Write(data []byte) (int, error) {
	// 尝试解析事件数据以提取内容
	dataStr := string(data)

	// 尝试解析JSON内容
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(dataStr), &jsonData)
	if err == nil {
		// 如果包含content字段，则收集内容
		// 先检查message字段是否存在并且是map类型
		if message, ok := jsonData["message"].(map[string]interface{}); ok {
			// 然后从message map中获取content字段
			if content, ok := message["content"].(string); ok {
				rc.CollectContent(content)
			}
		}
		
		// 检查是否是最后一条消息（done=true）
		if done, ok := jsonData["done"].(bool); ok && done {
			// 如果是最后一条消息，暂存数据，不立即发送
			rc.LastDoneData = make([]byte, len(data))
			copy(rc.LastDoneData, data)
			
			// 如果需要创建或更新历史记录
			if rc.UserID > 0 && rc.LastDoneData != nil {
				// 创建AI响应消息
				aiMessage := config.Message{
					Role:    "assistant",
					Content: rc.ResponseContent,
				}
				
				// 创建或更新聊天历史记录
				var historyID string
				if rc.HistoryID == "" {
					// 创建新的聊天历史记录
					historyID = createChatHistoryFromStream(rc.UserID, rc.ModelName, rc.UserMessages, aiMessage)
					fmt.Println("新的聊天历史记录已创建，ID:", historyID)
				} else {
					// 使用现有的历史记录ID
					historyID = rc.HistoryID
					// 更新现有历史记录
					updateChatHistoryFromStream(rc.HistoryID, rc.UserMessages, aiMessage)
					fmt.Println("聊天历史记录已更新，ID:", historyID)
				}
				
				// 修改原始JSON数据，添加history_id字段
				if historyID != "" {
					jsonData["history_id"] = historyID
					// 重新序列化JSON
					newJSONContent, err := json.Marshal(jsonData)
					if err == nil {
						// 构造新的SSE消息
						newData := []byte(string(newJSONContent) + "\n\n")
						// 发送修改后的数据
						return rc.Writer.Write(newData)
					}
				}
			}
			
			// 如果无法处理历史记录，则发送原始数据
			return rc.Writer.Write(data)
		}
	}
	
	
	// 对于非最后一条消息或解析失败的情况，直接转发数据
	return rc.Writer.Write(data)
}

// Header 实现http.ResponseWriter接口
func (rc *ResponseCollector) Header() http.Header {
	return rc.Writer.Header()
}

// WriteHeader 实现http.ResponseWriter接口
func (rc *ResponseCollector) WriteHeader(statusCode int) {
	rc.Writer.WriteHeader(statusCode)
}

// Flush 实现http.Flusher接口
func (rc *ResponseCollector) Flush() {
	if f, ok := rc.Writer.(http.Flusher); ok {
		f.Flush()
	}
}

// createChatHistoryFromStream 从流式聊天创建新的聊天历史记录
func createChatHistoryFromStream(userID uint, modelName string, userMessages []config.Message, aiMessage config.Message) string {
	// 创建包含用户消息和AI响应的完整消息列表
	messages := append(userMessages, aiMessage)
	
	// 创建聊天历史记录
	history := models.ChatHistory{
		HistoryID: uuid.New().String(), // 生成唯一的历史记录ID
		UserID:    userID,
		ModelName: modelName,
	}
	
	// 将消息转换为JSON字符串
	messagesJSON, err := json.Marshal(messages)
	if err == nil {
		history.Messages = string(messagesJSON)
		
		// 保存到数据库
		result := models.DB.Create(&history)
		if result.Error == nil {
			return history.HistoryID
		}
	}
	
	return ""
}

// updateChatHistoryFromStream 更新现有的聊天历史记录
func updateChatHistoryFromStream(historyID string, userMessages []config.Message, aiMessage config.Message) bool {
	// 获取现有的聊天历史记录
	history, err := models.GetChatHistoryByHistoryID(historyID)
	if err != nil {
		fmt.Println("获取聊天历史记录失败:", err)
		return false
	}
	
	// 创建包含用户消息和AI响应的完整消息列表
	messages := append(userMessages, aiMessage)
	
	// 将消息转换为JSON字符串
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		fmt.Println("序列化消息失败:", err)
		return false
	}
	
	// 更新消息内容
	history.Messages = string(messagesJSON)
	
	// 保存到数据库
	result := models.DB.Save(history)
	if result.Error != nil {
		fmt.Println("更新聊天历史记录失败:", result.Error)
		return false
	}
	
	return true
}

