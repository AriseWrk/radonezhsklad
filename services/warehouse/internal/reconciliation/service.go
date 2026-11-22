package reconciliation

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/radonezhsklad/warehouse/internal/models"
	"github.com/radonezhsklad/warehouse/internal/product"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

const salePriceTypeName = "Цена продажи"

type Service struct {
	ms            *MSClient
	products      *product.Client
	repo          *repository.Repo
	runs          *repository.ReconciliationRepo
	internalToken string
}

func NewService(
	ms *MSClient,
	products *product.Client,
	repo *repository.Repo,
	runs *repository.ReconciliationRepo,
	internalToken string,
) *Service {
	return &Service{
		ms:            ms,
		products:      products,
		repo:          repo,
		runs:          runs,
		internalToken: strings.TrimSpace(internalToken),
	}
}

type productMatch struct {
	Local    product.Product
	MoySklad MSProduct
	Method   string
}

type productMatchResult struct {
	Matches        []productMatch
	AmbiguousLocal map[uuid.UUID]string
	AmbiguousMS    map[uuid.UUID]string
	MissingLocalMS []product.Product
	MissingMSLocal []MSProduct
}

func normalizeSKU(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func matchProducts(local []product.Product, remote []MSProduct) productMatchResult {
	result := productMatchResult{
		Matches:        make([]productMatch, 0),
		AmbiguousLocal: make(map[uuid.UUID]string),
		AmbiguousMS:    make(map[uuid.UUID]string),
	}

	localByExternal := make(map[uuid.UUID][]int)
	msByID := make(map[uuid.UUID]int)

	for i := range local {
		if local[i].ExternalID != nil && *local[i].ExternalID != uuid.Nil {
			id := *local[i].ExternalID
			localByExternal[id] = append(localByExternal[id], i)
		}
	}
	for i := range remote {
		msByID[remote[i].ID] = i
	}

	usedLocal := make(map[int]bool)
	usedMS := make(map[int]bool)

	// Первый приоритет — UUID МойСклад в local external_id.
	for externalID, indexes := range localByExternal {
		msIndex, exists := msByID[externalID]
		if !exists {
			continue
		}
		if len(indexes) != 1 {
			for _, localIndex := range indexes {
				result.AmbiguousLocal[local[localIndex].ID] =
					"несколько локальных товаров имеют одинаковый external_id"
				usedLocal[localIndex] = true
			}
			result.AmbiguousMS[remote[msIndex].ID] =
				"несколько локальных товаров претендуют на этот товар МойСклад"
			usedMS[msIndex] = true
			continue
		}

		localIndex := indexes[0]
		result.Matches = append(result.Matches, productMatch{
			Local: local[localIndex], MoySklad: remote[msIndex], Method: "external_id",
		})
		usedLocal[localIndex] = true
		usedMS[msIndex] = true
	}

	// Второй приоритет — уникальное совпадение SKU и article.
	localBySKU := make(map[string][]int)
	msByArticle := make(map[string][]int)

	for i := range local {
		if usedLocal[i] || result.AmbiguousLocal[local[i].ID] != "" || local[i].SKU == nil {
			continue
		}
		key := normalizeSKU(*local[i].SKU)
		if key != "" {
			localBySKU[key] = append(localBySKU[key], i)
		}
	}

	for i := range remote {
		if usedMS[i] || result.AmbiguousMS[remote[i].ID] != "" {
			continue
		}
		key := normalizeSKU(remote[i].Article)
		if key != "" {
			msByArticle[key] = append(msByArticle[key], i)
		}
	}

	for sku, localIndexes := range localBySKU {
		msIndexes := msByArticle[sku]
		if len(msIndexes) == 0 {
			continue
		}

		if len(localIndexes) != 1 || len(msIndexes) != 1 {
			reason := fmt.Sprintf(
				"неоднозначный SKU/article: локальных товаров %d, товаров МойСклад %d",
				len(localIndexes), len(msIndexes),
			)
			for _, i := range localIndexes {
				result.AmbiguousLocal[local[i].ID] = reason
				usedLocal[i] = true
			}
			for _, i := range msIndexes {
				result.AmbiguousMS[remote[i].ID] = reason
				usedMS[i] = true
			}
			continue
		}

		li, mi := localIndexes[0], msIndexes[0]
		result.Matches = append(result.Matches, productMatch{
			Local: local[li], MoySklad: remote[mi], Method: "sku",
		})
		usedLocal[li] = true
		usedMS[mi] = true
	}

	for i := range local {
		if !usedLocal[i] {
			result.MissingLocalMS = append(result.MissingLocalMS, local[i])
		}
	}
	for i := range remote {
		if !usedMS[i] {
			result.MissingMSLocal = append(result.MissingMSLocal, remote[i])
		}
	}

	return result
}

func (s *Service) loadSourceData(ctx context.Context) (
	[]product.Product,
	[]MSProduct,
	[]MSStore,
	[]MSPriceType,
	[]MSStockRow,
	error,
) {
	if s.ms == nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("клиент МойСклад не настроен")
	}
	if s.products == nil || s.repo == nil || s.runs == nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("не настроены зависимости сервиса сверки")
	}
	if s.internalToken == "" {
		return nil, nil, nil, nil, nil, fmt.Errorf("INTERNAL_TOKEN не настроен")
	}

	localProducts, err := s.products.ListCatalog(ctx, s.internalToken, false)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("загрузка локального каталога: %w", err)
	}

	msProducts, err := s.ms.ListProducts(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	msStores, err := s.ms.ListStores(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	priceTypes, err := s.ms.ListPriceTypes(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	stockRows, err := s.ms.ListStockByStore(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return localProducts, msProducts, msStores, priceTypes, stockRows, nil
}

func findSalePriceType(types []MSPriceType) (MSPriceType, error) {
	var found *MSPriceType

	for i := range types {
		if strings.TrimSpace(types[i].Name) != salePriceTypeName {
			continue
		}
		if found != nil {
			return MSPriceType{}, fmt.Errorf(
				"в МойСклад найдено несколько типов цены %q",
				salePriceTypeName,
			)
		}
		found = &types[i]
	}

	if found == nil {
		return MSPriceType{}, fmt.Errorf(
			"в МойСклад не найден тип цены %q",
			salePriceTypeName,
		)
	}

	return *found, nil
}

// warehouseMatch — однозначно сопоставленная пара складов.
type warehouseMatch struct {
	Local models.Warehouse
	MS    MSStore
}

type stockKey struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
}

func stringPtr(value string) *string {
	return &value
}

func floatPtr(value float64) *float64 {
	return &value
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}

// matchWarehouses сопоставляет склады только по external_id и UUID МойСклад.
func matchWarehouses(
	local []models.Warehouse,
	remote []MSStore,
) ([]warehouseMatch, []models.Warehouse, []MSStore) {
	remoteByID := make(map[uuid.UUID][]int)
	localByExternal := make(map[uuid.UUID][]int)

	for i := range remote {
		if remote[i].ID != uuid.Nil {
			remoteByID[remote[i].ID] = append(remoteByID[remote[i].ID], i)
		}
	}

	for i := range local {
		if local[i].ExternalID != nil && *local[i].ExternalID != uuid.Nil {
			id := *local[i].ExternalID
			localByExternal[id] = append(localByExternal[id], i)
		}
	}

	usedLocal := make(map[int]bool)
	usedRemote := make(map[int]bool)
	matches := make([]warehouseMatch, 0)
	unmatchedLocal := make([]models.Warehouse, 0)
	unmatchedRemote := make([]MSStore, 0)

	for externalID, localIndexes := range localByExternal {
		remoteIndexes := remoteByID[externalID]

		if len(remoteIndexes) != 1 || len(localIndexes) != 1 {
			for _, index := range localIndexes {
				usedLocal[index] = true
				unmatchedLocal = append(unmatchedLocal, local[index])
			}
			for _, index := range remoteIndexes {
				usedRemote[index] = true
				unmatchedRemote = append(unmatchedRemote, remote[index])
			}
			continue
		}

		li := localIndexes[0]
		ri := remoteIndexes[0]

		usedLocal[li] = true
		usedRemote[ri] = true

		matches = append(matches, warehouseMatch{
			Local: local[li],
			MS:    remote[ri],
		})
	}

	for i := range local {
		if !usedLocal[i] {
			unmatchedLocal = append(unmatchedLocal, local[i])
		}
	}

	for i := range remote {
		if !usedRemote[i] {
			unmatchedRemote = append(unmatchedRemote, remote[i])
		}
	}

	return matches, unmatchedLocal, unmatchedRemote
}

// salePriceForType возвращает цену только для точного типа "Цена продажи".
func salePriceForType(
	prices []MSSalePrice,
	priceType MSPriceType,
) (float64, bool, error) {
	var value float64
	count := 0

	for _, price := range prices {
		refID := price.PriceType.ID

		if refID == uuid.Nil && strings.TrimSpace(price.PriceType.Meta.Href) != "" {
			parsed, err := entityID(price.PriceType.Meta.Href)
			if err == nil {
				refID = parsed
			}
		}

		matched := false
		if refID != uuid.Nil {
			matched = refID == priceType.ID
		} else {
			matched = strings.TrimSpace(price.PriceType.Name) == salePriceTypeName
		}

		if matched {
			count++
			value = price.PriceInRubles()
		}
	}

	if count > 1 {
		return 0, false, fmt.Errorf(
			"у товара несколько цен типа %q",
			salePriceTypeName,
		)
	}

	if count == 0 {
		return 0, false, nil
	}

	return value, true, nil
}

// stockProductID получает UUID товара из отчёта остатков.
func stockProductID(row MSStockRow) (uuid.UUID, error) {
	if row.ID != uuid.Nil {
		return row.ID, nil
	}
	return entityID(row.Meta.Href)
}

// stockStoreID получает UUID склада из строки отчёта остатков.
func stockStoreID(row MSStoreStock) (uuid.UUID, error) {
	return entityID(row.Meta.Href)
}

func productSKU(local product.Product, remote MSProduct) *string {
	if local.SKU != nil && strings.TrimSpace(*local.SKU) != "" {
		return local.SKU
	}
	if strings.TrimSpace(remote.Article) != "" {
		return stringPtr(remote.Article)
	}
	return nil
}

func detailsJSON(value any) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

// Run выполняет одну сверку и сохраняет её результат в историю.
func (s *Service) Run(ctx context.Context, runID uuid.UUID) error {
	if runID == uuid.Nil {
		return fmt.Errorf("пустой ID запуска сверки")
	}

	counters := repository.ReconciliationCounters{}

	fail := func(cause error) error {
		finishErr := s.runs.FinishRun(
			ctx,
			runID,
			"failed",
			cause.Error(),
			counters,
		)
		if finishErr != nil {
			return fmt.Errorf("%w; ошибка сохранения статуса запуска: %v", cause, finishErr)
		}
		return cause
	}

	localProducts, msProducts, msStores, priceTypes, stockRows, err :=
		s.loadSourceData(ctx)
	if err != nil {
		return fail(err)
	}

	priceType, err := findSalePriceType(priceTypes)
	if err != nil {
		return fail(err)
	}

	warehouses, err := s.repo.ListWarehouses(ctx)
	if err != nil {
		return fail(fmt.Errorf("загрузка локальных складов: %w", err))
	}

	activeWarehouses := make([]models.Warehouse, 0)
	for _, warehouse := range warehouses {
		if warehouse.IsActive {
			activeWarehouses = append(activeWarehouses, warehouse)
		}
	}

	stockBalances, err := s.repo.ListStock(ctx, nil, nil)
	if err != nil {
		return fail(fmt.Errorf("загрузка локальных остатков: %w", err))
	}

	matches := matchProducts(localProducts, msProducts)
	warehouseMatches, unmatchedLocalWarehouses, unmatchedMSStores :=
		matchWarehouses(activeWarehouses, msStores)

	counters.TotalProducts = len(localProducts)
	counters.MatchedProducts = len(matches.Matches)
	counters.MissingMSProducts = len(matches.MissingLocalMS)
	counters.MissingLocalProducts = len(matches.MissingMSLocal)
	counters.AmbiguousProducts =
		len(matches.AmbiguousLocal) + len(matches.AmbiguousMS)
	counters.UnmatchedWarehouses =
		len(unmatchedLocalWarehouses) + len(unmatchedMSStores)

	items := make([]repository.ReconciliationItem, 0)

	// Индекс локальных остатков. Сравниваем физическое количество Quantity.
	localStock := make(map[stockKey]float64)
	for _, balance := range stockBalances {
		key := stockKey{
			ProductID:   balance.ProductID,
			WarehouseID: balance.WarehouseID,
		}
		localStock[key] += balance.Quantity
	}

	// Индекс остатков МойСклад по UUID товара и UUID склада.
	msStock := make(map[string]float64)
	for _, row := range stockRows {
		productID, idErr := stockProductID(row)
		if idErr != nil {
			return fail(fmt.Errorf("UUID товара в отчёте остатков МойСклад: %w", idErr))
		}

		for _, storeStock := range row.StockByStore {
			storeID, idErr := stockStoreID(storeStock)
			if idErr != nil {
				return fail(fmt.Errorf(
					"UUID склада в остатках товара %s: %w",
					productID,
					idErr,
				))
			}

			key := productID.String() + ":" + storeID.String()
			msStock[key] += storeStock.Stock
		}
	}

	// Индекс однозначно сопоставленных товаров по UUID МойСклад.
	matchedByMSID := make(map[uuid.UUID]productMatch)
	for _, match := range matches.Matches {
		matchedByMSID[match.MoySklad.ID] = match
	}

	// Строки по товарам, сопоставленным по external_id или SKU.
	for _, match := range matches.Matches {
		localID := match.Local.ID
		remoteID := match.MoySklad.ID
		remoteIDString := remoteID.String()

		items = append(items, repository.ReconciliationItem{
			ItemType:         "product",
			ResultStatus:     "matched",
			MatchMethod:      stringPtr(match.Method),
			LocalProductID:   &localID,
			MSProductID:      &remoteIDString,
			SKU:              productSKU(match.Local, match.MoySklad),
			LocalProductName: stringPtr(match.Local.Name),
			MSProductName:    stringPtr(match.MoySklad.Name),
		})

		msPrice, found, priceErr := salePriceForType(
			match.MoySklad.SalePrices,
			priceType,
		)
		if priceErr != nil {
			return fail(fmt.Errorf(
				"цена товара %q: %w",
				match.MoySklad.Name,
				priceErr,
			))
		}

		if !found {
			items = append(items, repository.ReconciliationItem{
				ItemType:         "price",
				ResultStatus:     "missing_moysklad",
				MatchMethod:      stringPtr(match.Method),
				LocalProductID:   &localID,
				MSProductID:      &remoteIDString,
				SKU:              productSKU(match.Local, match.MoySklad),
				LocalProductName: stringPtr(match.Local.Name),
				MSProductName:    stringPtr(match.MoySklad.Name),
				LocalPrice:       floatPtr(round4(match.Local.Price)),
				Details: detailsJSON(map[string]string{
					"reason": "не найдена цена типа Цена продажи",
				}),
			})
		} else {
			localPrice := round4(match.Local.Price)
			remotePrice := round4(msPrice)
			difference := round4(localPrice - remotePrice)

			status := "matched"
			if difference != 0 {
				status = "difference"
				counters.PriceDifferences++
			}

			items = append(items, repository.ReconciliationItem{
				ItemType:         "price",
				ResultStatus:     status,
				MatchMethod:      stringPtr(match.Method),
				LocalProductID:   &localID,
				MSProductID:      &remoteIDString,
				SKU:              productSKU(match.Local, match.MoySklad),
				LocalProductName: stringPtr(match.Local.Name),
				MSProductName:    stringPtr(match.MoySklad.Name),
				LocalPrice:       floatPtr(localPrice),
				MSPrice:          floatPtr(remotePrice),
				PriceDifference:  floatPtr(difference),
			})
		}
	}

	// Локальные товары, которым не найден товар МойСклад.
	for _, local := range matches.MissingLocalMS {
		localID := local.ID

		items = append(items, repository.ReconciliationItem{
			ItemType:         "product",
			ResultStatus:     "missing_moysklad",
			LocalProductID:   &localID,
			SKU:              local.SKU,
			LocalProductName: stringPtr(local.Name),
			LocalPrice:       floatPtr(round4(local.Price)),
		})
	}

	// Товары МойСклад, которым не найден локальный товар.
	for _, remote := range matches.MissingMSLocal {
		remoteID := remote.ID.String()

		items = append(items, repository.ReconciliationItem{
			ItemType:      "product",
			ResultStatus:  "missing_local",
			MSProductID:   &remoteID,
			SKU:           stringPtr(remote.Article),
			MSProductName: stringPtr(remote.Name),
		})
	}

	// Неоднозначные локальные товары.
	for localID, reason := range matches.AmbiguousLocal {
		var local *product.Product
		for i := range localProducts {
			if localProducts[i].ID == localID {
				local = &localProducts[i]
				break
			}
		}
		if local == nil {
			continue
		}

		items = append(items, repository.ReconciliationItem{
			ItemType:         "product",
			ResultStatus:     "ambiguous",
			LocalProductID:   &localID,
			SKU:              local.SKU,
			LocalProductName: stringPtr(local.Name),
			Details:          detailsJSON(map[string]string{"reason": reason}),
		})
	}

	// Неоднозначные товары МойСклад.
	for remoteID, reason := range matches.AmbiguousMS {
		var remote *MSProduct
		for i := range msProducts {
			if msProducts[i].ID == remoteID {
				remote = &msProducts[i]
				break
			}
		}
		if remote == nil {
			continue
		}

		remoteIDString := remoteID.String()

		items = append(items, repository.ReconciliationItem{
			ItemType:      "product",
			ResultStatus:  "ambiguous",
			MSProductID:   &remoteIDString,
			SKU:           stringPtr(remote.Article),
			MSProductName: stringPtr(remote.Name),
			Details:       detailsJSON(map[string]string{"reason": reason}),
		})
	}

	// Несопоставленные локальные склады.
	for _, warehouse := range unmatchedLocalWarehouses {
		warehouseID := warehouse.ID

		items = append(items, repository.ReconciliationItem{
			ItemType:           "warehouse",
			ResultStatus:       "unmatched",
			LocalWarehouseID:   &warehouseID,
			LocalWarehouseName: stringPtr(warehouse.Name),
			Details: detailsJSON(map[string]string{
				"reason": "не найден однозначный склад МойСклад по external_id",
			}),
		})
	}

	// Несопоставленные склады МойСклад.
	for _, store := range unmatchedMSStores {
		storeID := store.ID.String()

		items = append(items, repository.ReconciliationItem{
			ItemType:     "warehouse",
			ResultStatus: "unmatched",
			MSStoreID:    &storeID,
			MSStoreName:  stringPtr(store.Name),
			Details: detailsJSON(map[string]string{
				"reason": "не найден однозначный локальный склад по external_id",
			}),
		})
	}

	// Сравнение остатков для однозначно сопоставленных пар товар × склад.
	for _, match := range matches.Matches {
		for _, warehouseMatch := range warehouseMatches {
			localQuantity := round4(localStock[stockKey{
				ProductID:   match.Local.ID,
				WarehouseID: warehouseMatch.Local.ID,
			}])

			remoteQuantity := round4(msStock[match.MoySklad.ID.String()+":"+warehouseMatch.MS.ID.String()])

			difference := round4(localQuantity - remoteQuantity)
			status := "matched"

			if difference != 0 {
				status = "difference"
				counters.StockDifferences++
			}

			isZero := localQuantity == 0 || remoteQuantity == 0
			if isZero {
				counters.ZeroStockItems++
			}

			localProductID := match.Local.ID
			remoteProductID := match.MoySklad.ID.String()
			localWarehouseID := warehouseMatch.Local.ID
			remoteStoreID := warehouseMatch.MS.ID.String()

			items = append(items, repository.ReconciliationItem{
				ItemType:           "stock",
				ResultStatus:       status,
				MatchMethod:        stringPtr(match.Method),
				LocalProductID:     &localProductID,
				MSProductID:        &remoteProductID,
				SKU:                productSKU(match.Local, match.MoySklad),
				LocalProductName:   stringPtr(match.Local.Name),
				MSProductName:      stringPtr(match.MoySklad.Name),
				LocalWarehouseID:   &localWarehouseID,
				MSStoreID:          &remoteStoreID,
				LocalWarehouseName: stringPtr(warehouseMatch.Local.Name),
				MSStoreName:        stringPtr(warehouseMatch.MS.Name),
				LocalStock:         floatPtr(localQuantity),
				MSStock:            floatPtr(remoteQuantity),
				StockDifference:    floatPtr(difference),
				IsZeroStock:        isZero,
			})
		}
	}

	if err := s.runs.InsertItems(ctx, runID, items); err != nil {
		return fail(fmt.Errorf("сохранение строк сверки: %w", err))
	}

	if err := s.runs.FinishRun(
		ctx,
		runID,
		"completed",
		"",
		counters,
	); err != nil {
		return fmt.Errorf("сохранение итогов сверки: %w", err)
	}

	return nil
}
