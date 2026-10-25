package services

// Интеграционные тесты сервисов на настоящей PostgreSQL: роли, передачи, статусы, ремонт.
// Запуск: TEST_DATABASE_URL=postgres://... go test ./internal/services/ -run Integration
// Без переменной тесты пропускаются. База очищается полностью (схема public пересоздаётся) — только тестовая!

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"HwWach/internal/tsclient"
	"HwWach/migrations"
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type fakeTS struct {
	persons map[uuid.UUID]*tsclient.Person
}

func (f *fakeTS) GetPerson(_ context.Context, id uuid.UUID) (*tsclient.Person, error) {
	p, ok := f.persons[id]
	if !ok {
		return nil, tsclient.ErrPersonNotFound
	}
	return p, nil
}
func (f *fakeTS) ReportFilled(context.Context, uuid.UUID) error { return nil }
func (f *fakeTS) Touch(context.Context, uuid.UUID) error        { return nil }

type fakeStorage struct{}

func (fakeStorage) PresignedGetURL(context.Context, string, time.Duration) (string, error) {
	return "", nil
}
func (fakeStorage) PresignedPutURL(context.Context, string, string, time.Duration) (string, error) {
	return "", nil
}
func (fakeStorage) Delete(context.Context, string) error { return nil }
func (fakeStorage) GetPublicURL(s string) string         { return s }

type env struct {
	db        *gorm.DB
	ts        *fakeTS
	assets    AssetService
	roles     RolesService
	lifecycle LifecycleService
	requests  RequestService
	orgs      OrgService
	org       uuid.UUID
}

func setup(t *testing.T) *env {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL не задан")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	goose.SetBaseFS(migrations.FS)
	_ = goose.SetDialect("postgres")
	goose.SetLogger(goose.NopLogger())
	if _, err := sqlDB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		t.Fatal(err)
	}
	ts := &fakeTS{persons: map[uuid.UUID]*tsclient.Person{}}
	audit := NewAuditService(db)
	orgs := NewOrgService(db, ts, fakeStorage{}, audit)
	roles := NewRolesService(db, ts, audit)
	lifecycle := NewLifecycleService(db, roles, audit)
	return &env{
		db: db, ts: ts, roles: roles, lifecycle: lifecycle, orgs: orgs,
		assets:   NewAssetService(db, roles, orgs, audit),
		requests: NewRequestService(db, lifecycle, audit),
		org:      uuid.New(),
	}
}

// person человек организации; principal — его учётная запись с правами роли
func (e *env) person(isMol bool) uuid.UUID {
	id := uuid.New()
	e.ts.persons[id] = &tsclient.Person{ID: id, OrganizationID: e.org, IsMol: isMol, IsActive: true}
	return id
}

func (e *env) principal(person uuid.UUID, perms ...string) *authz.Principal {
	p := &authz.Principal{AccountID: uuid.New(), PersonID: person, OrgID: e.org, Perms: map[string]bool{}}
	for _, perm := range perms {
		p.Perms[perm] = true
	}
	p.IsMol = e.ts.persons[person] != nil && e.ts.persons[person].IsMol
	return p
}

var adminPerms = []string{
	authz.PermAssetsManage, authz.PermAssetsViewAll, authz.PermAssetsPropose, authz.PermAssetsStatus, authz.PermAssetsDecommission,
	authz.PermAssetRolesAssign, authz.PermTransfersOwn, authz.PermRequestsCreate, authz.PermAuditView,
}
var memberPerms = []string{authz.PermAssetsPropose, authz.PermTransfersOwn, authz.PermRequestsCreate}

func statusOf(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status
	}
	return 0
}

func TestIntegrationRolesAndLifecycle(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	adminPerson := e.person(true)
	admin := e.principal(adminPerson, adminPerms...)
	molA := e.person(true)
	molB := e.person(true)
	worker := e.person(false)
	memberMol := e.principal(molA, memberPerms...)
	memberWorker := e.principal(worker, memberPerms...)

	// Создатель-участник без МОЛ: карточка на модерации, без МОЛ
	created, err := e.assets.Create(ctx, memberWorker, CreateAssetInput{Asset: &models.Asset{Name: "Монитор", Category: "монитор"}})
	if err != nil {
		t.Fatal(err)
	}
	if created.Asset.ModerationStatus != models.ModerationPending || created.Mol != nil {
		t.Fatalf("ожидалась pending-карточка без МОЛ: %+v mol=%v", created.Asset.ModerationStatus, created.Mol)
	}
	// Участник не может назначить МОЛ другого человека
	if _, err := e.assets.Create(ctx, memberWorker, CreateAssetInput{Asset: &models.Asset{Name: "X", Category: "x"}, MolPersonID: &molA}); statusOf(err) != http.StatusForbidden {
		t.Fatalf("ожидался 403, получено %v", err)
	}

	// «Без МОЛ» у админа
	list, _, err := e.assets.List(ctx, admin, AssetFilter{NoMol: true}, 1, 50)
	if err != nil || len(list) != 1 {
		t.Fatalf("Без МОЛ: %v %d", err, len(list))
	}

	// Одобрение с первичным назначением МОЛ
	view, err := e.assets.Moderate(ctx, admin, created.Asset.UUID, true, "ok", &molA)
	if err != nil {
		t.Fatal(err)
	}
	if view.Mol == nil || view.Mol.PersonID != molA || view.Mol.AckStatus != models.AckPending {
		t.Fatalf("МОЛ после одобрения: %+v", view.Mol)
	}
	assetID := created.Asset.UUID

	// Не МОЛ (работник) передать не может
	if _, err := e.roles.AssignMol(ctx, memberWorker, assetID, molB, ""); statusOf(err) != http.StatusForbidden {
		t.Fatalf("ожидался 403: %v", err)
	}
	// Нельзя записать актив за не-МОЛ
	if _, err := e.roles.AssignMol(ctx, admin, assetID, worker, ""); statusOf(err) != http.StatusBadRequest {
		t.Fatalf("ожидался 400 (не МОЛ): %v", err)
	}
	// МОЛ передаёт свой актив другому МОЛ — сразу, без подтверждения
	if _, err := e.roles.AssignMol(ctx, memberMol, assetID, molB, "в ремонт мастеру"); err != nil {
		t.Fatal(err)
	}
	// МОЛ (уже прежний) больше не может распоряжаться
	if _, err := e.roles.AddUser(ctx, memberMol, assetID, worker, ""); statusOf(err) != http.StatusForbidden {
		t.Fatalf("прежний МОЛ не должен назначать: %v", err)
	}
	memberMolB := e.principal(molB, memberPerms...)
	if _, err := e.roles.AddUser(ctx, memberMolB, assetID, worker, ""); err != nil {
		t.Fatal(err)
	}

	// Работающий отмечает «неисправен»; поставить «на склад» не может
	faulty := models.ConditionFaulty
	if err := e.lifecycle.ChangeStatus(ctx, memberWorker, assetID, StatusInput{Condition: &faulty}); err != nil {
		t.Fatal(err)
	}
	stock := models.AccountingInStock
	if err := e.lifecycle.ChangeStatus(ctx, memberWorker, assetID, StatusInput{AccountingStatus: &stock}); statusOf(err) != http.StatusForbidden {
		t.Fatalf("работающий не меняет учёт: %v", err)
	}

	// Заявка на ремонт → «В ремонт» по заявке → «Из ремонта»
	req, err := e.requests.Create(ctx, memberWorker, assetID, "не включается", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	due := time.Now().Add(-time.Hour) // уже просрочен
	if _, err := e.lifecycle.StartRepair(ctx, memberMolB, assetID, RepairStartInput{Kind: models.RepairKindRepair, ExpectedReturnAt: &due, RequestID: &req.UUID}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.lifecycle.StartRepair(ctx, admin, assetID, RepairStartInput{}); statusOf(err) != http.StatusConflict {
		t.Fatalf("второй ремонт: %v", err)
	}
	overdue, _, _ := e.assets.List(ctx, admin, AssetFilter{Overdue: true}, 1, 10)
	if len(overdue) != 1 {
		t.Fatalf("просроченный ремонт: %d", len(overdue))
	}
	// передача другому МОЛ ремонт не закрывает
	if _, err := e.roles.AssignMol(ctx, admin, assetID, molA, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := e.lifecycle.FinishRepair(ctx, memberMol, assetID, RepairFinishInput{ResultCondition: models.ConditionOK}); err != nil {
		t.Fatal(err)
	}
	got, _ := e.assets.Get(ctx, admin, assetID)
	if got.Asset.AccountingStatus != models.AccountingInUse || got.Asset.Condition != models.ConditionOK || got.OpenRepair != nil {
		t.Fatalf("после ремонта: %s %s %v", got.Asset.AccountingStatus, got.Asset.Condition, got.OpenRepair)
	}
	r, _ := e.requests.Get(ctx, admin, req.UUID)
	if r.Status != models.RequestDone {
		t.Fatalf("заявка после ремонта: %s", r.Status)
	}

	// «Мне назначено» → отказ → «На разбор» → разобрано
	mine, err := e.roles.MyAssignments(ctx, memberWorker)
	if err != nil || len(mine) != 1 {
		t.Fatalf("Мне назначено: %v %d", err, len(mine))
	}
	if _, err := e.roles.Ack(ctx, memberWorker, mine[0].ID, false, ""); statusOf(err) != http.StatusBadRequest {
		t.Fatalf("отказ без причины: %v", err)
	}
	if _, err := e.roles.Ack(ctx, memberWorker, mine[0].ID, false, "не пользуюсь"); err != nil {
		t.Fatal(err)
	}
	disputes, _ := e.roles.Disputes(ctx, admin)
	if len(disputes) != 1 {
		t.Fatalf("на разбор: %d", len(disputes))
	}
	if _, err := e.roles.Resolve(ctx, admin, disputes[0].ID, "оставить"); err != nil {
		t.Fatal(err)
	}
	if d, _ := e.roles.Disputes(ctx, admin); len(d) != 0 {
		t.Fatalf("после разбора: %d", len(d))
	}

	// Утерян — блокирует увольнение МОЛ; передать может только админ
	lost := models.AccountingLost
	if err := e.lifecycle.ChangeStatus(ctx, memberMol, assetID, StatusInput{AccountingStatus: &lost}); err != nil {
		t.Fatal(err)
	}
	if n, _, _ := e.roles.MolCount(ctx, molA); n != 1 {
		t.Fatalf("утерянный в счёт: %d", n)
	}
	if _, err := e.roles.AssignMol(ctx, memberMol, assetID, molB, ""); statusOf(err) != http.StatusBadRequest {
		t.Fatalf("утерянный МОЛ не передаёт: %v", err)
	}

	// Списать может только админ; списание закрывает роли
	dec := models.AccountingDecommissioned
	if err := e.lifecycle.ChangeStatus(ctx, memberMol, assetID, StatusInput{AccountingStatus: &dec}); statusOf(err) != http.StatusForbidden {
		t.Fatalf("МОЛ не списывает: %v", err)
	}
	if err := e.lifecycle.ChangeStatus(ctx, admin, assetID, StatusInput{AccountingStatus: &dec}); err != nil {
		t.Fatal(err)
	}
	if n, u, _ := e.roles.MolCount(ctx, molA); n != 0 || u != 0 {
		t.Fatalf("после списания роли закрыты: mol=%d", n)
	}
	if n, _, _ := e.roles.MolCount(ctx, worker); n != 0 {
		t.Fatal("работник")
	}
	if _, err := e.roles.AddUser(ctx, admin, assetID, worker, ""); statusOf(err) != http.StatusBadRequest {
		t.Fatalf("на списанный не назначить: %v", err)
	}
	// списанные не попадают в «Без МОЛ»
	if l, _, _ := e.assets.List(ctx, admin, AssetFilter{NoMol: true}, 1, 10); len(l) != 0 {
		t.Fatalf("Без МОЛ после списания: %d", len(l))
	}

	// История и журнал
	h, err := e.lifecycle.History(ctx, admin, assetID)
	if err != nil || len(h.Status) < 4 || len(h.Repairs) != 1 {
		t.Fatalf("история: %v %d %d", err, len(h.Status), len(h.Repairs))
	}
	transfers, roles, err := e.roles.Transfers(ctx, admin, assetID)
	if err != nil || len(transfers) < 5 || len(roles) < 4 {
		t.Fatalf("передачи: %v %d %d", err, len(transfers), len(roles))
	}
	var auditCount int64
	e.db.Model(&models.AuditLog{}).Where("organization_id = ?", e.org).Count(&auditCount)
	if auditCount == 0 {
		t.Fatal("журнал пуст")
	}
}

func TestIntegrationVisibilityAndDismiss(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	adminPerson := e.person(true)
	admin := e.principal(adminPerson, adminPerms...)
	mol := e.person(true)
	worker := e.person(false)
	stranger := e.person(false)
	memberStranger := e.principal(stranger, memberPerms...)
	memberWorker := e.principal(worker, memberPerms...)

	v, err := e.assets.Create(ctx, admin, CreateAssetInput{Asset: &models.Asset{Name: "Ноутбук", Category: "ноутбук", InventoryNum: strPtr("INV-1")}, MolPersonID: &mol})
	if err != nil {
		t.Fatal(err)
	}
	if v.Asset.ModerationStatus != models.ModerationApproved || v.Mol == nil {
		t.Fatal("админ создаёт сразу одобренную карточку с МОЛ")
	}
	// инвентарный номер уникален в организации
	if _, err := e.assets.Create(ctx, admin, CreateAssetInput{Asset: &models.Asset{Name: "Дубль", Category: "ноутбук", InventoryNum: strPtr("INV-1")}, NoMol: true}); statusOf(err) != http.StatusConflict {
		t.Fatalf("дубль инв. номера: %v", err)
	}
	if _, err := e.roles.AddUser(ctx, admin, v.Asset.UUID, worker, ""); err != nil {
		t.Fatal(err)
	}

	// Посторонний участник не видит чужой актив
	if _, err := e.assets.Get(ctx, memberStranger, v.Asset.UUID); statusOf(err) != http.StatusForbidden {
		t.Fatalf("посторонний видит: %v", err)
	}
	if l, _, _ := e.assets.List(ctx, memberStranger, AssetFilter{}, 1, 10); len(l) != 0 {
		t.Fatalf("посторонний в списке: %d", len(l))
	}
	if l, _, _ := e.assets.List(ctx, memberWorker, AssetFilter{}, 1, 10); len(l) != 1 {
		t.Fatalf("работающий видит свой: %d", len(l))
	}
	// Другая организация не видит вовсе
	other := &authz.Principal{AccountID: uuid.New(), PersonID: uuid.New(), OrgID: uuid.New(), Perms: map[string]bool{authz.PermAssetsViewAll: true}}
	if _, err := e.assets.Get(ctx, other, v.Asset.UUID); statusOf(err) != http.StatusNotFound {
		t.Fatalf("чужая организация: %v", err)
	}

	// Увольнение работника закрывает его роли «работает с активом»
	n, err := e.roles.PersonDismissed(ctx, worker)
	if err != nil || n != 1 {
		t.Fatalf("увольнение: %v %d", err, n)
	}
	got, _ := e.assets.Get(ctx, admin, v.Asset.UUID)
	if len(got.Users) != 0 {
		t.Fatal("роль работника должна закрыться")
	}

	// Окончательное удаление организации
	if err := e.orgs.DeleteOrganizationData(ctx, e.org); err != nil {
		t.Fatal(err)
	}
	var left int64
	e.db.Model(&models.Asset{}).Unscoped().Where("organization_id = ?", e.org).Count(&left)
	if left != 0 {
		t.Fatalf("после удаления организации осталось активов: %d", left)
	}
}
