// library_material_service.go 负责书库素材的分页聚合、自有素材读写和移除书籍前的素材迁移。
package library

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// OpenLibraryMaterialPage 聚合书库自有素材和仍在书库中的单书素材引用，并按请求分页返回。
func OpenLibraryMaterialPage(request LibraryMaterialPageRequest) (LibraryMaterialPage, error) {
	libraryPath, index, err := resolveLibraryMaterialContext(request.LibraryPath)
	if err != nil {
		return LibraryMaterialPage{}, err
	}
	ownMaterials, err := loadLibraryMaterials(libraryPath)
	if err != nil {
		return LibraryMaterialPage{}, err
	}
	allItems, err := buildLibraryMaterialItems(libraryPath, index, ownMaterials)
	if err != nil {
		return LibraryMaterialPage{}, err
	}
	sort.SliceStable(allItems, func(i, j int) bool {
		return libraryMaterialTimeValue(allItems[i]) > libraryMaterialTimeValue(allItems[j])
	})

	books := collectLibraryMaterialBooks(allItems)
	types := collectLibraryMaterialStrings(allItems, func(item LibraryMaterialListItem) []string {
		return []string{item.MaterialType}
	})
	tags := collectLibraryMaterialStrings(allItems, func(item LibraryMaterialListItem) []string {
		return item.TagIDs
	})
	filtered := filterLibraryMaterialItems(allItems, request)
	page, pageSize := normalizeLibraryMaterialPagination(request.Page, request.PageSize)
	total := len(filtered)
	if total > 0 {
		maxPage := (total + pageSize - 1) / pageSize
		if page > maxPage {
			page = maxPage
		}
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return LibraryMaterialPage{
		LibraryPath: libraryPath,
		Items:       filtered[start:end],
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		Books:       books,
		Types:       types,
		Tags:        tags,
	}, nil
}

// CreateLibraryMaterial 在书库目录中直接创建一条自有素材，不依赖任何单本小说。
func CreateLibraryMaterial(request CreateLibraryMaterialRequest) (LibraryMaterialListItem, error) {
	libraryPath, _, err := resolveLibraryMaterialContext(request.LibraryPath)
	if err != nil {
		return LibraryMaterialListItem{}, err
	}
	title, summary, content, err := normalizeMaterialText(request.Title, request.Summary, request.Content)
	if err != nil {
		return LibraryMaterialListItem{}, err
	}
	materials, err := loadLibraryMaterials(libraryPath)
	if err != nil {
		return LibraryMaterialListItem{}, err
	}
	materialType := strings.TrimSpace(request.MaterialType)
	if materialType == "" {
		materialType = analysisKindWritingTechnique
	}
	now := time.Now().Format(time.RFC3339)
	materialID := nextLibraryMaterialID(materials)
	material := LibraryMaterial{
		ID:            materialID,
		SourceMode:    libraryMaterialSourceOwned,
		MaterialType:  materialType,
		PlanTypeName:  strings.TrimSpace(request.PlanTypeName),
		DimensionName: strings.TrimSpace(request.DimensionName),
		Title:         title,
		Summary:       summary,
		ContentPath:   filepath.ToSlash(filepath.Join(libraryMaterialTextsDir, materialID+".txt")),
		TagIDs:        normalizeTags(request.TagIDs),
		SourceRefs:    []MaterialSourceRef{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := writeLibraryMaterialContent(libraryPath, material, content); err != nil {
		return LibraryMaterialListItem{}, err
	}
	materials.Items = append(materials.Items, material)
	if err := saveLibraryMaterials(libraryPath, materials); err != nil {
		return LibraryMaterialListItem{}, err
	}
	return libraryOwnedMaterialToItem(material, content), nil
}

// UpdateLibraryMaterial 更新书库自有素材或已迁移素材；实时引用的单书素材不能通过这里改写。
func UpdateLibraryMaterial(request UpdateLibraryMaterialRequest) (LibraryMaterialListItem, error) {
	libraryPath, _, err := resolveLibraryMaterialContext(request.LibraryPath)
	if err != nil {
		return LibraryMaterialListItem{}, err
	}
	title, summary, content, err := normalizeMaterialText(request.Title, request.Summary, request.Content)
	if err != nil {
		return LibraryMaterialListItem{}, err
	}
	materials, err := loadLibraryMaterials(libraryPath)
	if err != nil {
		return LibraryMaterialListItem{}, err
	}
	for index := range materials.Items {
		if materials.Items[index].ID != request.LibraryMaterialID {
			continue
		}
		materialType := strings.TrimSpace(request.MaterialType)
		if materialType == "" {
			materialType = analysisKindWritingTechnique
		}
		materials.Items[index].MaterialType = materialType
		materials.Items[index].PlanTypeName = strings.TrimSpace(request.PlanTypeName)
		materials.Items[index].DimensionName = strings.TrimSpace(request.DimensionName)
		materials.Items[index].Title = title
		materials.Items[index].Summary = summary
		materials.Items[index].TagIDs = normalizeTags(request.TagIDs)
		materials.Items[index].UpdatedAt = time.Now().Format(time.RFC3339)
		if err := writeLibraryMaterialContent(libraryPath, materials.Items[index], content); err != nil {
			return LibraryMaterialListItem{}, err
		}
		if err := saveLibraryMaterials(libraryPath, materials); err != nil {
			return LibraryMaterialListItem{}, err
		}
		return libraryOwnedMaterialToItem(materials.Items[index], content), nil
	}
	return LibraryMaterialListItem{}, fmt.Errorf("书库素材不存在: %s", request.LibraryMaterialID)
}

// DeleteLibraryMaterial 删除书库正式数据中的素材；本书素材引用必须在单书素材页删除。
func DeleteLibraryMaterial(request DeleteLibraryMaterialRequest) (LibraryMaterialCollection, error) {
	libraryPath, _, err := resolveLibraryMaterialContext(request.LibraryPath)
	if err != nil {
		return LibraryMaterialCollection{}, err
	}
	materials, err := loadLibraryMaterials(libraryPath)
	if err != nil {
		return LibraryMaterialCollection{}, err
	}
	removed := LibraryMaterial{}
	kept := make([]LibraryMaterial, 0, len(materials.Items))
	for _, material := range materials.Items {
		if material.ID == request.LibraryMaterialID {
			removed = material
			continue
		}
		kept = append(kept, material)
	}
	if removed.ID == "" {
		return LibraryMaterialCollection{}, fmt.Errorf("书库素材不存在: %s", request.LibraryMaterialID)
	}
	materials.Items = kept
	if err := saveLibraryMaterials(libraryPath, materials); err != nil {
		return LibraryMaterialCollection{}, err
	}
	if strings.TrimSpace(removed.ContentPath) != "" {
		if contentPath, pathErr := safeProjectPath(libraryPath, removed.ContentPath); pathErr == nil {
			if err := os.Remove(contentPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return LibraryMaterialCollection{}, fmt.Errorf("删除书库素材正文失败: %w", err)
			}
		}
	}
	return materials, nil
}

// MigrateBookMaterialsToLibrary 在删除单书项目前，把该书已有素材复制为书库正式素材。
func MigrateBookMaterialsToLibrary(libraryPath string, bookPath string) (int, error) {
	libraryPath, _, err := resolveLibraryMaterialContext(libraryPath)
	if err != nil {
		return 0, err
	}
	project, err := OpenBookProject(bookPath)
	if err != nil {
		return 0, err
	}
	bookPath = projectRootPath(bookPath)
	bookMaterials, err := loadMaterialsIfExists(bookPath)
	if err != nil {
		return 0, err
	}
	if len(bookMaterials.Items) == 0 {
		return 0, nil
	}
	contents, err := loadMaterialContents(bookPath, bookMaterials)
	if err != nil {
		return 0, err
	}
	libraryMaterials, err := loadLibraryMaterials(libraryPath)
	if err != nil {
		return 0, err
	}

	existing := map[string]bool{}
	for _, material := range libraryMaterials.Items {
		if material.BookID == "" || material.BookMaterialID == "" {
			continue
		}
		existing[material.BookID+"\x00"+material.BookMaterialID] = true
	}

	migrated := 0
	now := time.Now().Format(time.RFC3339)
	for _, bookMaterial := range bookMaterials.Items {
		key := project.Book.ID + "\x00" + bookMaterial.ID
		if existing[key] {
			continue
		}
		planTypeName, dimensionName := materialSourceNames(bookMaterial)
		materialID := nextLibraryMaterialID(libraryMaterials)
		createdAt := strings.TrimSpace(bookMaterial.CreatedAt)
		if createdAt == "" {
			createdAt = now
		}
		material := LibraryMaterial{
			ID:               materialID,
			SourceMode:       libraryMaterialSourceMigrated,
			BookID:           project.Book.ID,
			BookName:         project.Book.Name,
			OriginalBookPath: bookPath,
			BookMaterialID:   bookMaterial.ID,
			MaterialType:     bookMaterial.MaterialType,
			PlanTypeName:     planTypeName,
			DimensionName:    dimensionName,
			Title:            bookMaterial.Title,
			Summary:          bookMaterial.Summary,
			ContentPath:      filepath.ToSlash(filepath.Join(libraryMaterialTextsDir, materialID+".txt")),
			TagIDs:           append([]string{}, bookMaterial.TagIDs...),
			SourceRefs:       append([]MaterialSourceRef{}, bookMaterial.SourceRefs...),
			CreatedAt:        createdAt,
			UpdatedAt:        now,
			MigratedAt:       now,
		}
		if err := writeLibraryMaterialContent(libraryPath, material, contents[bookMaterial.ID]); err != nil {
			return migrated, err
		}
		libraryMaterials.Items = append(libraryMaterials.Items, material)
		existing[key] = true
		migrated++
	}
	if migrated == 0 {
		return 0, nil
	}
	if err := saveLibraryMaterials(libraryPath, libraryMaterials); err != nil {
		return migrated, err
	}
	return migrated, nil
}

func resolveLibraryMaterialContext(libraryPath string) (string, LibraryIndex, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(libraryPath))
	if cleanPath == "." || cleanPath == "" {
		state, err := LoadLibraryState()
		if err != nil {
			return "", LibraryIndex{}, err
		}
		cleanPath = filepath.Clean(strings.TrimSpace(state.LibraryPath))
	}
	if cleanPath == "." || cleanPath == "" {
		return "", LibraryIndex{}, errors.New("请先设置书库目录")
	}
	index, err := ensureLibraryIndex(cleanPath)
	if err != nil {
		return "", LibraryIndex{}, err
	}
	return cleanPath, enrichLibraryReadingProgress(index), nil
}

func loadLibraryMaterials(libraryPath string) (LibraryMaterialCollection, error) {
	collection := LibraryMaterialCollection{Version: 1, Items: []LibraryMaterial{}}
	path := filepath.Join(libraryPath, filepath.FromSlash(libraryMaterialsRelativePath))
	if err := readJSONFile(path, &collection); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return collection, nil
		}
		return collection, fmt.Errorf("读取书库素材失败: %w", err)
	}
	if collection.Version == 0 {
		collection.Version = 1
	}
	if collection.Items == nil {
		collection.Items = []LibraryMaterial{}
		return collection, nil
	}
	for index := range collection.Items {
		if strings.TrimSpace(collection.Items[index].SourceMode) == "" {
			collection.Items[index].SourceMode = libraryMaterialSourceOwned
		}
		if collection.Items[index].TagIDs == nil {
			collection.Items[index].TagIDs = []string{}
		}
		if collection.Items[index].SourceRefs == nil {
			collection.Items[index].SourceRefs = []MaterialSourceRef{}
		}
	}
	return collection, nil
}

func saveLibraryMaterials(libraryPath string, materials LibraryMaterialCollection) error {
	if materials.Items == nil {
		materials.Items = []LibraryMaterial{}
	}
	materials.Version = 1
	return writeJSONFileAtomic(filepath.Join(libraryPath, filepath.FromSlash(libraryMaterialsRelativePath)), materials)
}

func loadMaterialsIfExists(bookPath string) (MaterialCollection, error) {
	collection := MaterialCollection{Version: 1, Items: []Material{}}
	path := filepath.Join(bookPath, filepath.FromSlash(materialsRelativePath))
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return collection, nil
	} else if err != nil {
		return collection, fmt.Errorf("检查素材库失败: %w", err)
	}
	return loadMaterials(bookPath)
}

func buildLibraryMaterialItems(libraryPath string, index LibraryIndex, ownMaterials LibraryMaterialCollection) ([]LibraryMaterialListItem, error) {
	items := make([]LibraryMaterialListItem, 0, len(ownMaterials.Items))
	for _, material := range ownMaterials.Items {
		content, err := readLibraryMaterialContent(libraryPath, material)
		if err != nil {
			return items, err
		}
		items = append(items, libraryOwnedMaterialToItem(material, content))
	}
	for _, entry := range index.Books {
		project, err := OpenBookProject(entry.BookPath)
		if err != nil {
			continue
		}
		bookPath := projectRootPath(project.Book.BookPath)
		materials, err := loadMaterialsIfExists(bookPath)
		if err != nil {
			continue
		}
		contents, err := loadMaterialContents(bookPath, materials)
		if err != nil {
			continue
		}
		for _, material := range materials.Items {
			planTypeName, dimensionName := materialSourceNames(material)
			items = append(items, LibraryMaterialListItem{
				ID:                "book:" + project.Book.ID + ":" + material.ID,
				SourceMode:        libraryMaterialSourceReference,
				LibraryMaterialID: "",
				BookID:            project.Book.ID,
				BookName:          project.Book.Name,
				BookPath:          bookPath,
				BookMaterialID:    material.ID,
				MaterialType:      material.MaterialType,
				PlanTypeName:      planTypeName,
				DimensionName:     dimensionName,
				Title:             material.Title,
				Summary:           material.Summary,
				Content:           contents[material.ID],
				TagIDs:            append([]string{}, material.TagIDs...),
				SourceRefs:        append([]MaterialSourceRef{}, material.SourceRefs...),
				CreatedAt:         material.CreatedAt,
				UpdatedAt:         material.UpdatedAt,
				SourceAvailable:   true,
			})
		}
	}
	return items, nil
}

func libraryOwnedMaterialToItem(material LibraryMaterial, content string) LibraryMaterialListItem {
	sourceAvailable := false
	if strings.TrimSpace(material.OriginalBookPath) != "" {
		_, err := os.Stat(material.OriginalBookPath)
		sourceAvailable = err == nil
	}
	return LibraryMaterialListItem{
		ID:                "library:" + material.ID,
		SourceMode:        material.SourceMode,
		LibraryMaterialID: material.ID,
		BookID:            material.BookID,
		BookName:          material.BookName,
		BookPath:          material.OriginalBookPath,
		BookMaterialID:    material.BookMaterialID,
		MaterialType:      material.MaterialType,
		PlanTypeName:      material.PlanTypeName,
		DimensionName:     material.DimensionName,
		Title:             material.Title,
		Summary:           material.Summary,
		Content:           content,
		TagIDs:            append([]string{}, material.TagIDs...),
		SourceRefs:        append([]MaterialSourceRef{}, material.SourceRefs...),
		CreatedAt:         material.CreatedAt,
		UpdatedAt:         material.UpdatedAt,
		SourceAvailable:   sourceAvailable,
	}
}

func readLibraryMaterialContent(libraryPath string, material LibraryMaterial) (string, error) {
	if strings.TrimSpace(material.ContentPath) == "" {
		return "", nil
	}
	path, err := safeProjectPath(libraryPath, material.ContentPath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("读取书库素材正文失败: %w", err)
	}
	return string(trimUTF8BOM(content)), nil
}

func writeLibraryMaterialContent(libraryPath string, material LibraryMaterial, content string) error {
	path, err := safeProjectPath(libraryPath, material.ContentPath)
	if err != nil {
		return err
	}
	if err := writeTextFileAtomic(path, content); err != nil {
		return fmt.Errorf("保存书库素材正文失败: %w", err)
	}
	return nil
}

func filterLibraryMaterialItems(items []LibraryMaterialListItem, request LibraryMaterialPageRequest) []LibraryMaterialListItem {
	keyword := strings.ToLower(strings.TrimSpace(request.Keyword))
	bookID := strings.TrimSpace(request.BookID)
	materialType := strings.TrimSpace(request.MaterialType)
	tagID := strings.TrimSpace(request.TagID)
	filtered := make([]LibraryMaterialListItem, 0, len(items))
	for _, item := range items {
		if bookID != "" && item.BookID != bookID {
			continue
		}
		if materialType != "" && item.MaterialType != materialType {
			continue
		}
		if tagID != "" && !containsString(item.TagIDs, tagID) {
			continue
		}
		if keyword != "" && !libraryMaterialKeywordHit(item, keyword) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func libraryMaterialKeywordHit(item LibraryMaterialListItem, keyword string) bool {
	values := []string{
		item.BookName,
		item.Title,
		item.Summary,
		item.Content,
		item.MaterialType,
		item.PlanTypeName,
		item.DimensionName,
	}
	for _, source := range item.SourceRefs {
		values = append(values,
			source.Snapshot.BookName,
			source.Snapshot.ChapterTitle,
			source.Snapshot.DeconstructionTitle,
			source.Snapshot.SourceSummary,
			source.Snapshot.SelectedText,
		)
	}
	values = append(values, item.TagIDs...)
	return strings.Contains(strings.ToLower(strings.Join(values, " ")), keyword)
}

func collectLibraryMaterialBooks(items []LibraryMaterialListItem) []LibraryMaterialBookOption {
	seen := map[string]string{}
	for _, item := range items {
		if strings.TrimSpace(item.BookID) == "" {
			continue
		}
		bookName := strings.TrimSpace(item.BookName)
		if bookName == "" {
			bookName = item.BookID
		}
		seen[item.BookID] = bookName
	}
	books := make([]LibraryMaterialBookOption, 0, len(seen))
	for bookID, bookName := range seen {
		books = append(books, LibraryMaterialBookOption{BookID: bookID, BookName: bookName})
	}
	sort.SliceStable(books, func(i, j int) bool {
		return strings.ToLower(books[i].BookName) < strings.ToLower(books[j].BookName)
	})
	return books
}

func collectLibraryMaterialStrings(items []LibraryMaterialListItem, picker func(LibraryMaterialListItem) []string) []string {
	seen := map[string]bool{}
	values := []string{}
	for _, item := range items {
		for _, value := range picker(item) {
			clean := strings.TrimSpace(value)
			if clean == "" || seen[clean] {
				continue
			}
			seen[clean] = true
			values = append(values, clean)
		}
	}
	sort.SliceStable(values, func(i, j int) bool {
		return strings.ToLower(values[i]) < strings.ToLower(values[j])
	})
	return values
}

func normalizeLibraryMaterialPagination(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func materialSourceNames(material Material) (string, string) {
	source := material.SourceRefs
	if len(source) == 0 {
		return "", ""
	}
	planTypeName := strings.TrimSpace(source[0].Snapshot.PlanTypeName)
	dimensionName := strings.TrimSpace(source[0].Snapshot.DimensionName)
	if planTypeName == "" || dimensionName == "" {
		parts := strings.Split(source[0].Snapshot.DeconstructionTitle, "/")
		if planTypeName == "" && len(parts) > 0 {
			planTypeName = strings.TrimSpace(parts[0])
		}
		if dimensionName == "" && len(parts) > 1 {
			dimensionName = strings.TrimSpace(parts[1])
		}
	}
	return planTypeName, dimensionName
}

func libraryMaterialTimeValue(item LibraryMaterialListItem) int64 {
	value, err := time.Parse(time.RFC3339, item.CreatedAt)
	if err != nil {
		return 0
	}
	return value.Unix()
}

func nextLibraryMaterialID(materials LibraryMaterialCollection) string {
	existing := map[string]bool{}
	for _, material := range materials.Items {
		existing[material.ID] = true
	}
	for next := len(materials.Items) + 1; ; next++ {
		id := nextSequentialID("library-material", next)
		if !existing[id] {
			return id
		}
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
