export namespace library {
	
	export class BookChapter {
	    id: string;
	    title: string;
	    order: number;
	    contentPath: string;
	    startOffset: number;
	    endOffset: number;
	    wordCount: number;
	
	    static createFrom(source: any = {}) {
	        return new BookChapter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.order = source["order"];
	        this.contentPath = source["contentPath"];
	        this.startOffset = source["startOffset"];
	        this.endOffset = source["endOffset"];
	        this.wordCount = source["wordCount"];
	    }
	}
	export class BookMeta {
	    id: string;
	    name: string;
	    description: string;
	    sourceFileName: string;
	    sourceTextPath: string;
	    bookPath: string;
	    bookSize: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BookMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.sourceFileName = source["sourceFileName"];
	        this.sourceTextPath = source["sourceTextPath"];
	        this.bookPath = source["bookPath"];
	        this.bookSize = source["bookSize"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class BookProject {
	    version: number;
	    book: BookMeta;
	    chapters: BookChapter[];
	
	    static createFrom(source: any = {}) {
	        return new BookProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.book = this.convertValues(source["book"], BookMeta);
	        this.chapters = this.convertValues(source["chapters"], BookChapter);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BuildMaterialDraftFromCardRequest {
	    bookPath: string;
	    cardId: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildMaterialDraftFromCardRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.cardId = source["cardId"];
	    }
	}
	export class ChapterContentRequest {
	    bookPath: string;
	    chapterId: string;
	
	    static createFrom(source: any = {}) {
	        return new ChapterContentRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.chapterId = source["chapterId"];
	    }
	}
	export class ChapterContentResult {
	    chapter: BookChapter;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ChapterContentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chapter = this.convertValues(source["chapter"], BookChapter);
	        this.content = source["content"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChapterMemo {
	    bookId: string;
	    chapterId: string;
	    chapterOrder: number;
	    content: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ChapterMemo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookId = source["bookId"];
	        this.chapterId = source["chapterId"];
	        this.chapterOrder = source["chapterOrder"];
	        this.content = source["content"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ChapterMemoCollection {
	    version: number;
	    items: ChapterMemo[];
	
	    static createFrom(source: any = {}) {
	        return new ChapterMemoCollection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.items = this.convertValues(source["items"], ChapterMemo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConfirmImportRequest {
	    bookId: string;
	    sourcePath: string;
	    bookName: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfirmImportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookId = source["bookId"];
	        this.sourcePath = source["sourcePath"];
	        this.bookName = source["bookName"];
	        this.description = source["description"];
	    }
	}
	export class ImportTimingStage {
	    name: string;
	    label: string;
	    durationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportTimingStage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.label = source["label"];
	        this.durationMs = source["durationMs"];
	    }
	}
	export class ImportDiagnostics {
	    action: string;
	    totalMs: number;
	    bookSize: number;
	    chapterCount: number;
	    stages: ImportTimingStage[];
	
	    static createFrom(source: any = {}) {
	        return new ImportDiagnostics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.totalMs = source["totalMs"];
	        this.bookSize = source["bookSize"];
	        this.chapterCount = source["chapterCount"];
	        this.stages = this.convertValues(source["stages"], ImportTimingStage);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryBookEntry {
	    bookName: string;
	    bookPath: string;
	    bookSize: number;
	    readingProgressPercent: number;
	
	    static createFrom(source: any = {}) {
	        return new LibraryBookEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookName = source["bookName"];
	        this.bookPath = source["bookPath"];
	        this.bookSize = source["bookSize"];
	        this.readingProgressPercent = source["readingProgressPercent"];
	    }
	}
	export class LibraryIndex {
	    version: number;
	    books: LibraryBookEntry[];
	
	    static createFrom(source: any = {}) {
	        return new LibraryIndex(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.books = this.convertValues(source["books"], LibraryBookEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryState {
	    libraryPath: string;
	    index: LibraryIndex;
	
	    static createFrom(source: any = {}) {
	        return new LibraryState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryPath = source["libraryPath"];
	        this.index = this.convertValues(source["index"], LibraryIndex);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConfirmImportResult {
	    library: LibraryState;
	    project: BookProject;
	    diagnostics: ImportDiagnostics;
	
	    static createFrom(source: any = {}) {
	        return new ConfirmImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.library = this.convertValues(source["library"], LibraryState);
	        this.project = this.convertValues(source["project"], BookProject);
	        this.diagnostics = this.convertValues(source["diagnostics"], ImportDiagnostics);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SourceSelection {
	    chapterId: string;
	    chapterOrder: number;
	    startOffset: number;
	    endOffset: number;
	    selectedText: string;
	
	    static createFrom(source: any = {}) {
	        return new SourceSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chapterId = source["chapterId"];
	        this.chapterOrder = source["chapterOrder"];
	        this.startOffset = source["startOffset"];
	        this.endOffset = source["endOffset"];
	        this.selectedText = source["selectedText"];
	    }
	}
	export class CreateDeconstructionRequest {
	    bookPath: string;
	    selection: SourceSelection;
	    planTypeId: string;
	    planTypeName: string;
	    dimensionId: string;
	    dimensionName: string;
	    content: string;
	    fieldValues: Record<string, string>;
	    tagIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateDeconstructionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.selection = this.convertValues(source["selection"], SourceSelection);
	        this.planTypeId = source["planTypeId"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionId = source["dimensionId"];
	        this.dimensionName = source["dimensionName"];
	        this.content = source["content"];
	        this.fieldValues = source["fieldValues"];
	        this.tagIds = source["tagIds"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeconstructionCardCollection {
	    version: number;
	    items: DeconstructionCard[];
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionCardCollection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.items = this.convertValues(source["items"], DeconstructionCard);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SourceTextCollection {
	    version: number;
	    items: SourceText[];
	
	    static createFrom(source: any = {}) {
	        return new SourceTextCollection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.items = this.convertValues(source["items"], SourceText);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeconstructionCard {
	    id: string;
	    bookId: string;
	    sourceTextId: string;
	    analysisKind: string;
	    planTypeId: string;
	    planTypeName: string;
	    dimensionId: string;
	    dimensionName: string;
	    content: string;
	    fieldValues: Record<string, string>;
	    tagIds: string[];
	    materialIds: string[];
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bookId = source["bookId"];
	        this.sourceTextId = source["sourceTextId"];
	        this.analysisKind = source["analysisKind"];
	        this.planTypeId = source["planTypeId"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionId = source["dimensionId"];
	        this.dimensionName = source["dimensionName"];
	        this.content = source["content"];
	        this.fieldValues = source["fieldValues"];
	        this.tagIds = source["tagIds"];
	        this.materialIds = source["materialIds"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class SourceText {
	    id: string;
	    bookId: string;
	    chapterId: string;
	    chapterOrder: number;
	    startOffset: number;
	    endOffset: number;
	    selectedText: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SourceText(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bookId = source["bookId"];
	        this.chapterId = source["chapterId"];
	        this.chapterOrder = source["chapterOrder"];
	        this.startOffset = source["startOffset"];
	        this.endOffset = source["endOffset"];
	        this.selectedText = source["selectedText"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class CreateDeconstructionResult {
	    sourceText: SourceText;
	    card: DeconstructionCard;
	    sourceTexts: SourceTextCollection;
	    deconstructionCards: DeconstructionCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new CreateDeconstructionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceText = this.convertValues(source["sourceText"], SourceText);
	        this.card = this.convertValues(source["card"], DeconstructionCard);
	        this.sourceTexts = this.convertValues(source["sourceTexts"], SourceTextCollection);
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateLibraryMaterialRequest {
	    libraryPath: string;
	    title: string;
	    summary: string;
	    content: string;
	    materialType: string;
	    planTypeName: string;
	    dimensionName: string;
	    tagIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateLibraryMaterialRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryPath = source["libraryPath"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	        this.materialType = source["materialType"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionName = source["dimensionName"];
	        this.tagIds = source["tagIds"];
	    }
	}
	export class CreateMaterialFromCardRequest {
	    bookPath: string;
	    cardId: string;
	    title: string;
	    summary: string;
	    content: string;
	    categoryIds: string[];
	    tagIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateMaterialFromCardRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.cardId = source["cardId"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	        this.categoryIds = source["categoryIds"];
	        this.tagIds = source["tagIds"];
	    }
	}
	export class MaterialCollection {
	    version: number;
	    items: Material[];
	
	    static createFrom(source: any = {}) {
	        return new MaterialCollection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.items = this.convertValues(source["items"], Material);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MaterialSourceSnapshot {
	    bookName: string;
	    chapterTitle: string;
	    deconstructionTitle: string;
	    planTypeName: string;
	    dimensionName: string;
	    sourceSummary: string;
	    selectedText: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new MaterialSourceSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookName = source["bookName"];
	        this.chapterTitle = source["chapterTitle"];
	        this.deconstructionTitle = source["deconstructionTitle"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionName = source["dimensionName"];
	        this.sourceSummary = source["sourceSummary"];
	        this.selectedText = source["selectedText"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class MaterialChapterRef {
	    chapterId: string;
	    chapterOrder: number;
	    chapterTitleSnapshot: string;
	
	    static createFrom(source: any = {}) {
	        return new MaterialChapterRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chapterId = source["chapterId"];
	        this.chapterOrder = source["chapterOrder"];
	        this.chapterTitleSnapshot = source["chapterTitleSnapshot"];
	    }
	}
	export class MaterialSourceRef {
	    sourceKind: string;
	    deconstructionId: string;
	    sourceTextId: string;
	    chapterRefs: MaterialChapterRef[];
	    snapshot: MaterialSourceSnapshot;
	
	    static createFrom(source: any = {}) {
	        return new MaterialSourceRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceKind = source["sourceKind"];
	        this.deconstructionId = source["deconstructionId"];
	        this.sourceTextId = source["sourceTextId"];
	        this.chapterRefs = this.convertValues(source["chapterRefs"], MaterialChapterRef);
	        this.snapshot = this.convertValues(source["snapshot"], MaterialSourceSnapshot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Material {
	    id: string;
	    bookId: string;
	    materialType: string;
	    title: string;
	    summary: string;
	    contentPath: string;
	    categoryIds: string[];
	    tagIds: string[];
	    sourceRefs: MaterialSourceRef[];
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Material(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bookId = source["bookId"];
	        this.materialType = source["materialType"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.contentPath = source["contentPath"];
	        this.categoryIds = source["categoryIds"];
	        this.tagIds = source["tagIds"];
	        this.sourceRefs = this.convertValues(source["sourceRefs"], MaterialSourceRef);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateMaterialResult {
	    material: Material;
	    materials: MaterialCollection;
	    materialContents: Record<string, string>;
	    deconstructionCards: DeconstructionCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new CreateMaterialResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.material = this.convertValues(source["material"], Material);
	        this.materials = this.convertValues(source["materials"], MaterialCollection);
	        this.materialContents = source["materialContents"];
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class DeconstructionField {
	    id: string;
	    label: string;
	    type: string;
	    order: number;
	    required: boolean;
	    visible: boolean;
	    options: string[];
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.order = source["order"];
	        this.required = source["required"];
	        this.visible = source["visible"];
	        this.options = source["options"];
	    }
	}
	export class DeconstructionDimension {
	    id: string;
	    name: string;
	    order: number;
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionDimension(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.order = source["order"];
	    }
	}
	export class DeconstructionType {
	    id: string;
	    name: string;
	    order: number;
	    dimensions: DeconstructionDimension[];
	    dimension?: string;
	    fields?: DeconstructionField[];
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.order = source["order"];
	        this.dimensions = this.convertValues(source["dimensions"], DeconstructionDimension);
	        this.dimension = source["dimension"];
	        this.fields = this.convertValues(source["fields"], DeconstructionField);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeconstructionPlan {
	    version: number;
	    bookId: string;
	    analysisKind: string;
	    templateName: string;
	    types: DeconstructionType[];
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.bookId = source["bookId"];
	        this.analysisKind = source["analysisKind"];
	        this.templateName = source["templateName"];
	        this.types = this.convertValues(source["types"], DeconstructionType);
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeconstructionCardsWorkspace {
	    project: BookProject;
	    plan: DeconstructionPlan;
	    sourceTexts: SourceTextCollection;
	    deconstructionCards: DeconstructionCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionCardsWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], BookProject);
	        this.plan = this.convertValues(source["plan"], DeconstructionPlan);
	        this.sourceTexts = this.convertValues(source["sourceTexts"], SourceTextCollection);
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class DeconstructionPlanState {
	    version: number;
	    bookId: string;
	    currentTemplates: Record<string, string>;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionPlanState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.bookId = source["bookId"];
	        this.currentTemplates = source["currentTemplates"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class DeconstructionTemplateSummary {
	    name: string;
	    fileName: string;
	    analysisKind: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionTemplateSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fileName = source["fileName"];
	        this.analysisKind = source["analysisKind"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class DeconstructionPlanWorkspace {
	    currentPlan: DeconstructionPlan;
	    currentTemplateFileName: string;
	    currentPlanState: DeconstructionPlanState;
	    currentPlans: Record<string, DeconstructionPlan>;
	    currentTemplateFileNames: Record<string, string>;
	    templates: DeconstructionTemplateSummary[];
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionPlanWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentPlan = this.convertValues(source["currentPlan"], DeconstructionPlan);
	        this.currentTemplateFileName = source["currentTemplateFileName"];
	        this.currentPlanState = this.convertValues(source["currentPlanState"], DeconstructionPlanState);
	        this.currentPlans = this.convertValues(source["currentPlans"], DeconstructionPlan, true);
	        this.currentTemplateFileNames = source["currentTemplateFileNames"];
	        this.templates = this.convertValues(source["templates"], DeconstructionTemplateSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class DeleteDeconstructionCardRequest {
	    bookPath: string;
	    cardId: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteDeconstructionCardRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.cardId = source["cardId"];
	    }
	}
	export class DeleteDeconstructionCardResult {
	    sourceTexts: SourceTextCollection;
	    deconstructionCards: DeconstructionCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new DeleteDeconstructionCardResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceTexts = this.convertValues(source["sourceTexts"], SourceTextCollection);
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeleteLibraryMaterialRequest {
	    libraryPath: string;
	    libraryMaterialId: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteLibraryMaterialRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryPath = source["libraryPath"];
	        this.libraryMaterialId = source["libraryMaterialId"];
	    }
	}
	export class DeleteMaterialRequest {
	    bookPath: string;
	    materialId: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteMaterialRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.materialId = source["materialId"];
	    }
	}
	export class DeleteMaterialResult {
	    materials: MaterialCollection;
	    materialContents: Record<string, string>;
	    deconstructionCards: DeconstructionCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new DeleteMaterialResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.materials = this.convertValues(source["materials"], MaterialCollection);
	        this.materialContents = source["materialContents"];
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeleteStoryPlotCardRequest {
	    bookPath: string;
	    cardId: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteStoryPlotCardRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.cardId = source["cardId"];
	    }
	}
	export class StoryPlotTypeSnapshot {
	    id: string;
	    name: string;
	    order: number;
	    dimensions: StoryPlotDimension[];
	
	    static createFrom(source: any = {}) {
	        return new StoryPlotTypeSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.order = source["order"];
	        this.dimensions = this.convertValues(source["dimensions"], StoryPlotDimension);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StoryPlotDimension {
	    id: string;
	    name: string;
	    order: number;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new StoryPlotDimension(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.order = source["order"];
	        this.content = source["content"];
	    }
	}
	export class StoryPlotChapterRef {
	    chapterId: string;
	    chapterOrder: number;
	    chapterTitleSnapshot: string;
	
	    static createFrom(source: any = {}) {
	        return new StoryPlotChapterRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chapterId = source["chapterId"];
	        this.chapterOrder = source["chapterOrder"];
	        this.chapterTitleSnapshot = source["chapterTitleSnapshot"];
	    }
	}
	export class StoryPlotCard {
	    id: string;
	    bookId: string;
	    analysisKind: string;
	    status: string;
	    title: string;
	    summary: string;
	    chapterRangeInput: string;
	    chapterRefs: StoryPlotChapterRef[];
	    parseErrors: string[];
	    templateFileName: string;
	    templateName: string;
	    planTypeId: string;
	    planTypeName: string;
	    dimensions: StoryPlotDimension[];
	    typeSnapshots: StoryPlotTypeSnapshot[];
	    tagIds: string[];
	    materialIds: string[];
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new StoryPlotCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bookId = source["bookId"];
	        this.analysisKind = source["analysisKind"];
	        this.status = source["status"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.chapterRangeInput = source["chapterRangeInput"];
	        this.chapterRefs = this.convertValues(source["chapterRefs"], StoryPlotChapterRef);
	        this.parseErrors = source["parseErrors"];
	        this.templateFileName = source["templateFileName"];
	        this.templateName = source["templateName"];
	        this.planTypeId = source["planTypeId"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensions = this.convertValues(source["dimensions"], StoryPlotDimension);
	        this.typeSnapshots = this.convertValues(source["typeSnapshots"], StoryPlotTypeSnapshot);
	        this.tagIds = source["tagIds"];
	        this.materialIds = source["materialIds"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StoryPlotCardCollection {
	    version: number;
	    items: StoryPlotCard[];
	
	    static createFrom(source: any = {}) {
	        return new StoryPlotCardCollection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.items = this.convertValues(source["items"], StoryPlotCard);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeleteStoryPlotCardResult {
	    storyPlotCards: StoryPlotCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new DeleteStoryPlotCardResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.storyPlotCards = this.convertValues(source["storyPlotCards"], StoryPlotCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ImportChapterPreview {
	    id: string;
	    order: number;
	    title: string;
	    startOffset: number;
	    endOffset: number;
	    wordCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportChapterPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.order = source["order"];
	        this.title = source["title"];
	        this.startOffset = source["startOffset"];
	        this.endOffset = source["endOffset"];
	        this.wordCount = source["wordCount"];
	    }
	}
	
	export class ImportDraft {
	    bookId: string;
	    bookName: string;
	    description: string;
	    sourcePath: string;
	    libraryPath: string;
	    targetPath: string;
	    bookSize: number;
	    chapterCount: number;
	    chapters: ImportChapterPreview[];
	    diagnostics: ImportDiagnostics;
	
	    static createFrom(source: any = {}) {
	        return new ImportDraft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookId = source["bookId"];
	        this.bookName = source["bookName"];
	        this.description = source["description"];
	        this.sourcePath = source["sourcePath"];
	        this.libraryPath = source["libraryPath"];
	        this.targetPath = source["targetPath"];
	        this.bookSize = source["bookSize"];
	        this.chapterCount = source["chapterCount"];
	        this.chapters = this.convertValues(source["chapters"], ImportChapterPreview);
	        this.diagnostics = this.convertValues(source["diagnostics"], ImportDiagnostics);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class LibraryMaterial {
	    id: string;
	    sourceMode: string;
	    bookId: string;
	    bookName: string;
	    originalBookPath: string;
	    bookMaterialId: string;
	    materialType: string;
	    planTypeName: string;
	    dimensionName: string;
	    title: string;
	    summary: string;
	    contentPath: string;
	    tagIds: string[];
	    sourceRefs: MaterialSourceRef[];
	    createdAt: string;
	    updatedAt: string;
	    migratedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new LibraryMaterial(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourceMode = source["sourceMode"];
	        this.bookId = source["bookId"];
	        this.bookName = source["bookName"];
	        this.originalBookPath = source["originalBookPath"];
	        this.bookMaterialId = source["bookMaterialId"];
	        this.materialType = source["materialType"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionName = source["dimensionName"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.contentPath = source["contentPath"];
	        this.tagIds = source["tagIds"];
	        this.sourceRefs = this.convertValues(source["sourceRefs"], MaterialSourceRef);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.migratedAt = source["migratedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryMaterialBookOption {
	    bookId: string;
	    bookName: string;
	
	    static createFrom(source: any = {}) {
	        return new LibraryMaterialBookOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookId = source["bookId"];
	        this.bookName = source["bookName"];
	    }
	}
	export class LibraryMaterialCollection {
	    version: number;
	    items: LibraryMaterial[];
	
	    static createFrom(source: any = {}) {
	        return new LibraryMaterialCollection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.items = this.convertValues(source["items"], LibraryMaterial);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryMaterialListItem {
	    id: string;
	    sourceMode: string;
	    libraryMaterialId: string;
	    bookId: string;
	    bookName: string;
	    bookPath: string;
	    bookMaterialId: string;
	    materialType: string;
	    planTypeName: string;
	    dimensionName: string;
	    title: string;
	    summary: string;
	    content: string;
	    tagIds: string[];
	    sourceRefs: MaterialSourceRef[];
	    createdAt: string;
	    updatedAt: string;
	    sourceAvailable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LibraryMaterialListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourceMode = source["sourceMode"];
	        this.libraryMaterialId = source["libraryMaterialId"];
	        this.bookId = source["bookId"];
	        this.bookName = source["bookName"];
	        this.bookPath = source["bookPath"];
	        this.bookMaterialId = source["bookMaterialId"];
	        this.materialType = source["materialType"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionName = source["dimensionName"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	        this.tagIds = source["tagIds"];
	        this.sourceRefs = this.convertValues(source["sourceRefs"], MaterialSourceRef);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.sourceAvailable = source["sourceAvailable"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryMaterialPage {
	    libraryPath: string;
	    items: LibraryMaterialListItem[];
	    total: number;
	    page: number;
	    pageSize: number;
	    books: LibraryMaterialBookOption[];
	    types: string[];
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new LibraryMaterialPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryPath = source["libraryPath"];
	        this.items = this.convertValues(source["items"], LibraryMaterialListItem);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.books = this.convertValues(source["books"], LibraryMaterialBookOption);
	        this.types = source["types"];
	        this.tags = source["tags"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryMaterialPageRequest {
	    libraryPath: string;
	    keyword: string;
	    bookId: string;
	    materialType: string;
	    tagId: string;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new LibraryMaterialPageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryPath = source["libraryPath"];
	        this.keyword = source["keyword"];
	        this.bookId = source["bookId"];
	        this.materialType = source["materialType"];
	        this.tagId = source["tagId"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	
	
	
	
	export class MaterialDraft {
	    materialType: string;
	    title: string;
	    summary: string;
	    content: string;
	    categoryIds: string[];
	    tagIds: string[];
	    sourceRefs: MaterialSourceRef[];
	
	    static createFrom(source: any = {}) {
	        return new MaterialDraft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.materialType = source["materialType"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	        this.categoryIds = source["categoryIds"];
	        this.tagIds = source["tagIds"];
	        this.sourceRefs = this.convertValues(source["sourceRefs"], MaterialSourceRef);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class MaterialWorkspace {
	    project: BookProject;
	    materials: MaterialCollection;
	    materialContents: Record<string, string>;
	    sourceTexts: SourceTextCollection;
	    deconstructionCards: DeconstructionCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new MaterialWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], BookProject);
	        this.materials = this.convertValues(source["materials"], MaterialCollection);
	        this.materialContents = source["materialContents"];
	        this.sourceTexts = this.convertValues(source["sourceTexts"], SourceTextCollection);
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ReadingPosition {
	    chapterId: string;
	    chapterOrder: number;
	    charOffset: number;
	    paragraphIndex: number;
	    renderedPageNumber: number;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ReadingPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chapterId = source["chapterId"];
	        this.chapterOrder = source["chapterOrder"];
	        this.charOffset = source["charOffset"];
	        this.paragraphIndex = source["paragraphIndex"];
	        this.renderedPageNumber = source["renderedPageNumber"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ReadingPreferences {
	    fontSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ReadingPreferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fontSize = source["fontSize"];
	    }
	}
	export class ReadingState {
	    version: number;
	    bookId: string;
	    lastPosition: ReadingPosition;
	    lastMode: string;
	    readingProgressPercent: number;
	    preferences: ReadingPreferences;
	
	    static createFrom(source: any = {}) {
	        return new ReadingState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.bookId = source["bookId"];
	        this.lastPosition = this.convertValues(source["lastPosition"], ReadingPosition);
	        this.lastMode = source["lastMode"];
	        this.readingProgressPercent = source["readingProgressPercent"];
	        this.preferences = this.convertValues(source["preferences"], ReadingPreferences);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ReadingSession {
	    project: BookProject;
	    state: ReadingState;
	    currentChapter: BookChapter;
	    chapterContent: string;
	    plan: DeconstructionPlan;
	    sourceTexts: SourceTextCollection;
	    deconstructionCards: DeconstructionCardCollection;
	    chapterMemos: ChapterMemoCollection;
	
	    static createFrom(source: any = {}) {
	        return new ReadingSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], BookProject);
	        this.state = this.convertValues(source["state"], ReadingState);
	        this.currentChapter = this.convertValues(source["currentChapter"], BookChapter);
	        this.chapterContent = source["chapterContent"];
	        this.plan = this.convertValues(source["plan"], DeconstructionPlan);
	        this.sourceTexts = this.convertValues(source["sourceTexts"], SourceTextCollection);
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	        this.chapterMemos = this.convertValues(source["chapterMemos"], ChapterMemoCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SaveChapterContentRequest {
	    bookPath: string;
	    chapterId: string;
	    title: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveChapterContentRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.chapterId = source["chapterId"];
	        this.title = source["title"];
	        this.content = source["content"];
	    }
	}
	export class SaveChapterMemoRequest {
	    bookPath: string;
	    chapterId: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveChapterMemoRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.chapterId = source["chapterId"];
	        this.content = source["content"];
	    }
	}
	export class SaveChapterMemoResult {
	    memo: ChapterMemo;
	    chapterMemos: ChapterMemoCollection;
	
	    static createFrom(source: any = {}) {
	        return new SaveChapterMemoResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.memo = this.convertValues(source["memo"], ChapterMemo);
	        this.chapterMemos = this.convertValues(source["chapterMemos"], ChapterMemoCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveDeconstructionPlanRequest {
	    bookPath: string;
	    plan: DeconstructionPlan;
	
	    static createFrom(source: any = {}) {
	        return new SaveDeconstructionPlanRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.plan = this.convertValues(source["plan"], DeconstructionPlan);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveDeconstructionTemplateRequest {
	    bookPath: string;
	    analysisKind: string;
	    templateName: string;
	    fileName: string;
	    plan: DeconstructionPlan;
	    overwrite: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SaveDeconstructionTemplateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.analysisKind = source["analysisKind"];
	        this.templateName = source["templateName"];
	        this.fileName = source["fileName"];
	        this.plan = this.convertValues(source["plan"], DeconstructionPlan);
	        this.overwrite = source["overwrite"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveReadingPositionRequest {
	    bookPath: string;
	    chapterId: string;
	    charOffset: number;
	    paragraphIndex: number;
	    renderedPageNumber: number;
	    lastMode: string;
	    readingProgressPercent: number;
	
	    static createFrom(source: any = {}) {
	        return new SaveReadingPositionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.chapterId = source["chapterId"];
	        this.charOffset = source["charOffset"];
	        this.paragraphIndex = source["paragraphIndex"];
	        this.renderedPageNumber = source["renderedPageNumber"];
	        this.lastMode = source["lastMode"];
	        this.readingProgressPercent = source["readingProgressPercent"];
	    }
	}
	export class SaveReadingPreferencesRequest {
	    bookPath: string;
	    fontSize: number;
	
	    static createFrom(source: any = {}) {
	        return new SaveReadingPreferencesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.fontSize = source["fontSize"];
	    }
	}
	export class SaveStoryPlotCardRequest {
	    bookPath: string;
	    cardId: string;
	    status: string;
	    title: string;
	    summary: string;
	    chapterRangeInput: string;
	    templateFileName: string;
	    templateName: string;
	    planTypeId: string;
	    planTypeName: string;
	    dimensions: StoryPlotDimension[];
	    typeSnapshots: StoryPlotTypeSnapshot[];
	    tagIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new SaveStoryPlotCardRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.cardId = source["cardId"];
	        this.status = source["status"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.chapterRangeInput = source["chapterRangeInput"];
	        this.templateFileName = source["templateFileName"];
	        this.templateName = source["templateName"];
	        this.planTypeId = source["planTypeId"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensions = this.convertValues(source["dimensions"], StoryPlotDimension);
	        this.typeSnapshots = this.convertValues(source["typeSnapshots"], StoryPlotTypeSnapshot);
	        this.tagIds = source["tagIds"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveStoryPlotCardResult {
	    card: StoryPlotCard;
	    storyPlotCards: StoryPlotCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new SaveStoryPlotCardResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.card = this.convertValues(source["card"], StoryPlotCard);
	        this.storyPlotCards = this.convertValues(source["storyPlotCards"], StoryPlotCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SetCurrentDeconstructionTemplateRequest {
	    bookPath: string;
	    analysisKind: string;
	    fileName: string;
	
	    static createFrom(source: any = {}) {
	        return new SetCurrentDeconstructionTemplateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.analysisKind = source["analysisKind"];
	        this.fileName = source["fileName"];
	    }
	}
	
	
	
	export class SplitChapterRequest {
	    bookPath: string;
	    chapterId: string;
	    newTitle: string;
	    newContent: string;
	    remainingContent: string;
	
	    static createFrom(source: any = {}) {
	        return new SplitChapterRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.chapterId = source["chapterId"];
	        this.newTitle = source["newTitle"];
	        this.newContent = source["newContent"];
	        this.remainingContent = source["remainingContent"];
	    }
	}
	export class SplitChapterResult {
	    project: BookProject;
	    newChapter: BookChapter;
	    remainingChapter: BookChapter;
	    newContent: string;
	    remainingContent: string;
	
	    static createFrom(source: any = {}) {
	        return new SplitChapterResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], BookProject);
	        this.newChapter = this.convertValues(source["newChapter"], BookChapter);
	        this.remainingChapter = this.convertValues(source["remainingChapter"], BookChapter);
	        this.newContent = source["newContent"];
	        this.remainingContent = source["remainingContent"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	export class StoryPlotWorkspace {
	    project: BookProject;
	    plan: DeconstructionPlan;
	    currentTemplateFileName: string;
	    storyPlotCards: StoryPlotCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new StoryPlotWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], BookProject);
	        this.plan = this.convertValues(source["plan"], DeconstructionPlan);
	        this.currentTemplateFileName = source["currentTemplateFileName"];
	        this.storyPlotCards = this.convertValues(source["storyPlotCards"], StoryPlotCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateDeconstructionCardRequest {
	    bookPath: string;
	    cardId: string;
	    planTypeId: string;
	    planTypeName: string;
	    dimensionId: string;
	    dimensionName: string;
	    content: string;
	    tagIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new UpdateDeconstructionCardRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.cardId = source["cardId"];
	        this.planTypeId = source["planTypeId"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionId = source["dimensionId"];
	        this.dimensionName = source["dimensionName"];
	        this.content = source["content"];
	        this.tagIds = source["tagIds"];
	    }
	}
	export class UpdateDeconstructionCardResult {
	    card: DeconstructionCard;
	    deconstructionCards: DeconstructionCardCollection;
	
	    static createFrom(source: any = {}) {
	        return new UpdateDeconstructionCardResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.card = this.convertValues(source["card"], DeconstructionCard);
	        this.deconstructionCards = this.convertValues(source["deconstructionCards"], DeconstructionCardCollection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateLibraryMaterialRequest {
	    libraryPath: string;
	    libraryMaterialId: string;
	    title: string;
	    summary: string;
	    content: string;
	    materialType: string;
	    planTypeName: string;
	    dimensionName: string;
	    tagIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new UpdateLibraryMaterialRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryPath = source["libraryPath"];
	        this.libraryMaterialId = source["libraryMaterialId"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	        this.materialType = source["materialType"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionName = source["dimensionName"];
	        this.tagIds = source["tagIds"];
	    }
	}
	export class UpdateMaterialRequest {
	    bookPath: string;
	    materialId: string;
	    title: string;
	    summary: string;
	    content: string;
	    categoryIds: string[];
	    tagIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new UpdateMaterialRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
	        this.materialId = source["materialId"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	        this.categoryIds = source["categoryIds"];
	        this.tagIds = source["tagIds"];
	    }
	}
	export class UpdateMaterialResult {
	    material: Material;
	    materials: MaterialCollection;
	    materialContents: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new UpdateMaterialResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.material = this.convertValues(source["material"], Material);
	        this.materials = this.convertValues(source["materials"], MaterialCollection);
	        this.materialContents = source["materialContents"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class ApplicationInfo {
	    name: string;
	    displayName: string;
	    version: string;
	    description: string;
	    authors: string[];
	    contactEmail: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplicationInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.authors = source["authors"];
	        this.contactEmail = source["contactEmail"];
	    }
	}

}

