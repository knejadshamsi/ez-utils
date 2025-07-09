export namespace database {
	
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
	export class Process {
	    process_id: number;
	    file_path: string;
	    status: string;
	    timestamp: string;
	    table_name: string;
	    record_count: number;
	
	    static createFrom(source: any = {}) {
	        return new Process(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.process_id = source["process_id"];
	        this.file_path = source["file_path"];
	        this.status = source["status"];
	        this.timestamp = source["timestamp"];
	        this.table_name = source["table_name"];
	        this.record_count = source["record_count"];
	    }
	}
	export class ProcessTelemetry {
	    process_id: number;
	    total_file_size: number;
	    bytes_read: number;
	    persons_extracted: number;
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

}

