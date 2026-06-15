package routes

import (
	"HwWach/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(
	r *gin.Engine,
	assetHandler handlers.AssetHandler,
	photoHandler handlers.PhotoHandler,
	requestHandler handlers.RequestHandler,
	changeRequestHandler handlers.AssetChangeRequestHandler,
	jwtMiddleware gin.HandlerFunc,
) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte(indexHTML))
	})

	r.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	assets := r.Group("/assets", jwtMiddleware)
	{
		assets.POST("", assetHandler.CreateAsset)
		assets.GET("", assetHandler.ListUserAssets)
		assets.GET("/paginated", assetHandler.ListUserAssetsPaginated)
		assets.GET("/check-inventory", assetHandler.CheckInventoryUnique)
		assets.GET("/:id", assetHandler.GetAsset)
		assets.PUT("/:id", assetHandler.UpdateAsset)
		assets.GET("/:id/photos", assetHandler.ListAssetPhotos)

		assets.POST("/:id/change-requests", changeRequestHandler.CreateRequest)
		assets.GET("/change-requests", changeRequestHandler.ListPending)
		assets.PATCH("/change-requests/:id", changeRequestHandler.ApproveRequest)
	}

	photos := r.Group("/photos", jwtMiddleware)
	{
		photos.POST("/upload-url", photoHandler.UploadPhoto)
		photos.POST("/complete-upload", photoHandler.ConfirmUpload)
		photos.GET("", photoHandler.ListUserPhotos)
		photos.DELETE("/:id", photoHandler.DeletePhoto)
	}

	req := r.Group("/requests", jwtMiddleware)
	{
		req.POST("", requestHandler.CreateRequest)
		req.GET("/:id", requestHandler.GetRequest)
		req.DELETE("/:id", requestHandler.DeleteRequest)
	}
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HwWach API Service</title>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #0b0f19;
            --card-bg: rgba(255, 255, 255, 0.03);
            --border-color: rgba(255, 255, 255, 0.08);
            --text-primary: #f3f4f6;
            --text-secondary: #9ca3af;
            --accent-green: #10b981;
            --accent-blue: #3b82f6;
        }
        
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-primary);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            overflow: hidden;
            position: relative;
        }

        body::before {
            content: "";
            position: absolute;
            width: 600px;
            height: 600px;
            background: radial-gradient(circle, rgba(59, 130, 246, 0.08) 0%, rgba(0, 0, 0, 0) 70%);
            top: -100px;
            right: -100px;
            z-index: 0;
        }

        body::after {
            content: "";
            position: absolute;
            width: 600px;
            height: 600px;
            background: radial-gradient(circle, rgba(16, 185, 129, 0.05) 0%, rgba(0, 0, 0, 0) 70%);
            bottom: -100px;
            left: -100px;
            z-index: 0;
        }

        .container {
            position: relative;
            z-index: 10;
            width: 100%;
            max-width: 480px;
            padding: 24px;
        }

        .card {
            background: var(--card-bg);
            backdrop-filter: blur(16px);
            -webkit-backdrop-filter: blur(16px);
            border: 1px solid var(--border-color);
            border-radius: 24px;
            padding: 40px 32px;
            text-align: center;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
            animation: fadeIn 0.8s ease-out;
        }

        @keyframes fadeIn {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        .logo {
            font-size: 28px;
            font-weight: 700;
            background: linear-gradient(135deg, #3b82f6, #10b981);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 24px;
            letter-spacing: -0.5px;
        }

        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            background: rgba(16, 185, 129, 0.1);
            border: 1px solid rgba(16, 185, 129, 0.2);
            padding: 6px 16px;
            border-radius: 9999px;
            color: var(--accent-green);
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 24px;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            background-color: var(--accent-green);
            border-radius: 50%;
            box-shadow: 0 0 12px var(--accent-green);
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0% {
                transform: scale(0.95);
                box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
            }
            70% {
                transform: scale(1);
                box-shadow: 0 0 0 8px rgba(16, 185, 129, 0);
            }
            100% {
                transform: scale(0.95);
                box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
            }
        }

        h1 {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 12px;
            color: var(--text-primary);
        }

        p {
            font-size: 14px;
            color: var(--text-secondary);
            line-height: 1.6;
            margin-bottom: 32px;
        }

        .divider {
            height: 1px;
            background: var(--border-color);
            margin: 24px 0;
        }

        .meta-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 16px;
            text-align: left;
        }

        .meta-item {
            background: rgba(255, 255, 255, 0.015);
            border: 1px solid rgba(255, 255, 255, 0.04);
            border-radius: 12px;
            padding: 12px 16px;
        }

        .meta-label {
            font-size: 11px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: var(--text-secondary);
            margin-bottom: 4px;
        }

        .meta-value {
            font-size: 13px;
            font-weight: 500;
            color: var(--text-primary);
        }

        .meta-value a {
            color: #3b82f6;
            text-decoration: none;
            transition: color 0.2s ease;
        }

        .meta-value a:hover {
            color: #60a5fa;
            text-decoration: underline;
        }

        .footer {
            margin-top: 24px;
            font-size: 12px;
            color: rgba(255, 255, 255, 0.3);
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">HwWach Backend</div>
            <div class="status-badge">
                <span class="status-dot"></span>
                Service Operational
            </div>
            <h1>API is running</h1>
            <p>The Go backend service is fully operational and healthy. Standard endpoints are active.</p>
            
            <div class="divider"></div>
            
            <div class="meta-grid">
                <div class="meta-item">
                    <div class="meta-label">Environment</div>
                    <div class="meta-value">Render Production</div>
                </div>
                <div class="meta-item">
                    <div class="meta-label">API Documentation</div>
                    <div class="meta-value"><a href="/swagger/index.html">View Swagger UI</a></div>
                </div>
            </div>
        </div>
        <div class="footer">
            &copy; 2026 HwWach. All rights reserved.
        </div>
    </div>
</body>
</html>`
