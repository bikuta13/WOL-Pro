export namespace app {

	export class Device {
	    id: string;
	    name: string;
	    mac_address: string;
	    ip_address: string;

	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mac_address = source["mac_address"];
	        this.ip_address = source["ip_address"];
	    }
	}
	export class Result {
	    success: boolean;
	    message: string;
	    id?: string;
	    mac?: string;

	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.id = source["id"];
	        this.mac = source["mac"];
	    }
	}

}

