package handlers

import "github.com/gin-gonic/gin"

type AssetHandler interface {
	CreateAsset(c *gin.Context)
	ListUserAssets(c *gin.Context)
	GetAsset(c *gin.Context)
	UpdateAsset(c *gin.Context)
	ListAssetPhotos(c *gin.Context)
}

type PhotoHandler interface {
	UploadPhoto(c *gin.Context)
	ConfirmUpload(c *gin.Context)
	ListUserPhotos(c *gin.Context)
	DeletePhoto(c *gin.Context)
}

type RequestHandler interface {
	CreateRequest(c *gin.Context)
	GetRequest(c *gin.Context)
	DeleteRequest(c *gin.Context)
}

type AssetChangeRequestHandler interface {
	CreateRequest(c *gin.Context)
	ListPending(c *gin.Context)
	ApproveRequest(c *gin.Context)
}
