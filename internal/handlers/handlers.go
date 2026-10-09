package handlers

import "github.com/gin-gonic/gin"

type AssetHandler interface {
	CreateAsset(c *gin.Context)
	ListUserAssets(c *gin.Context)
	GetAsset(c *gin.Context)
	UpdateAsset(c *gin.Context)
	ListAssetPhotos(c *gin.Context)
	CheckInventoryUnique(c *gin.Context)
	ListUserAssetsPaginated(c *gin.Context)
}

type PhotoHandler interface {
	UploadPhoto(c *gin.Context)
	ConfirmUpload(c *gin.Context)
	ListUserPhotos(c *gin.Context)
	DownloadPhoto(c *gin.Context)
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

type CategoryHandler interface {
	Search(c *gin.Context)
	Sync(c *gin.Context)
	ListPending(c *gin.Context)
	Moderate(c *gin.Context)
	ListAdmin(c *gin.Context)
	GetAffectedAssets(c *gin.Context)
	AdminCreate(c *gin.Context)
	AdminUpdate(c *gin.Context)
	AdminMerge(c *gin.Context)
}
