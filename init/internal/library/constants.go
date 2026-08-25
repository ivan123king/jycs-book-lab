// constants.go 集中定义 library 包内的本地文件名、目录名和第一阶段业务枚举。
package library

const (
	appDataFolderName = "jiayi-book-lab"
	settingsFileName  = "settings.json"
	libraryIndexName  = "index.json"
	bookJSONName      = "book.json"
	originalTextName  = "original.txt"

	readingStateRelativePath        = "reading/reading-state.json"
	chapterMemosRelativePath        = "reading/chapter-memos.json"
	sourceTextsRelativePath         = "source-texts/source-texts.json"
	deconstructionCardsRelativePath = "deconstructions/writing-technique-cards.json"
	legacyDeconstructionCardsPath   = "deconstructions/deconstruction-cards.json"
	materialsRelativePath           = "materials/materials.json"
	materialTextsDir                = "materials/texts"
	libraryMaterialsRelativePath    = "materials/library-materials.json"
	libraryMaterialTextsDir         = "materials/texts"
	globalMaterialIndexFileName     = "global-material-index.json"
	deconstructionPlanRelativePath  = "deconstruction-plan.json"
	deconstructionTemplatesDir      = "deconstruction-templates"
	analysisKindWritingTechnique    = "writingTechnique"
	analysisKindStoryPlot           = "storyPlot"
	systemWritingTemplateFileName   = "系统内置写作手法.json"
	systemWritingTemplateName       = "系统内置写作手法"
	systemStoryTemplateFileName     = "系统内置剧情分析.json"
	systemStoryTemplateName         = "系统内置剧情分析"
	defaultReaderFontSize           = 18
	minReaderFontSize               = 14
	maxReaderFontSize               = 28
	libraryMaterialSourceReference  = "bookReference"
	libraryMaterialSourceOwned      = "libraryOwned"
	libraryMaterialSourceMigrated   = "migratedBookCopy"
)
