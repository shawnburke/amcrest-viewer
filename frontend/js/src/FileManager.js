import { hour, month, day, second, Time } from './time';

// FileManager takes in a list of files
// and manages them by time
//
// Range: 
//      Min: minimum date for available files
//      Max: maximum date for available files
//
// Window: Range.Min < Window < Range.Max
//      Start/End: Time range of files currently 
//                 in scope, e.g. a given day
//      Change: Sets current to media file closest to 
//              current time
//
// Position: Window.Start < Position < Window.End
//      Change: selects a file
//              If outside window, may bump the window?
//
// Current: File that contains the Postition time.  Setting this sets Position to beginning of file
//
// Events:
//      OnCurrentFileChange
//      OnWindowChange
//      OnPositionChange
// 
export class FileManager {

    constructor(camid, broker) {
        this.camid = camid;
        this.filesService = broker.newFilesService(camid);
        this.camerasServer = broker.newCamsService();

        this.maxWindow = day;
        this.refreshIntervalSeconds = 15;
        this.state = {}

        // initialize info
        var today = new Time();
        this.state.range = {
            min: today.add(-1, month),
            max: today,
        }
        
        this.state.window = {
            start: today.add(-1, day),
            end: today,
        }
    
        this.state.position = today.add(-1, hour);
    }

   

    log(s) {
        console.log(s);
    }

    async start() {
        return this.camerasServer.getStats(this.camid).then(
            s => {

                if (!s) {
                    // in case of server error
                    console.error(`Call to server cameras/${this.camid}/stats failed.`)
                    return;
                }

                return this._setStats(s, this.refreshIntervalSeconds * 1000);
            }
        );
    }

    _setStats(stats, refreshInterval) {
        if (stats.canLiveStream === false) {
            this.liveDisabled = true;
        } else if (!this._streamingWarmup) {
            // kick live so it's fast if user tries it.
            var start = new Time();
            this.camerasServer.getLiveStreamUrl(this.camid).then(uri => {
                this.log(`Live streaming ready @ ${uri} (${new Time().delta(start, second)}s)`)
                this._streamingWarmup = true;
            });
        }

        var promise = this.setRange(stats.min_date, stats.max_date);

        if (refreshInterval) {
            setTimeout(() => {
                this.log("Timed range refresh");
                try {
                    this._refreshing = true;
                    this.start();
                }
                finally {
                    this._refreshing = false;
                }
            }, refreshInterval);
        }
        return promise;
    }


    loadFiles(start, end) {

        // snap them to start / end on midnight and 1s before next
        // midnight
        start = start.iso().replace(/\d{2}:\d{2}:\d{2}/, "00:00:00")
        start = new Time(start);

        end = end || start.add(-1, day);

        return this.filesService.retrieveItems(start, end, "");
    }


    _startBatch() {

        var batch = this._batch;

        if (!batch) {
            batch = {
                count: 0,
            }
            this._batch = batch;
        }

        if (++batch.count === 1) {
            batch.value = {}
            batch.state = Object.assign({}, this.state)
            batch.promise = new Promise((resolve) => {
                batch.resolve = resolve;
            });
        }
        return batch.promise;
    }

    _endBatch() {

        if (!this._batch || this._batch.count === 0) {
            console.error("Too many batch ends");
            return;
        }

        var promise = this._batch.promise;
        var resolve = this._batch.resolve;
        if (--this._batch.count === 0) {
            this._onchange(this._batch.value, this._batch.state);
            delete this._batch.value;
            delete this._batch.promise;
            resolve(this._batch.state, this.state);
        }
        return promise;
    }

    _onchange(value, old) {

        if (!value) {
            return;
        }

        var batching = this._batch && this._batch.count > 0;

        var printSpan = function(s,e) {
            return `${s.iso()} => ${e.iso()}`;
        }

        var rangeChange = value.range;
        if (rangeChange && !batching) {
            this.log(`Range:\n\tOld: ${printSpan(old.range.min, old.range.max)}\n\tNew: ${printSpan(value.range.min, value.range.max)}`);
        }
        var windowChange = value.window;
        if (windowChange && !batching) {
            this.log(`Window:\n\tOld: ${printSpan(old.window.start, old.window.end)}\n\tNew: ${printSpan(value.window.start, value.window.end)}`)
        }
        var positionChange = value.position !== undefined;
        if (positionChange && !batching) {
            this.log(`Pos:\n\tOld: ${this.state.position && this.state.position.iso()}\n\tNew: ${value.position.iso()}`);
        }

        var fileChange = value.file !== undefined;
        if (fileChange && !batching) {
            if (value.file) {
                this.log(`Setting file to ${value.file.id} ${value.file.path}`);
            } else {
                this.log(`Clearing file`);
            }
        }

        Object.assign(this.state, value);

        if (batching) {
            Object.assign(this._batch.value, value);
            return;
        }

        if (this.onChange) {
            this.onChange(value);
        };
    }

    getState() {
        var state = Object.assign({
            fileCount: this.state.files && (this.state.files.length || 0),
            live: this.isLive(),
        }, this.state);
        return state;
    }


    isLive() {
        return this.state.file && this.state.file.type === 2;
    }


    async startLive() {

        if (this.isLive()) {
            return Promise.resolve(true);
        }

        if (this.liveDisabled) {
            console.log(`Live not supported by ${this.camid}`)
            return Promise.resolve(false);
        }

        console.log(`Initiating live view for ${this.camid}`)

        return this.camerasServer.getLiveStreamUrl(this.camid).then(uri => {

            try {
                this._startBatch();

                // for debugging
                if (uri === "prompt") {
                    uri = window.prompt("Enter streaming source");
                    if (uri === "") {
                        return false;
                    }
                }

                console.log(`Live URL: ${uri}`)

                var file = {
                    type: 2,
                    path: uri,

                }

                this.setPosition(this.state.window.end, file);
                this.setCurrentFile(file);

            } finally {
                this._endBatch();
            }
            return true;
        });
    }

    stopLive() {
        if (!this.isLive()) {
            return
        }

        try {
            this._startBatch();
            console.log(`Stopping live `)
            this.selectLastFile();
        }
        finally {
            this._endBatch();
        }

    }

    setRange(min, max) {
        if (max < min) {
            throw new Error("bad range");
        }

        var oldMin = this.state.range && this.state.range.min;
        var oldMax = this.state.range && this.state.range.max;

        if (min.same(oldMin) && max.same(oldMax)) {
            return Promise.resolve();
        }

        var promise = Promise.resolve();
        try {
            this._startBatch();

            this._onchange({
                range: {
                    min: min,
                    max: max,
                },
            });

           promise = this.setWindow(this.state.window.start, this.state.window.end, true);
        } finally {
           promise = Promise.all([promise, this._endBatch()]);
        }
        return promise;
    }

    setWindow(start, end, reload) {

        if (end < start) {
            throw new Error("bad window");
        }

        var boxedStart = start.box(this.state.range.min, this.state.range.max);   
        var boxedEnd = end.box(this.state.range.min, this.state.range.max);

        boxedStart = boxedStart.floor(day);
        boxedEnd = boxedEnd.ceil(day);

        if (!reload && boxedStart.same(this.state.window.start) && boxedEnd.same(this.state.window.end)) {
            return Promise.resolve(true);
        }

        var windowSize = boxedEnd.delta(boxedStart);
        if (windowSize > this.maxWindow) {
            boxedStart = boxedEnd.add(-1 * this.maxWindow);
        }

        try {
            this._startBatch();

            this._onchange({
                window: {
                    start: boxedStart,
                    end: boxedEnd,
                }
            });

        } finally {
            var promise = this._endBatch();

            promise = promise.then(() => {
                return this.refreshFiles(boxedStart, boxedEnd);
            });
        }
        return promise;
    }

    isInFile(time, file) {
        return file.start.before(time, true) && file.end.after(time);
    }

    setPosition(time, file) {

        var boxed = time && time.box(this.state.window.start, this.state.window.end);
 
        if (boxed.same(this.state.position)) {
            return Promise.resolve(true);
        }

        var promise = null;
        try {
            this._startBatch();
            this._onchange({ position: boxed });

            // if this position is in a file, set that file
            //
            if (!file && this.state.files && boxed) {
                file = this.state.files.find(f => this.isInFile(boxed, f));
            }
            promise = this.setCurrentFile(file);
        }
        finally {
            return Promise.all([promise, this._endBatch()]);
        }
       
    }

    selectLastFile(pos) {

       
        var lastFile = null;

        // find the last file
        var files = this.state.files || [];
        if ((!pos || (!pos.same(this.state.position) && files.length))) {
            this.log(`Selecting last file`);
            lastFile = files[files.length - 1];
            pos = lastFile.timestamp;
        }

        return this.setPosition(pos, lastFile);
    }

    refreshFiles(start, end) {
        this.log(`Loading files for range ${start.iso()} => ${end.iso()}`)
        return this.loadFiles(start, end).then(items => {

            this.log(`Loaded ${items.length} files`);

            var currentFileId = this.state.file && this.state.file.id;
            var targetPos = null;

            // sort ascending
            var files = items.sort((a, b) => a.timestamp.unix - b.timestamp.unix);

            // set an end for each file item;
            files.forEach(file => {
                if (!file.end) {
                    var end = file.timestamp

                    if (file.duration_seconds) {
                        end = file.timestamp.add(file.duration_seconds, second)
                    } else {
                        end.add(1, second);
                    }
                    file.start = file.timestamp;
                    file.end = end;
                }

                if (currentFileId === file.id) {
                    targetPos = this.state.position;
                    if (!this.isInFile(this.state.position, file)) {
                        targetPos = file.timestamp;
                    }
                }
                   
            });

            try {
                var promise = this._startBatch();

                this._onchange({ files: files });

                var liveRefresh = this.isLive() && this._refreshing;

                if (!liveRefresh) {
                    promise.then(() => this.selectLastFile(targetPos));
                }
            }
            finally {
                return Promise.all([promise,this._endBatch()]);
            }

        })
    }


    setCurrentFile(file) {
        var oldid = this.state.file && this.state.file.id;
        var newid = file && file.id;
        if (oldid === newid) {
            return Promise.resolve(true);
        }

        if (newid) {
            file = this.state.files.find(f => f.id === newid);
            if (!file) {
                console.warn(`Can't find file ${newid}`);
                return Promise.resolve(false);
            }
        }

        var update = {
            file: file || null
        }

        var promise = Promise.resolve(true);

        var isLive = file && file.type === 2;

        if (isLive) {
            update.position = this.state.window.end;
        } else if (file) {
            // set position if not in file
            var boxed = file.timestamp.box( this.state.window.start, this.state.window.end);

            if (!boxed.same( file.timestamp)) {
                console.warn(`Selected file timestamp ${file.timestamp} outside of window ${this.state.window.start} => ${this.state.window.end}`);
                return Promise.resolve(false);
            }

            promise = this.setPosition(file.timestamp, file);
        } 

        this._onchange(update);
        return promise;
    }

  

}