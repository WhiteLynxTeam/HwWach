package routes

import (
	"HwWach/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Asset         handlers.AssetHandler
	Roles         handlers.RolesHandler
	Photo         handlers.PhotoHandler
	Request       handlers.RequestHandler
	ChangeRequest handlers.AssetChangeRequestHandler
	Category      handlers.CategoryHandler
	Audit         handlers.AuditHandler
	Internal      handlers.InternalHandler
}

func SetupRoutes(r *gin.Engine, h Handlers, jwtMiddleware, internalMiddleware gin.HandlerFunc) {
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
		assets.POST("", h.Asset.CreateAsset)
		assets.GET("", h.Asset.ListUserAssets)
		assets.GET("/paginated", h.Asset.ListUserAssetsPaginated)
		assets.GET("/check-inventory", h.Asset.CheckInventoryUnique)
		assets.GET("/change-requests", h.ChangeRequest.ListPending)
		assets.PATCH("/change-requests/:id", h.ChangeRequest.ApproveRequest)
		assets.GET("/:id", h.Asset.GetAsset)
		assets.PUT("/:id", h.Asset.UpdateAsset)
		assets.DELETE("/:id", h.Asset.DeleteAsset)
		assets.POST("/:id/moderate", h.Asset.ModerateAsset)
		assets.GET("/:id/photos", h.Asset.ListAssetPhotos)
		assets.GET("/:id/history", h.Asset.History)
		assets.POST("/:id/change-requests", h.ChangeRequest.CreateRequest)

		// состояние актива (3.6)
		assets.POST("/:id/status", h.Asset.ChangeStatus)
		assets.POST("/:id/repair/start", h.Asset.StartRepair)
		assets.POST("/:id/repair/finish", h.Asset.FinishRepair)

		// бизнес-роли и передачи (3.2, 3.3)
		assets.POST("/:id/mol", h.Roles.AssignMol)
		assets.DELETE("/:id/mol", h.Roles.RemoveMol)
		assets.POST("/:id/users", h.Roles.AddUser)
		assets.DELETE("/:id/users/:person_id", h.Roles.RemoveUser)
		assets.GET("/:id/transfers", h.Roles.Transfers)
	}

	r.GET("/me/assignments", jwtMiddleware, h.Roles.MyAssignments)

	roles := r.Group("/asset-roles", jwtMiddleware)
	{
		roles.GET("/disputes", h.Roles.Disputes)
		roles.POST("/:id/ack", h.Roles.Ack)
		roles.POST("/:id/resolve", h.Roles.Resolve)
	}

	categories := r.Group("/categories", jwtMiddleware)
	{
		categories.GET("", h.Category.ListAdmin)
		categories.POST("", h.Category.AdminCreate)
		categories.GET("/search", h.Category.Search)
		categories.GET("/sync", h.Category.Sync)
		categories.GET("/pending", h.Category.ListPending)
		categories.GET("/:id/affected", h.Category.GetAffectedAssets)
		categories.PUT("/:id", h.Category.AdminUpdate)
		categories.POST("/:id/merge", h.Category.AdminMerge)
		categories.PATCH("/:id/moderate", h.Category.Moderate)
	}

	r.GET("/photos/:id/download", h.Photo.DownloadPhoto)

	photos := r.Group("/photos", jwtMiddleware)
	{
		photos.POST("/upload-url", h.Photo.UploadPhoto)
		photos.POST("/complete-upload", h.Photo.ConfirmUpload)
		photos.GET("", h.Photo.ListUserPhotos)
	}

	req := r.Group("/requests", jwtMiddleware)
	{
		req.POST("", h.Request.CreateRequest)
		req.GET("", h.Request.ListRequests)
		req.GET("/:id", h.Request.GetRequest)
		req.PATCH("/:id", h.Request.UpdateRequestStatus)
		req.DELETE("/:id", h.Request.DeleteRequest)
	}

	r.GET("/audit", jwtMiddleware, h.Audit.List)

	// внутренние запросы ts → go (5.1)
	internal := r.Group("/internal", internalMiddleware)
	{
		internal.GET("/persons/:id/assets-count", h.Internal.PersonAssetsCount)
		internal.POST("/persons/:id/dismissed", h.Internal.PersonDismissed)
		internal.DELETE("/organizations/:id", h.Internal.DeleteOrganization)
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
