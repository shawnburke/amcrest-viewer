
import CamsServiceMock from "./mock/CamerasService";

import CamsService from "./CamerasService";


import FilesServiceMock from "./mock/FilesService";

import FilesService from "./FilesService";

class ServiceBroker {

    constructor(camsService, filesService) {

        if (camsService === true) {
            this.mock = true;
        } else {
            this.cs = camsService;
            this.fs = filesService;
        }
        this.camsServiceCtor = this.useMock() ? CamsServiceMock : CamsService;
        this.filesServiceCtor = this.useMock() ? FilesServiceMock : FilesService;
    }


    useMock() {
        return this.mock || (process.env.NODE_ENV !== "production" && this.root() === "");
    }

    root() {
        var r = process.env.REACT_APP_ROOT;
        return r || "";
    }

    newCamsService() {
        return this.cs || new this.camsServiceCtor(this.root());
    }

    newFilesService(cam) {
        return this.fs || new this.filesServiceCtor(cam, this.root())
    }
}

export default ServiceBroker;