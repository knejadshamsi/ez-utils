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
	    id: number;
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
	        this.id = source["id"];
	        this.file_path = source["file_path"];
	        this.status = source["status"];
	        this.timestamp = source["timestamp"];
	        this.table_name = source["table_name"];
	        this.record_count = source["record_count"];
	    }
	}

}

