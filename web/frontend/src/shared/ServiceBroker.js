
import CamsServiceMock from "./mock/CamerasService";

import CamsService from "./CamerasService";


import FilesServiceMock from "./mock/FilesService";

import FilesService from "./FilesService";

class ServiceBroker {

    constructor() {
        this.camsService = this.isDev() ? CamsServiceMock : CamsService;
        this.filesService = this.isDev() ? FilesServiceMock : FilesService;
    }


    isDev() {
        return process.env.NODE_ENV !== "production";
    }

    newCamsService() {
        return new this.camsService();
    }

    newFilesService(cam) {
        return new this.filesService(cam)
    }
}

export default ServiceBroker;