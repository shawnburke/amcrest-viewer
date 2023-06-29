import {GetData} from './file_data';

class FilesService {

    
    constructor(cam, root, files) {
        this.files = files;
    }

    async retrieveItems(startDate, endDate, sort) {
        console.log(`Retrieve ${startDate.iso()} => ${endDate.iso()} (${sort}) [supplied=${Boolean(this.files)}]`);

        var files = GetData(startDate, endDate, sort, this.files);
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
