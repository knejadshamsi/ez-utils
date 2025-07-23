export namespace gui {
	
	export class BoundingBox {
	    north: number;
	    south: number;
	    east: number;
	    west: number;
	
	    static createFrom(source: any = {}) {
	        return new BoundingBox(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.north = source["north"];
	        this.south = source["south"];
	        this.east = source["east"];
	        this.west = source["west"];
	    }
	}
	export class Database {
	
	
	    static createFrom(source: any = {}) {
	        return new Database(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class LinkResult {
	    id: string;
	    from_node: string;
	    to_node: string;
	    raw_xml: string;
	
	    static createFrom(source: any = {}) {
	        return new LinkResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.from_node = source["from_node"];
	        this.to_node = source["to_node"];
	        this.raw_xml = source["raw_xml"];
	    }
	}
	export class ViewportBounds {
	    minLat: number;
	    maxLat: number;
	    minLng: number;
	    maxLng: number;
	
	    static createFrom(source: any = {}) {
	        return new ViewportBounds(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minLat = source["minLat"];
	        this.maxLat = source["maxLat"];
	        this.minLng = source["minLng"];
	        this.maxLng = source["maxLng"];
	    }
	}
	export class LoadingParams {
	    processId: number;
	    fileEditMode: string;
	    viewport: ViewportBounds;
	    randomFactor: number;
	    maxElements: number;
	
	    static createFrom(source: any = {}) {
	        return new LoadingParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processId = source["processId"];
	        this.fileEditMode = source["fileEditMode"];
	        this.viewport = this.convertValues(source["viewport"], ViewportBounds);
	        this.randomFactor = source["randomFactor"];
	        this.maxElements = source["maxElements"];
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
	export class NodeResult {
	    id: string;
	    lng: number;
	    lat: number;
	    raw_xml: string;
	
	    static createFrom(source: any = {}) {
	        return new NodeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.lng = source["lng"];
	        this.lat = source["lat"];
	        this.raw_xml = source["raw_xml"];
	    }
	}
	export class PTDeparture {
	    id: string;
	    route_id: string;
	    departure_time: string;
	
	    static createFrom(source: any = {}) {
	        return new PTDeparture(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.route_id = source["route_id"];
	        this.departure_time = source["departure_time"];
	    }
	}
	export class PTLine {
	    id: string;
	    mode: string;
	    raw_xml: string;
	
	    static createFrom(source: any = {}) {
	        return new PTLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.mode = source["mode"];
	        this.raw_xml = source["raw_xml"];
	    }
	}
	export class PTLineSummary {
	    id: string;
	    name: string;
	    mode: string;
	    route_count: number;
	    departure_count: number;
	
	    static createFrom(source: any = {}) {
	        return new PTLineSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mode = source["mode"];
	        this.route_count = source["route_count"];
	        this.departure_count = source["departure_count"];
	    }
	}
	export class PTRoute {
	    id: string;
	    line_id: string;
	    raw_xml: string;
	
	    static createFrom(source: any = {}) {
	        return new PTRoute(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.line_id = source["line_id"];
	        this.raw_xml = source["raw_xml"];
	    }
	}
	export class PTRouteStop {
	    route_id: string;
	    stop_ref_id: string;
	    stop_order: number;
	    arrival_offset: string;
	    departure_offset: string;
	
	    static createFrom(source: any = {}) {
	        return new PTRouteStop(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.route_id = source["route_id"];
	        this.stop_ref_id = source["stop_ref_id"];
	        this.stop_order = source["stop_order"];
	        this.arrival_offset = source["arrival_offset"];
	        this.departure_offset = source["departure_offset"];
	    }
	}
	export class PTStop {
	    id: string;
	    lng: number;
	    lat: number;
	    name: string;
	    raw_xml: string;
	
	    static createFrom(source: any = {}) {
	        return new PTStop(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.lng = source["lng"];
	        this.lat = source["lat"];
	        this.name = source["name"];
	        this.raw_xml = source["raw_xml"];
	    }
	}
	export class PTTelemetry {
	    process_id: number;
	    total_file_size: number;
	    bytes_read: number;
	    stops_extracted: number;
	    lines_extracted: number;
	    routes_extracted: number;
	    error_count: number;
	    last_updated: string;
	
	    static createFrom(source: any = {}) {
	        return new PTTelemetry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.process_id = source["process_id"];
	        this.total_file_size = source["total_file_size"];
	        this.bytes_read = source["bytes_read"];
	        this.stops_extracted = source["stops_extracted"];
	        this.lines_extracted = source["lines_extracted"];
	        this.routes_extracted = source["routes_extracted"];
	        this.error_count = source["error_count"];
	        this.last_updated = source["last_updated"];
	    }
	}
	export class Person {
	    id: string;
	    coords: string;
	    raw_xml: string;
	
	    static createFrom(source: any = {}) {
	        return new Person(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.coords = source["coords"];
	        this.raw_xml = source["raw_xml"];
	    }
	}
	export class PaginatedResponse {
	    persons: Person[];
	    totalCount: number;
	    currentPage: number;
	    totalPages: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.persons = this.convertValues(source["persons"], Person);
	        this.totalCount = source["totalCount"];
	        this.currentPage = source["currentPage"];
	        this.totalPages = source["totalPages"];
	        this.pageSize = source["pageSize"];
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
	
	export class Point {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new Point(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class Process {
	    process_id: number;
	    file_path: string;
	    status: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Process(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.process_id = source["process_id"];
	        this.file_path = source["file_path"];
	        this.status = source["status"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
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
	export class ProcessResult {
	    process_id: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.process_id = source["process_id"];
	        this.message = source["message"];
	    }
	}
	export class ProcessTelemetry {
	    process_id: number;
	    total_file_size: number;
	    bytes_read: number;
	    persons_extracted: number;
	    nodes_read: number;
	    links_read: number;
	    error_count: number;
	    // Go type: time
	    last_updated: any;
	
	    static createFrom(source: any = {}) {
	        return new ProcessTelemetry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.process_id = source["process_id"];
	        this.total_file_size = source["total_file_size"];
	        this.bytes_read = source["bytes_read"];
	        this.persons_extracted = source["persons_extracted"];
	        this.nodes_read = source["nodes_read"];
	        this.links_read = source["links_read"];
	        this.error_count = source["error_count"];
	        this.last_updated = this.convertValues(source["last_updated"], null);
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
	export class ValidationResult {
	    isValid: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isValid = source["isValid"];
	        this.error = source["error"];
	    }
	}
	
	export class Zone {
	    id: string;
	    name: string;
	    polygon: Point[];
	    // Go type: struct { MinX float64 "json:\"minX\""; MinY float64 "json:\"minY\""; MaxX float64 "json:\"maxX\""; MaxY float64 "json:\"maxY\"" }
	    boundingBox: any;
	
	    static createFrom(source: any = {}) {
	        return new Zone(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.polygon = this.convertValues(source["polygon"], Point);
	        this.boundingBox = this.convertValues(source["boundingBox"], Object);
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
	export class ZoneCount {
	    zoneId: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new ZoneCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.zoneId = source["zoneId"];
	        this.count = source["count"];
	    }
	}

}

