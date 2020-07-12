
import { GetStats } from "./file_data.js";
import { Time } from "../../time.js";

class CamerasService {

    constructor() {


        var latest = {
            "id": 7156,
            "camera_id": 1,
            "path": "/1591027084.jpg",
            "type": 0,
            "timestamp": new Time("2020-06-16T03:07:09Z"),
            "length": 32503
        };
        this.cameras = [
            { name: "Garage Cam", type: "amcrest", id: "amcrest-1", latest_snapshot: latest },
            { name: "Front Cam", type: "amcrest", id: "amcrest-2" },
        ];

    }

    async retrieveItems() {

        return Promise.resolve(this.cameras);

    }


    async getLiveStreamUrl(id) { 
        return Promise.resolve("prompt");
    }

    async getItem(id) {

        for (var i = 0; i < this.cameras.length; i++) {

            if (this.cameras[i].id === id) {

                return Promise.resolve(this.cameras[i]);

            }

        }

        return null;

    }

    async getStats(id) {



        return Promise.resolve(GetStats());

    }

    // async createItem(item) {

    //     console.log("ItemService.createItem():");

    //     console.log(item);

    //     return Promise.resolve(item);

    // }

    // async deleteItem(itemId) {

    //     console.log("ItemService.deleteItem():");

    //     console.log("item ID:" + itemId);

    // }

    // async updateItem(item) {

    //     console.log("ItemService.updateItem():");

    //     console.log(item);

    // }

}

export default CamerasService;