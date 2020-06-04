
import CamsServiceMock from "./mock/CamerasService";

import CamsService from "./CamerasService";


import FilesServiceMock from "./mock/FilesService";

import FilesService from "./FilesService";

class ServiceBroker {

    constructor() {
        this.camsService = this.useMock() ? CamsServiceMock : CamsService;
        this.filesService = this.useMock() ? FilesServiceMock : FilesService;
    }


    useMock() {
        return process.env.NODE_ENV !== "production" && this.root() === "";
    }

    root() {
        var r = process.env.REACT_APP_ROOT;
        return r || "";
    }

    newCamsService() {
        return new this.camsService(this.root());
    }

    newFilesService(cam) {
        return new this.filesService(cam, this.root())
    }
}

export default ServiceBroker;