export namespace automation {
	
	export class ScriptPublicAPIVariable {
	    name: string;
	    defaultValue: string;
	    description: string;
	    required: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScriptPublicAPIVariable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.defaultValue = source["defaultValue"];
	        this.description = source["description"];
	        this.required = source["required"];
	    }
	}
	export class ScriptPublicAPIConfig {
	    enabled: boolean;
	    method: string;
	    path: string;
	    requestMode: string;
	    responseMode: string;
	    timeoutMs: number;
	    requestBodyText: string;
	    responseBodyText: string;
	    variables: ScriptPublicAPIVariable[];
	
	    static createFrom(source: any = {}) {
	        return new ScriptPublicAPIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.method = source["method"];
	        this.path = source["path"];
	        this.requestMode = source["requestMode"];
	        this.responseMode = source["responseMode"];
	        this.timeoutMs = source["timeoutMs"];
	        this.requestBodyText = source["requestBodyText"];
	        this.responseBodyText = source["responseBodyText"];
	        this.variables = this.convertValues(source["variables"], ScriptPublicAPIVariable);
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
	
	export class ScriptSource {
	    type: string;
	    uri: string;
	    ref: string;
	    path: string;
	    importedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.uri = source["uri"];
	        this.ref = source["ref"];
	        this.path = source["path"];
	        this.importedAt = source["importedAt"];
	    }
	}
	export class ScriptTargetSelector {
	    code: string;
	    profileId: string;
	    profileName: string;
	    groupId: string;
	    keywords: string[];
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new ScriptTargetSelector(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.profileId = source["profileId"];
	        this.profileName = source["profileName"];
	        this.groupId = source["groupId"];
	        this.keywords = source["keywords"];
	        this.tags = source["tags"];
	    }
	}
	export class ScriptTargetConfig {
	    mode: string;
	    selector: ScriptTargetSelector;
	    templateSelector: ScriptTargetSelector;
	    createNameTemplate: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptTargetConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.selector = this.convertValues(source["selector"], ScriptTargetSelector);
	        this.templateSelector = this.convertValues(source["templateSelector"], ScriptTargetSelector);
	        this.createNameTemplate = source["createNameTemplate"];
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
	export class ScriptRecord {
	    packageFormat: string;
	    manifestVersion: number;
	    id: string;
	    name: string;
	    description: string;
	    type: string;
	    status: string;
	    entryFile: string;
	    tags: string[];
	    selectorText: string;
	    paramsText: string;
	    scriptText: string;
	    notes: string;
	    targetConfig: ScriptTargetConfig;
	    publicAPI: ScriptPublicAPIConfig;
	    source: ScriptSource;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.packageFormat = source["packageFormat"];
	        this.manifestVersion = source["manifestVersion"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.entryFile = source["entryFile"];
	        this.tags = source["tags"];
	        this.selectorText = source["selectorText"];
	        this.paramsText = source["paramsText"];
	        this.scriptText = source["scriptText"];
	        this.notes = source["notes"];
	        this.targetConfig = this.convertValues(source["targetConfig"], ScriptTargetConfig);
	        this.publicAPI = this.convertValues(source["publicAPI"], ScriptPublicAPIConfig);
	        this.source = this.convertValues(source["source"], ScriptSource);
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
	export class ScriptRunRecord {
	    id: string;
	    scriptId: string;
	    scriptName: string;
	    scriptType: string;
	    status: string;
	    summary: string;
	    error: string;
	    resultText: string;
	    logText: string;
	    startedAt: string;
	    finishedAt: string;
	    durationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new ScriptRunRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.scriptId = source["scriptId"];
	        this.scriptName = source["scriptName"];
	        this.scriptType = source["scriptType"];
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.error = source["error"];
	        this.resultText = source["resultText"];
	        this.logText = source["logText"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	        this.durationMs = source["durationMs"];
	    }
	}
	export class ScriptRunRequest {
	    scriptId: string;
	    selectorText: string;
	    targetMode?: string;
	    targetInput?: any;
	    paramsText: string;
	    useScriptSelector: boolean;
	    useScriptParams: boolean;
	    timeoutMs?: number;
	
	    static createFrom(source: any = {}) {
	        return new ScriptRunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scriptId = source["scriptId"];
	        this.selectorText = source["selectorText"];
	        this.targetMode = source["targetMode"];
	        this.targetInput = source["targetInput"];
	        this.paramsText = source["paramsText"];
	        this.useScriptSelector = source["useScriptSelector"];
	        this.useScriptParams = source["useScriptParams"];
	        this.timeoutMs = source["timeoutMs"];
	    }
	}
	
	

}

export namespace backend {
	
	
	export class BookmarkSyncResult {
	    total: number;
	    synced: number;
	    skipped: number;
	    failed: number;
	    skippedList: string[];
	    failedList: string[];
	
	    static createFrom(source: any = {}) {
	        return new BookmarkSyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.synced = source["synced"];
	        this.skipped = source["skipped"];
	        this.failed = source["failed"];
	        this.skippedList = source["skippedList"];
	        this.failedList = source["failedList"];
	    }
	}
	export class BrowserCorePickResult {
	    corePath: string;
	    suggestedName: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowserCorePickResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.corePath = source["corePath"];
	        this.suggestedName = source["suggestedName"];
	    }
	}
	export class BrowserExtensionWebStoreRequest {
	    query: string;
	    useProxy: boolean;
	    proxyConfig: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowserExtensionWebStoreRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.useProxy = source["useProxy"];
	        this.proxyConfig = source["proxyConfig"];
	    }
	}
	export class CookieInfo {
	    name: string;
	    value: string;
	    domain: string;
	    path: string;
	    expires: number;
	    httpOnly: boolean;
	    secure: boolean;
	    sameSite: string;
	
	    static createFrom(source: any = {}) {
	        return new CookieInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.domain = source["domain"];
	        this.path = source["path"];
	        this.expires = source["expires"];
	        this.httpOnly = source["httpOnly"];
	        this.secure = source["secure"];
	        this.sameSite = source["sameSite"];
	    }
	}
	export class ProfilePackageExportResult {
	    cancelled: boolean;
	    zipPath: string;
	    profileCount: number;
	    fileCount: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfilePackageExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cancelled = source["cancelled"];
	        this.zipPath = source["zipPath"];
	        this.profileCount = source["profileCount"];
	        this.fileCount = source["fileCount"];
	        this.message = source["message"];
	    }
	}
	export class ProfilePackageImportResult {
	    cancelled: boolean;
	    importedCount: number;
	    profileMappings: Record<string, string>;
	    warnings: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfilePackageImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cancelled = source["cancelled"];
	        this.importedCount = source["importedCount"];
	        this.profileMappings = source["profileMappings"];
	        this.warnings = source["warnings"];
	        this.message = source["message"];
	    }
	}
	export class ProxyIPHealthResult {
	    proxyId: string;
	    ok: boolean;
	    source: string;
	    error: string;
	    ip: string;
	    fraudScore: number;
	    isResidential: boolean;
	    isBroadcast: boolean;
	    country: string;
	    region: string;
	    city: string;
	    asOrganization: string;
	    rawData: Record<string, any>;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyIPHealthResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.ok = source["ok"];
	        this.source = source["source"];
	        this.error = source["error"];
	        this.ip = source["ip"];
	        this.fraudScore = source["fraudScore"];
	        this.isResidential = source["isResidential"];
	        this.isBroadcast = source["isBroadcast"];
	        this.country = source["country"];
	        this.region = source["region"];
	        this.city = source["city"];
	        this.asOrganization = source["asOrganization"];
	        this.rawData = source["rawData"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ProxyLocationOption {
	    label: string;
	    timezone: string;
	    lang: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyLocationOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.timezone = source["timezone"];
	        this.lang = source["lang"];
	    }
	}
	export class ProxyLocationResolveResult {
	    proxyId: string;
	    ok: boolean;
	    auto: boolean;
	    source: string;
	    error: string;
	    ip: string;
	    country: string;
	    region: string;
	    city: string;
	    timezone: string;
	    lang: string;
	    health?: ProxyIPHealthResult;
	    alternates?: ProxyLocationOption[];
	    resolvedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyLocationResolveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.ok = source["ok"];
	        this.auto = source["auto"];
	        this.source = source["source"];
	        this.error = source["error"];
	        this.ip = source["ip"];
	        this.country = source["country"];
	        this.region = source["region"];
	        this.city = source["city"];
	        this.timezone = source["timezone"];
	        this.lang = source["lang"];
	        this.health = this.convertValues(source["health"], ProxyIPHealthResult);
	        this.alternates = this.convertValues(source["alternates"], ProxyLocationOption);
	        this.resolvedAt = source["resolvedAt"];
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
	export class ProxyTestResult {
	    proxyId: string;
	    ok: boolean;
	    latencyMs: number;
	    engine: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.ok = source["ok"];
	        this.latencyMs = source["latencyMs"];
	        this.engine = source["engine"];
	        this.error = source["error"];
	    }
	}
	export class ProxyValidationResult {
	    supported: boolean;
	    errorMsg: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported = source["supported"];
	        this.errorMsg = source["errorMsg"];
	    }
	}
	export class SnapshotInfo {
	    snapshotId: string;
	    profileId: string;
	    name: string;
	    sizeMB: number;
	    createdAt: string;
	    filePath?: string;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.snapshotId = source["snapshotId"];
	        this.profileId = source["profileId"];
	        this.name = source["name"];
	        this.sizeMB = source["sizeMB"];
	        this.createdAt = source["createdAt"];
	        this.filePath = source["filePath"];
	    }
	}

}

export namespace backup {
	
	export class ManifestEntry {
	    id: string;
	    category: string;
	    entryType: string;
	    required: boolean;
	    archivePath: string;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new ManifestEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.entryType = source["entryType"];
	        this.required = source["required"];
	        this.archivePath = source["archivePath"];
	        this.description = source["description"];
	    }
	}
	export class ManifestAppInfo {
	    name: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new ManifestAppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	    }
	}
	export class Manifest {
	    format: string;
	    manifestVersion: number;
	    createdAt: string;
	    app: ManifestAppInfo;
	    entries: ManifestEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Manifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.manifestVersion = source["manifestVersion"];
	        this.createdAt = source["createdAt"];
	        this.app = this.convertValues(source["app"], ManifestAppInfo);
	        this.entries = this.convertValues(source["entries"], ManifestEntry);
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
	
	
	export class ScopeEntry {
	    id: string;
	    category: string;
	    entryType: string;
	    required: boolean;
	    sourcePath: string;
	    archivePath: string;
	    exists: boolean;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new ScopeEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.entryType = source["entryType"];
	        this.required = source["required"];
	        this.sourcePath = source["sourcePath"];
	        this.archivePath = source["archivePath"];
	        this.exists = source["exists"];
	        this.description = source["description"];
	    }
	}
	export class Scope {
	    format: string;
	    manifestVersion: number;
	    appRoot: string;
	    entries: ScopeEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Scope(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.manifestVersion = source["manifestVersion"];
	        this.appRoot = source["appRoot"];
	        this.entries = this.convertValues(source["entries"], ScopeEntry);
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

export namespace browser {
	
	export class CoreExtendedInfo {
	    coreId: string;
	    chromeVersion: string;
	    instanceCount: number;
	
	    static createFrom(source: any = {}) {
	        return new CoreExtendedInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coreId = source["coreId"];
	        this.chromeVersion = source["chromeVersion"];
	        this.instanceCount = source["instanceCount"];
	    }
	}
	export class CoreInput {
	    coreId: string;
	    coreName: string;
	    corePath: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CoreInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coreId = source["coreId"];
	        this.coreName = source["coreName"];
	        this.corePath = source["corePath"];
	        this.isDefault = source["isDefault"];
	    }
	}
	export class CoreValidateResult {
	    valid: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new CoreValidateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.message = source["message"];
	    }
	}
	export class Extension {
	    extensionId: string;
	    name: string;
	    version: string;
	    description: string;
	    iconDataUrl: string;
	    manifestJson: string;
	    sourceUrl: string;
	    installDir: string;
	    enabled: boolean;
	    installedAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Extension(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.extensionId = source["extensionId"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.iconDataUrl = source["iconDataUrl"];
	        this.manifestJson = source["manifestJson"];
	        this.sourceUrl = source["sourceUrl"];
	        this.installDir = source["installDir"];
	        this.enabled = source["enabled"];
	        this.installedAt = source["installedAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ExtensionLookupResult {
	    extensionId: string;
	    name: string;
	    version: string;
	    description: string;
	    storeUrl: string;
	    installable: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ExtensionLookupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.extensionId = source["extensionId"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.storeUrl = source["storeUrl"];
	        this.installable = source["installable"];
	        this.message = source["message"];
	    }
	}
	export class Group {
	    groupId: string;
	    groupName: string;
	    parentId: string;
	    sortOrder: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupId = source["groupId"];
	        this.groupName = source["groupName"];
	        this.parentId = source["parentId"];
	        this.sortOrder = source["sortOrder"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class GroupInput {
	    groupName: string;
	    parentId: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new GroupInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupName = source["groupName"];
	        this.parentId = source["parentId"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class GroupWithCount {
	    groupId: string;
	    groupName: string;
	    parentId: string;
	    sortOrder: number;
	    createdAt: string;
	    updatedAt: string;
	    instanceCount: number;
	
	    static createFrom(source: any = {}) {
	        return new GroupWithCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupId = source["groupId"];
	        this.groupName = source["groupName"];
	        this.parentId = source["parentId"];
	        this.sortOrder = source["sortOrder"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.instanceCount = source["instanceCount"];
	    }
	}
	export class Profile {
	    profileId: string;
	    profileName: string;
	    userDataDir: string;
	    coreId: string;
	    fingerprintArgs: string[];
	    proxyId: string;
	    proxyConfig: string;
	    proxyBindName: string;
	    proxyBindUpdatedAt: string;
	    launchArgs: string[];
	    tags: string[];
	    keywords: string[];
	    groupId: string;
	    launchCode: string;
	    running: boolean;
	    debugPort: number;
	    debugReady: boolean;
	    pid: number;
	    runtimeWarning: string;
	    lastError: string;
	    createdAt: string;
	    updatedAt: string;
	    deletedAt: string;
	    lastStartAt: string;
	    lastStopAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileId = source["profileId"];
	        this.profileName = source["profileName"];
	        this.userDataDir = source["userDataDir"];
	        this.coreId = source["coreId"];
	        this.fingerprintArgs = source["fingerprintArgs"];
	        this.proxyId = source["proxyId"];
	        this.proxyConfig = source["proxyConfig"];
	        this.proxyBindName = source["proxyBindName"];
	        this.proxyBindUpdatedAt = source["proxyBindUpdatedAt"];
	        this.launchArgs = source["launchArgs"];
	        this.tags = source["tags"];
	        this.keywords = source["keywords"];
	        this.groupId = source["groupId"];
	        this.launchCode = source["launchCode"];
	        this.running = source["running"];
	        this.debugPort = source["debugPort"];
	        this.debugReady = source["debugReady"];
	        this.pid = source["pid"];
	        this.runtimeWarning = source["runtimeWarning"];
	        this.lastError = source["lastError"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.deletedAt = source["deletedAt"];
	        this.lastStartAt = source["lastStartAt"];
	        this.lastStopAt = source["lastStopAt"];
	    }
	}
	export class ProfileCopyOptions {
	    mode: string;
	    automationTargets: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProfileCopyOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.automationTargets = source["automationTargets"];
	    }
	}
	export class ProfileExtensionSettings {
	    profileId: string;
	    configured: boolean;
	    extensionIds: string[];
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileExtensionSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileId = source["profileId"];
	        this.configured = source["configured"];
	        this.extensionIds = source["extensionIds"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ProfileInput {
	    profileName: string;
	    userDataDir: string;
	    coreId: string;
	    fingerprintArgs: string[];
	    proxyId: string;
	    proxyConfig: string;
	    launchArgs: string[];
	    tags: string[];
	    keywords: string[];
	    groupId: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileName = source["profileName"];
	        this.userDataDir = source["userDataDir"];
	        this.coreId = source["coreId"];
	        this.fingerprintArgs = source["fingerprintArgs"];
	        this.proxyId = source["proxyId"];
	        this.proxyConfig = source["proxyConfig"];
	        this.launchArgs = source["launchArgs"];
	        this.tags = source["tags"];
	        this.keywords = source["keywords"];
	        this.groupId = source["groupId"];
	    }
	}
	export class Settings {
	    userDataRoot: string;
	    defaultFingerprintArgs: string[];
	    defaultLaunchArgs: string[];
	    defaultStartUrls: string[];
	    lightStartEnabled: boolean;
	    restoreLastSession: boolean;
	    startReadyTimeoutMs: number;
	    startStableWindowMs: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userDataRoot = source["userDataRoot"];
	        this.defaultFingerprintArgs = source["defaultFingerprintArgs"];
	        this.defaultLaunchArgs = source["defaultLaunchArgs"];
	        this.defaultStartUrls = source["defaultStartUrls"];
	        this.lightStartEnabled = source["lightStartEnabled"];
	        this.restoreLastSession = source["restoreLastSession"];
	        this.startReadyTimeoutMs = source["startReadyTimeoutMs"];
	        this.startStableWindowMs = source["startStableWindowMs"];
	    }
	}
	export class Tab {
	    tabId: string;
	    title: string;
	    url: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Tab(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tabId = source["tabId"];
	        this.title = source["title"];
	        this.url = source["url"];
	        this.active = source["active"];
	    }
	}

}

export namespace config {
	
	export class BrowserBookmark {
	    name: string;
	    url: string;
	    openOnStart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BrowserBookmark(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.openOnStart = source["openOnStart"];
	    }
	}
	export class BrowserCore {
	    coreId: string;
	    coreName: string;
	    corePath: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BrowserCore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coreId = source["coreId"];
	        this.coreName = source["coreName"];
	        this.corePath = source["corePath"];
	        this.isDefault = source["isDefault"];
	    }
	}
	export class BrowserProxy {
	    proxyId: string;
	    proxyName: string;
	    proxyConfig: string;
	    groupName?: string;
	    sortOrder?: number;
	    lastLatencyMs: number;
	    lastTestOk: boolean;
	    lastTestedAt: string;
	    lastIPHealthJson?: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowserProxy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyId = source["proxyId"];
	        this.proxyName = source["proxyName"];
	        this.proxyConfig = source["proxyConfig"];
	        this.groupName = source["groupName"];
	        this.sortOrder = source["sortOrder"];
	        this.lastLatencyMs = source["lastLatencyMs"];
	        this.lastTestOk = source["lastTestOk"];
	        this.lastTestedAt = source["lastTestedAt"];
	        this.lastIPHealthJson = source["lastIPHealthJson"];
	    }
	}
	export class ProxyCheckTarget {
	    id: string;
	    name: string;
	    type: string;
	    url: string;
	    parser?: string;
	    timeoutMs?: number;
	    expectedStatus?: number[];
	
	    static createFrom(source: any = {}) {
	        return new ProxyCheckTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.parser = source["parser"];
	        this.timeoutMs = source["timeoutMs"];
	        this.expectedStatus = source["expectedStatus"];
	    }
	}
	export class ProxyCheckConfig {
	    prepareTimeoutMs: number;
	    speedTargetId: string;
	    ipHealthTargetId: string;
	    targets: ProxyCheckTarget[];
	
	    static createFrom(source: any = {}) {
	        return new ProxyCheckConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prepareTimeoutMs = source["prepareTimeoutMs"];
	        this.speedTargetId = source["speedTargetId"];
	        this.ipHealthTargetId = source["ipHealthTargetId"];
	        this.targets = this.convertValues(source["targets"], ProxyCheckTarget);
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

export namespace launchcode {
	
	export class LaunchRequestParams {
	    launchArgs: string[];
	    startUrls: string[];
	    skipDefaultStartUrls: boolean;
	    proxyId: string;
	    proxyConfig: string;
	
	    static createFrom(source: any = {}) {
	        return new LaunchRequestParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.launchArgs = source["launchArgs"];
	        this.startUrls = source["startUrls"];
	        this.skipDefaultStartUrls = source["skipDefaultStartUrls"];
	        this.proxyId = source["proxyId"];
	        this.proxyConfig = source["proxyConfig"];
	    }
	}

}

export namespace logger {
	
	export class MemoryLogEntry {
	    time: string;
	    level: string;
	    component: string;
	    message: string;
	    fields?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new MemoryLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.level = source["level"];
	        this.component = source["component"];
	        this.message = source["message"];
	        this.fields = source["fields"];
	    }
	}
	export class MethodInterceptor {
	
	
	    static createFrom(source: any = {}) {
	        return new MethodInterceptor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}


