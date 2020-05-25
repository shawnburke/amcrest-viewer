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