package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LLMConfig 模型配置
type LLMConfig struct {
	APIURL     string
	MaxTokens  int
	Temperature float64
	Timeout     time.Duration
}

// DefaultLLMConfig 默认模型配置
var DefaultLLMConfig = LLMConfig{
	APIURL:     "http://localhost:11434/api/chat", // 默认本地deepseek模型API地址
	MaxTokens:  2048,                                       // 默认最大生成token数
	Temperature: 0.7,                                        // 默认温度参数
	Timeout:     time.Second * 120,                           // 默认超时时间
}

// Message 聊天消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Model    string                  `json:"model"`
	Messages []Message               `json:"messages"`
	Options  map[string]interface{} `json:"options"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// LLMClient 模型客户端
type LLMClient struct {
	Config LLMConfig
	Client *http.Client
}

// NewLLMClient 创建新的模型客户端
func NewLLMClient() *LLMClient {
	// 从环境变量获取配置
	apiURL := os.Getenv("LLM_API_URL")
	if apiURL == "" {
		apiURL = DefaultLLMConfig.APIURL
	}

	// 创建客户端
	client := &LLMClient{
		Config: LLMConfig{
			APIURL:     apiURL,
			MaxTokens:  DefaultLLMConfig.MaxTokens,
			Temperature: DefaultLLMConfig.Temperature,
			Timeout:     DefaultLLMConfig.Timeout,
		},
		Client: &http.Client{
			Timeout: DefaultLLMConfig.Timeout,
		},
	}

	return client
}

// Chat 发送聊天请求并获取响应
func (c *LLMClient) Chat(messages []Message, options map[string]interface{}) (*ChatResponse, error) {
	// 准备请求数据
	reqData := ChatRequest{
		Model:    "deepseek-r1:7b", // 使用deepseek模型
		Messages: messages,
		Options:  options,
	}

	// 序列化请求数据
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 打印请求信息
	reqJSON, _ := json.MarshalIndent(reqData, "", "  ")
	fmt.Println("发送给模型的请求数据:")
	fmt.Println(string(reqJSON))

	// 打印完整请求信息
	fmt.Println("\n完整请求信息:")
	fmt.Printf("请求URL: %s\n", c.Config.APIURL)
	fmt.Println("请求方法: POST")
	fmt.Println("请求头信息:")
	fmt.Println("Content-Type: application/json")
	fmt.Println("\n请求体原始数据:")
	fmt.Println(string(reqBody))

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.Config.APIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &chatResp, nil
}
