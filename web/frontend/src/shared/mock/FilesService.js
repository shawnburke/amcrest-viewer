import {GetData} from './file_data';

class FilesService {

    

    async retrieveItems(startDate, endDate, sort) {
        console.log(`Retrieve ${startDate} => ${endDate} (${sort})`);

        var files = GetData(startDate, endDate, sort);
        return Promise.resolve(files);

    }

    async getItem(id) {

        for (var i = 0; i < this.cameras.length; i++) {

            if (this.files[i].id === id) {
                return Promise.resolve(this.files[i]);
            }

        }

        return null;

    }

}

export default FilesService;
