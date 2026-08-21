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
	
	    static createFrom(source: any = {}) {
	        return new LibraryBookEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bookName = source["bookName"];
	        this.bookPath = source["bookPath"];
	        this.bookSize = source["bookSize"];
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
	
	

}

