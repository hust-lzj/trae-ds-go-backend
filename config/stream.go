package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StreamChat 发送聊天请求并以流式方式处理响应
func (c *LLMClient) StreamChat(w http.ResponseWriter, messages []Message, options map[string]interface{},model string) error {
	// 准备请求数据
	reqData := ChatRequest{
		Model:    model, // 使用deepseek模型
		Messages: messages,
		Options:  options,
	}

	// 添加流式选项
	if reqData.Options == nil {
		reqData.Options = make(map[string]interface{})
	}
	reqData.Options["stream"] = true

	// 序列化请求数据
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.Config.APIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	// 设置Accept头以接收流式响应
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// 创建缓冲区读取器
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// 将读取到的数据写入响应
			if _, err := w.Write(buffer[:n]); err != nil {
				fmt.Println("写入响应失败:", err)
				break
			}
			// 刷新响应，确保数据立即发送
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}

		// 检查是否读取完毕
		if err != nil {
			if err != io.EOF {
				fmt.Println("读取响应失败:", err)
			}
			break
		}
	}

	return nil
}