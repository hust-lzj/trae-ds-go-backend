package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/trae-ds-go-backend/controllers"
	"github.com/trae-ds-go-backend/middleware"
	"github.com/trae-ds-go-backend/models"
)

func main() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Println("未找到.env文件，使用默认环境变量")
	}

	// 初始化数据库
	models.ConnectDatabase()

	// 设置Gin模式
	gin.SetMode(getGinMode())

	// 创建Gin路由
	r := gin.Default()

	// 注册路由
	setupRoutes(r)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	log.Printf("服务器运行在 http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("无法启动服务器: %v", err)
	}
}

// 设置路由
func setupRoutes(r *gin.Engine) {
	// 公开路由
	public := r.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.GET("/models", controllers.GetModels) // 添加获取模型列表的路由
	}

	// 需要认证的路由
	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		protected.POST("/stream-chat", controllers.StreamChat) // 添加流式聊天路由
		
		// 聊天历史记录相关路由
		protected.POST("/chat-history", controllers.SaveChatHistory)
		protected.GET("/chat-histories", controllers.GetUserChatHistories)
		protected.GET("/chat-history/:history_id", controllers.GetChatHistoryDetail)
		protected.DELETE("/chat-history/:id", controllers.DeleteChatHistory)
	}
}

// 获取Gin模式
func getGinMode() string {
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		return gin.DebugMode
	}
	return mode
}