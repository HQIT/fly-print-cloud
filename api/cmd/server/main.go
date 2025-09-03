package main

import (
	"log"
	"net/http"

	"fly-print-cloud/api/internal/config"
	"fly-print-cloud/api/internal/database"
	"fly-print-cloud/api/internal/handlers"
	"fly-print-cloud/api/internal/middleware"
	"fly-print-cloud/api/internal/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 设置Gin模式
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 连接数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 初始化数据库表
	if err := db.InitTables(); err != nil {
		log.Fatal("Failed to initialize database tables:", err)
	}



	// 初始化服务
	userRepo := database.NewUserRepository(db)
	edgeNodeRepo := database.NewEdgeNodeRepository(db)
	printerRepo := database.NewPrinterRepository(db)
	printJobRepo := database.NewPrintJobRepository(db)

	// 初始化处理器
	userHandler := handlers.NewUserHandler(userRepo)
	edgeNodeHandler := handlers.NewEdgeNodeHandler(edgeNodeRepo)
	printerHandler := handlers.NewPrinterHandler(printerRepo, edgeNodeRepo)
	printJobHandler := handlers.NewPrintJobHandler(printJobRepo)
	oauth2Handler := handlers.NewOAuth2Handler(&cfg.OAuth2, &cfg.Admin)

	// 初始化 WebSocket 管理器
	wsManager := websocket.NewConnectionManager()
	wsHandler := websocket.NewWebSocketHandler(wsManager)

	// 启动 WebSocket 管理器
	go wsManager.Run()

	// 创建Gin路由
	r := gin.New()

	// 添加中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// 设置路由
	setupRoutes(r, userHandler, edgeNodeHandler, printerHandler, printJobHandler, wsHandler, oauth2Handler)

	// 启动服务器
	serverAddr := cfg.Server.GetServerAddr()
	log.Printf("Starting %s server on %s", cfg.App.Name, serverAddr)
	log.Printf("Environment: %s, Debug: %v", cfg.App.Environment, cfg.App.Debug)
	
	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(r *gin.Engine, userHandler *handlers.UserHandler, edgeNodeHandler *handlers.EdgeNodeHandler, printerHandler *handlers.PrinterHandler, printJobHandler *handlers.PrintJobHandler, wsHandler *websocket.WebSocketHandler, oauth2Handler *handlers.OAuth2Handler) {
	// 公开路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "success",
			"data": gin.H{
				"status":  "ok",
				"service": "fly-print-cloud-api",
			},
		})
	})

	// OAuth2 认证路由
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/login", oauth2Handler.Login)
		authGroup.GET("/callback", oauth2Handler.Callback)
		authGroup.GET("/me", oauth2Handler.Me)
		authGroup.GET("/verify", oauth2Handler.Verify)  // Nginx auth_request 使用
		authGroup.GET("/logout", oauth2Handler.Logout)   // 支持 GET 请求登出
		authGroup.POST("/logout", oauth2Handler.Logout)  // 保留 POST 支持
	}

	// 统一 API 路由组（/api/v1）- OAuth2 Resource Server
	apiV1Group := r.Group("/api/v1")
	{
		apiV1Group.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": "success",
				"data": gin.H{
					"status":  "ok",
					"service": "fly-print-cloud-api",
					"version": "1.0.0",
				},
			})
		})

		// Admin Console API - 需要 admin:* scope
		adminGroup := apiV1Group.Group("/admin")
		{
			// 用户管理路由 - 需要 admin 权限
			userGroup := adminGroup.Group("/users", middleware.OAuth2ResourceServer("fly-print-admin"))
			{
				userGroup.GET("", userHandler.ListUsers)
				userGroup.POST("", userHandler.CreateUser)
				userGroup.GET("/:id", userHandler.GetUser)
				userGroup.PUT("/:id", userHandler.UpdateUser)
				userGroup.DELETE("/:id", userHandler.DeleteUser)
				userGroup.PUT("/:id/password", userHandler.ChangePassword)
			}
			
			// 当前用户业务信息 - 任何管理员都可以访问自己的档案
			adminGroup.GET("/profile", userHandler.GetCurrentUserProfile)

			// Edge Node 管理路由 - 需要 admin 或 operator 权限
			edgeNodeGroup := adminGroup.Group("/edge-nodes", middleware.OAuth2ResourceServer("fly-print-admin", "fly-print-operator"))
			{
				edgeNodeGroup.GET("", edgeNodeHandler.ListEdgeNodes)
				edgeNodeGroup.GET("/:id", edgeNodeHandler.GetEdgeNode)
				edgeNodeGroup.PUT("/:id", edgeNodeHandler.UpdateEdgeNode)
				edgeNodeGroup.DELETE("/:id", edgeNodeHandler.DeleteEdgeNode)
			}

			// 打印机管理路由 - 需要 admin 或 operator 权限
			printerGroup := adminGroup.Group("/printers", middleware.OAuth2ResourceServer("fly-print-admin", "fly-print-operator"))
			{
				printerGroup.GET("", printerHandler.ListPrinters)
				printerGroup.GET("/:id", printerHandler.GetPrinter)
				printerGroup.PUT("/:id", printerHandler.UpdatePrinter)
				printerGroup.DELETE("/:id", printerHandler.DeletePrinter)
			}

			// 打印任务管理路由 - 需要 admin 或 operator 权限
			printJobGroup := adminGroup.Group("/print-jobs", middleware.OAuth2ResourceServer("fly-print-admin", "fly-print-operator"))
			{
				printJobGroup.POST("", printJobHandler.CreatePrintJob)
				printJobGroup.GET("", printJobHandler.ListPrintJobs)
				printJobGroup.GET("/:id", printJobHandler.GetPrintJob)
				printJobGroup.PUT("/:id", printJobHandler.UpdatePrintJob)
				printJobGroup.DELETE("/:id", printJobHandler.DeletePrintJob)
				printJobGroup.POST("/:id/cancel", printJobHandler.CancelPrintJob)
				printJobGroup.POST("/:id/retry", printJobHandler.RetryPrintJob)
			}
		}

		// Edge Node API - 需要 edge:* scope
		edgeGroup := apiV1Group.Group("/edge")
		{
			edgeGroup.POST("/register", middleware.OAuth2ResourceServer("edge:register"), edgeNodeHandler.RegisterEdgeNode)
			edgeGroup.POST("/heartbeat", middleware.OAuth2ResourceServer("edge:heartbeat"), edgeNodeHandler.Heartbeat)
			
			// Edge Node 的打印机管理
			edgeGroup.POST("/:node_id/printers", middleware.OAuth2ResourceServer("edge:printer"), printerHandler.EdgeRegisterPrinter)
			edgeGroup.GET("/:node_id/printers", middleware.OAuth2ResourceServer("edge:printer"), printerHandler.EdgeListPrinters)
			
			// WebSocket 连接
			edgeGroup.GET("/ws", wsHandler.HandleConnection)
		}
	}


}
