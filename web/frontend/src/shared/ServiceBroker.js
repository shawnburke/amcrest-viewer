
import CamsServiceMock from "./mock/CamerasService";

import CamsService from "./CamerasService";


class ServiceBroker {

    constructor() {
        this.camsService = this.isDev() ? CamsServiceMock : CamsService;
    }


    isDev() {
        return process.env.NODE_ENV !== "production";
    }

    newCamsService() {
        return new this.camsService();
    }
}

export default ServiceBroker;