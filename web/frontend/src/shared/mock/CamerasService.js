
import {FileData} from "./FilesService";

class CamerasService {

    constructor() {

        this.cameras = [
            { name: "Garage Cam", type: "amcrest", id: "amcrest-1" },
            { name: "Front Cam", type: "amcrest", id: "amcrest-2" },
        ];

    }

    async retrieveItems() {

        return Promise.resolve(this.cameras);

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

        var res =
        {
            min_date: FileData[0].timestamp,
            max_date: FileData[FileData.length - 1].timestamp,
            file_count: 0,
            file_size: 0
        }

        FileData.forEach(v => {
            res.file_count++;
            res.file_size += v.length;
        })

      
        return Promise.resolve(res);

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