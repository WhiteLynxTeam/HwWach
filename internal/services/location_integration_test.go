package services

import (
	"HwWach/internal/dto"
	"HwWach/internal/models"
	"context"
	"testing"

	"github.com/google/uuid"
)

// Места: цепочка заводится один раз, повтор с другим id и другим регистром находит те же узлы,
// фильтр по верхнему уровню находит вложенные, правка меняет место и набор фото
func TestIntegrationLocations(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	admin := e.principal(e.person(true), adminPerms...)
	locs := NewLocationService(e.db)

	phoneID := uuid.New()
	chain := []LocationNode{{Name: "Каленковичи"}, {Name: "Ферма  700"}, {ID: &phoneID, Name: "Кабинет зав. фермы"}}
	room, err := locs.Create(ctx, admin, chain)
	if err != nil {
		t.Fatal(err)
	}
	if room.ID != phoneID || room.Path != "Каленковичи / Ферма 700 / Кабинет зав. фермы" || room.Kind != models.LocationRoom {
		t.Fatalf("room: %+v", room)
	}

	// второй телефон без сети завёл те же места со своими id
	other := uuid.New()
	again, err := locs.Create(ctx, admin, []LocationNode{{ID: &other, Name: "каленковичи"}, {Name: "ферма 700"}, {Name: "КАБИНЕТ ЗАВ. ФЕРМЫ"}})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != room.ID {
		t.Fatalf("duplicate location created: %s vs %s", again.ID, room.ID)
	}
	all, _ := locs.List(ctx, admin, "")
	if len(all) != 3 {
		t.Fatalf("want 3 locations, got %d", len(all))
	}

	a, err := e.assets.Create(ctx, admin, CreateAssetInput{
		Asset:         &models.Asset{Name: "Системный блок", Category: "ПК"},
		LocationChain: []LocationNode{{ID: &room.ID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.Asset.LocationKind != models.AssetLocationPlace || a.Asset.Location == nil || a.Asset.Location.ID != room.ID {
		t.Fatalf("asset location: %+v", a.Asset)
	}
	// повтор отправки с тем же client_id — та же карточка, а не ошибка
	cid := uuid.New()
	first, err := e.assets.Create(ctx, admin, CreateAssetInput{Asset: &models.Asset{Name: "Монитор", Category: "Монитор", ClientID: &cid}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.assets.Create(ctx, admin, CreateAssetInput{Asset: &models.Asset{Name: "Монитор", Category: "Монитор", ClientID: &cid}})
	if err != nil || second.Asset.UUID != first.Asset.UUID {
		t.Fatalf("retry with same client_id: %v", err)
	}

	phone, err := e.assets.Create(ctx, admin, CreateAssetInput{
		Asset: &models.Asset{Name: "Смартфон", Category: "Телефон"}, LocationKind: models.AssetLocationPortable,
	})
	if err != nil {
		t.Fatal(err)
	}
	if phone.Asset.LocationID != nil {
		t.Fatal("portable asset must not have a place")
	}
	if _, err := e.assets.Create(ctx, admin, CreateAssetInput{
		Asset: &models.Asset{Name: "Х", Category: "Х"}, LocationKind: models.AssetLocationPlace,
	}); statusOf(err) != 400 {
		t.Fatalf("place without chain: %v", err)
	}

	town := all[0]
	list, total, err := e.assets.List(ctx, admin, AssetFilter{LocationID: &town.ID}, 1, 100)
	if err != nil || total != 1 || list[0].Asset.UUID != a.Asset.UUID || list[0].Asset.Location == nil {
		t.Fatalf("filter by town: total=%d err=%v", total, err)
	}
	if _, total, _ := e.assets.List(ctx, admin, AssetFilter{Query: "ферма 700"}, 1, 100); total != 1 {
		t.Fatalf("search by path: %d", total)
	}

	// правка: место «нет», название
	none := models.AssetLocationNone
	name := "Системный блок Intel"
	upd, err := e.assets.Update(ctx, admin, a.Asset.UUID, UpdateAssetInput{Req: &dto.UpdateAssetRequest{Name: &name}, LocationKind: &none})
	if err != nil {
		t.Fatal(err)
	}
	if upd.Asset.Name != name || upd.Asset.LocationKind != models.AssetLocationNone || upd.Asset.LocationID != nil {
		t.Fatalf("update: %+v", upd.Asset)
	}
	empty := []uuid.UUID{}
	if _, err := e.assets.Update(ctx, admin, a.Asset.UUID, UpdateAssetInput{PhotoClientIDs: &empty}); err != nil {
		t.Fatal(err)
	}

	// удаление организации чистит и места
	if err := e.orgs.DeleteOrganizationData(ctx, e.org); err != nil {
		t.Fatal(err)
	}
	if left, _ := locs.List(ctx, admin, ""); len(left) != 0 {
		t.Fatalf("locations left: %d", len(left))
	}
}
