package handlers

import "github.com/gin-gonic/gin"

type AssetHandler interface {
	CreateAsset(c *gin.Context)
	ListUserAssets(c *gin.Context)
	GetAsset(c *gin.Context)
	UpdateAsset(c *gin.Context)
	DeleteAsset(c *gin.Context)
	ModerateAsset(c *gin.Context)
	ListAssetPhotos(c *gin.Context)
	CheckInventoryUnique(c *gin.Context)
	ListUserAssetsPaginated(c *gin.Context)

	ChangeStatus(c *gin.Context)
	StartRepair(c *gin.Context)
	FinishRepair(c *gin.Context)
	History(c *gin.Context)
}

type RolesHandler interface {
	AssignMol(c *gin.Context)
	RemoveMol(c *gin.Context)
	AddUser(c *gin.Context)
	RemoveUser(c *gin.Context)
	MyAssignments(c *gin.Context)
	Ack(c *gin.Context)
	Disputes(c *gin.Context)
	Resolve(c *gin.Context)
	Transfers(c *gin.Context)
}

type PhotoHandler interface {
	UploadPhoto(c *gin.Context)
	ConfirmUpload(c *gin.Context)
	ListUserPhotos(c *gin.Context)
	DownloadPhoto(c *gin.Context)
}

type RequestHandler interface {
	CreateRequest(c *gin.Context)
	ListRequests(c *gin.Context)
	GetRequest(c *gin.Context)
	UpdateRequestStatus(c *gin.Context)
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

type AuditHandler interface {
	List(c *gin.Context)
}

// InternalHandler внутренние запросы ts → go (5.1)
type InternalHandler interface {
	PersonAssetsCount(c *gin.Context)
	PersonDismissed(c *gin.Context)
	DeleteOrganization(c *gin.Context)
}
