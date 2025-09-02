package main

import (
	"log"
	"net/http"

	"fly-print-cloud/api/internal/auth"
	"fly-print-cloud/api/internal/config"
	"fly-print-cloud/api/internal/database"
	"fly-print-cloud/api/internal/handlers"
	"fly-print-cloud/api/internal/middleware"
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

	// 创建默认管理员账户
	if err := db.CreateDefaultAdmin(); err != nil {
		log.Fatal("Failed to create default admin:", err)
	}

	// 初始化服务
	userRepo := database.NewUserRepository(db)
	edgeNodeRepo := database.NewEdgeNodeRepository(db)
	printerRepo := database.NewPrinterRepository(db)
	printJobRepo := database.NewPrintJobRepository(db)
	authService := auth.NewService(&cfg.JWT)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userRepo, authService)
	userHandler := handlers.NewUserHandler(userRepo)
	edgeNodeHandler := handlers.NewEdgeNodeHandler(edgeNodeRepo)
	printerHandler := handlers.NewPrinterHandler(printerRepo, edgeNodeRepo)
	printJobHandler := handlers.NewPrintJobHandler(printJobRepo)

	// 创建Gin路由
	r := gin.New()

	// 添加中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// 设置路由
	setupRoutes(r, authHandler, userHandler, edgeNodeHandler, printerHandler, printJobHandler, authService, userRepo)

	// 启动服务器
	serverAddr := cfg.Server.GetServerAddr()
	log.Printf("Starting %s server on %s", cfg.App.Name, serverAddr)
	log.Printf("Environment: %s, Debug: %v", cfg.App.Environment, cfg.App.Debug)
	
	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, edgeNodeHandler *handlers.EdgeNodeHandler, printerHandler *handlers.PrinterHandler, printJobHandler *handlers.PrintJobHandler, authService *auth.Service, userRepo *database.UserRepository) {
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

	// 认证路由组（/auth）
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.GET("/me", middleware.AuthMiddleware(authService, userRepo), authHandler.Me)
	}

	// 管理路由组（/admin）- 需要认证
	adminGroup := r.Group("/admin", middleware.AuthMiddleware(authService, userRepo))
	{
		// 用户管理路由 - 需要admin权限
		userGroup := adminGroup.Group("/users", middleware.RequireRole("admin"))
		{
			userGroup.GET("", userHandler.ListUsers)
			userGroup.POST("", userHandler.CreateUser)
			userGroup.GET("/:id", userHandler.GetUser)
			userGroup.PUT("/:id", userHandler.UpdateUser)
			userGroup.DELETE("/:id", userHandler.DeleteUser)
			userGroup.PUT("/:id/password", userHandler.ChangePassword)
		}

		// Edge Node 管理路由 - 需要admin或operator权限
		edgeNodeGroup := adminGroup.Group("/edge-nodes", middleware.RequireRole("admin", "operator"))
		{
			edgeNodeGroup.GET("", edgeNodeHandler.ListEdgeNodes)
			edgeNodeGroup.GET("/:id", edgeNodeHandler.GetEdgeNode)
			edgeNodeGroup.PUT("/:id", edgeNodeHandler.UpdateEdgeNode)
			edgeNodeGroup.DELETE("/:id", edgeNodeHandler.DeleteEdgeNode)
		}

		// 打印机管理路由 - 需要admin或operator权限
		printerGroup := adminGroup.Group("/printers", middleware.RequireRole("admin", "operator"))
		{
			printerGroup.GET("", printerHandler.ListPrinters)
			printerGroup.GET("/:id", printerHandler.GetPrinter)
			printerGroup.PUT("/:id", printerHandler.UpdatePrinter)
			printerGroup.DELETE("/:id", printerHandler.DeletePrinter)
		}

		// 打印任务管理路由 - 需要认证，所有角色都可以访问
		printJobGroup := adminGroup.Group("/print-jobs")
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

	// API路由组（/api）- 用于对外接口
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": "success",
				"data": gin.H{
					"status":  "ok",
					"service": "fly-print-cloud-api",
					"version": "0.1.0",
				},
			})
		})

		// Edge 路由组 - 用于Edge Node连接，需要 OAuth2 验证
		edgeGroup := apiGroup.Group("/edge")
		{
			edgeGroup.POST("/register", middleware.OAuth2ResourceServer("edge:register"), edgeNodeHandler.RegisterEdgeNode)
			edgeGroup.POST("/heartbeat", middleware.OAuth2ResourceServer("edge:heartbeat"), edgeNodeHandler.Heartbeat)
			
			// Edge Node 的打印机管理
			edgeGroup.POST("/:node_id/printers", middleware.OAuth2ResourceServer("edge:printer"), printerHandler.EdgeRegisterPrinter)
			edgeGroup.GET("/:node_id/printers", middleware.OAuth2ResourceServer("edge:printer"), printerHandler.EdgeListPrinters)
		}
	}

	// WebSocket路由（将来实现）
	// r.GET("/ws/edge", wsHandler.HandleEdgeConnection)
}
