export namespace main {
	
	export class UpdateInfo {
	    hasUpdate: boolean;
	    version: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.version = source["version"];
	        this.url = source["url"];
	    }
	}

}

