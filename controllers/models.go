package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// ModelInfo 模型信息结构
type ModelInfo struct {
	Name string `json:"name"`
}

// ModelsResponse Ollama API返回的模型列表响应
type ModelsResponse struct {
	Models []struct {
		Name      string `json:"name"`
		Model     string `json:"model"`
		ModifiedAt string `json:"modified_at"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
		Details   struct {
			ParentModel      string   `json:"parent_model"`
			Format           string   `json:"format"`
			Family           string   `json:"family"`
			Families         []string `json:"families"`
			ParameterSize    string   `json:"parameter_size"`
			QuantizationLevel string   `json:"quantization_level"`
		} `json:"details"`
	} `json:"models"`
}

// GetModels 获取本地模型列表
func GetModels(c *gin.Context) {
	// 从环境变量获取LLM API URL的基础部分
	apiURLBase := os.Getenv("LLM_API_URL")
	if apiURLBase == "" {
		apiURLBase = "http://localhost:11434/api"
	} else {
		// 如果环境变量中的URL包含/chat/，则去掉这部分，只保留基础URL
		apiURLBase = strings.TrimSuffix(apiURLBase, "/chat/")
		apiURLBase = strings.TrimSuffix(apiURLBase, "/chat")
	}

	// 构建获取模型列表的URL
	modelsURL := apiURLBase + "/tags"

	// 创建HTTP客户端
	client := &http.Client{}

	// 创建GET请求
	req, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建请求失败: %v", err)})
		return
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("发送请求失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("获取模型列表失败，状态码: %d", resp.StatusCode)})
		return
	}

	// 解析响应
	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析响应失败: %v", err)})
		return
	}

	// 提取模型名称
	var modelNames []string
	for _, model := range modelsResp.Models {
		modelNames = append(modelNames, model.Name)
	}

	// 返回模型名称列表
	c.JSON(http.StatusOK, gin.H{"models": modelNames})
}