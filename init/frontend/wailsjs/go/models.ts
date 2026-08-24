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
	
	    static createFrom(source: any = {}) {
	        return new ConfirmImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.library = this.convertValues(source["library"], LibraryState);
	        this.project = this.convertValues(source["project"], BookProject);
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
	    dimensionId: string;
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
	        this.dimensionId = source["dimensionId"];
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
	    planTypeId: string;
	    planTypeName: string;
	    dimensionId: string;
	    dimensionName: string;
	    content: string;
	    fieldValues: Record<string, string>;
	    tagIds: string[];
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
	        this.planTypeId = source["planTypeId"];
	        this.planTypeName = source["planTypeName"];
	        this.dimensionId = source["dimensionId"];
	        this.dimensionName = source["dimensionName"];
	        this.content = source["content"];
	        this.fieldValues = source["fieldValues"];
	        this.tagIds = source["tagIds"];
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
	export class DeconstructionTemplateSummary {
	    name: string;
	    fileName: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionTemplateSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fileName = source["fileName"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class DeconstructionPlanWorkspace {
	    currentPlan: DeconstructionPlan;
	    currentTemplateFileName: string;
	    templates: DeconstructionTemplateSummary[];
	
	    static createFrom(source: any = {}) {
	        return new DeconstructionPlanWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentPlan = this.convertValues(source["currentPlan"], DeconstructionPlan);
	        this.currentTemplateFileName = source["currentTemplateFileName"];
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
	export class SetCurrentDeconstructionTemplateRequest {
	    bookPath: string;
	    fileName: string;
	
	    static createFrom(source: any = {}) {
	        return new SetCurrentDeconstructionTemplateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookPath = source["bookPath"];
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

}

