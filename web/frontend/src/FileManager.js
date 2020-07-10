import { hour, month, day, toUnix, Time } from './time';

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


        // initialize info
        var today = new Date();
        this.state = {}
        this.state.range = {
            min: this.dateAdd(today, -1, "month"),
            max: today,
        }

        this.state.window = {
            start: this.dateAdd(today, -1, "day"),
            end: today,
        }

        this.state.position = this.dateAdd(today, -1, "hour");
        this.timify();
    }

    timify() {
        // if (this.state.range) {
        //     this.state.range.tmin = new Time(this.state.range.min);
        //     this.state.range.tmax = new Time(this.state.range.tmax);
        // }

        // if (this.state.window) {
        //     this.state.window.tstart = new Time(this.state.window.start);
        //     this.state.window.tend = new Time(this.state.window.end);
        // }

        // if (this.state.position) {
        //     this.tposition = new Time(this.state.position);
        // }
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
        } else {
            // kick live so it's fast if user tries it.
            var start = Date.now();
            this.camerasServer.getLiveStreamUrl(this.camid).then(uri => {
                this.log(`Live streaming ready @ ${uri} (${(new Date().getTime() - start) / 1000}s)`)
            });
        }

        var promise = this.setRange(new Date(stats.min_date), new Date(stats.max_date));

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
        start = start.toString().replace(/\d{2}:\d{2}:\d{2}/, "00:00:00")
        start = new Date(start);

        end = end || new Date(start.getTime() + (24 * 60 * 60 * 1000) - 1);


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

        var old = old || this.state;


        var printSpan = function (s, e) {
            return `${new Time(s).iso()} => ${new Time(e).iso()}`;
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
            this.log(`Pos:\n\tOld: ${new Time(old.position).iso()}\n\tNew: ${new Time(value.position).iso()}`);
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
        this.timify();

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
            fileCount: this.state.files && this.state.files.length || 0,
            live: this.isLive(),
        }, this.state);
        delete state.files;
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
            this.setPosition(new Date());
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

        if (this.timeEqual(min, this.state.range.min) && this.timeEqual(max, this.state.range.max)) {
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
            this._endBatch();
        }
        return promise;
    }

    setWindow(start, end, reload) {

        if (end < start) {
            throw new Error("bad window");
        }

        var boxedStart = this.boxTime(start, this.state.range.min, this.state.range.max, "min");
        var boxedEnd = this.boxTime(end, this.state.range.min, this.state.range.max, "max");

        boxedStart = this.snapTime(boxedStart, "day", -1);
        boxedEnd = this.snapTime(boxedEnd, "day", 1);

        if (!reload && this.timeEqual(boxedStart, this.state.window.start) && this.timeEqual(boxedEnd, this.state.window.end)) {
            return Promise.resolve(true);
        }

        var windowSize = toUnix(boxedEnd) - toUnix(boxedStart);
        if (windowSize > this.maxWindow) {
            boxedStart = new Date(toUnix(boxedEnd) - this.maxWindow);
        }

        try {
            var promise = this._startBatch();

            this._onchange({
                window: {
                    start: new Date(boxedStart),
                    end: new Date(boxedEnd),
                }
            });

            promise.then(() => {
                return this.refreshFiles(boxedStart, boxedEnd);
            });

        } finally {
            this._endBatch();
        }
        return promise;
    }

    isInFile(time, file) {
        time = toUnix(time);
        return time >= toUnix(file.start) && time <= toUnix(file.end);
    }

    setPosition(time, file) {

        var boxed = time && this.boxTime(time, this.state.window.start, this.state.window.end);

        if (this.timeEqual(boxed, this.state.position)) {
            return Promise.resolve(true);
        }

        var promise = Promise.resolve(true);
        try {
            this._startBatch();
            this._onchange({ position: boxed && new Date(boxed) });

            // find the file
            if (!file && this.state.files && boxed) {
                file = this.state.files.find(f => this.isInFile(boxed, f));
            }
            promise = this.setCurrentFile(file);
        }
        finally {
            this._endBatch();
        }
        return promise;
    }

    selectLastFile(pos) {

        var lastFile = null;

        // find the last file
        var files = this.state.files || [];
        if ((!pos || !this.timeEqual(pos, this.state.position)) && files.length) {
            lastFile = files[files.length - 1];
            pos = lastFile.timestamp;
        }

        this.setPosition(pos, lastFile);
    }

    refreshFiles(start, end) {
        this.log(`Loading files for range ${start} => ${end}`)
        return this.loadFiles(start, end).then(items => {

            this.log(`Loaded ${items.length} files`);

            // sort ascending
            var files = items.sort((a, b) => toUnix(a.timestamp) - toUnix(b.timestamp));

            // set an end for each file item;
            files.forEach(file => {
                if (!file.end) {
                    var end = toUnix(file.timestamp);

                    if (file.duration_seconds) {
                        end += 1000 * file.duration_seconds;
                    } else {
                        end += 1000;
                    }
                    file.start = file.timestamp;
                    file.end = end;
                }
            });

            try {
                this._startBatch();

                this._onchange({ files: files });

                var pos = this.state.position && this.boxTime(this.state.position, start, end, "max");
                var liveRefresh = this.isLive() && this._refreshing;

                if (!liveRefresh) {
                    this.selectLastFile(pos);
                }
            }
            finally {
                this._endBatch();
            }

        })
    }


    setCurrentFile(file) {
        var oldid = this.file && this.file.id;
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

        if (file && file.type !== 2) {
            // set position if not in file
            var boxed = this.boxTime(file.timestamp, this.state.window.start, this.state.window.end);

            if (!this.timeEqual(boxed, file.timestamp)) {
                console.warn(`Selected file timestamp ${file.timestamp} outside of window ${this.state.window.start} => ${this.state.window.end}`);
                return Promise.resolve(false);
            }

            promise = this.setPosition(file.timestamp, file);
        } else {
            update.position = this.state.window.end;
        }

        this._onchange(update);
        return promise;
    }

    //
    // Utility Functions
    //
    timeEqual(t1, t2) {
        if (t1 === t2) {
            return true;
        }

        if (!t1 || !t2) {
            return false;
        }

        t1 = toUnix(t1);
        t2 = toUnix(t2);

        return t1 === t2;
    }


    dateAdd(date, n, unit) {


        var base = 0;

        switch (unit) {
            case "hour":
                base = hour;
                break;
            case "day":
                base = day;
                break;
            case "month":
                base = month;
                break;
            default:
                throw new Error("Unknown unit: " + unit);
        }

        base *= n;

        if (!date) {
            date = new Date();
        }

        return new Date(date.getTime() + base);
    }


    // todo: migrate to time lib
    snapTime(t, unit, bias) {

        var wasDate = false;
        if (t.getTime) {
            t = t.getTime();
            wasDate = true;
        }

        var chunk = 0;


        var offset = 0;

        switch (unit) {
            case "hour":
                chunk = hour;
                break;
            case "day":
                chunk = day;
                offset = new Date().getTimezoneOffset() * 60 * 1000;
                t -= offset;
                break;
            default:
                throw new Error("Unknown unit: " + unit);
        }

        var delta = t % chunk;

        if (!bias) {
            bias = (delta < chunk / 2) ? -1 : 1;
        }

        switch (bias) {
            case -1:
                t -= delta;
                break;
            case 1:
                t += (chunk - delta);
                break;
            default:
        }

        t += offset;

        if (wasDate) {
            t = new Date(t);
        }
        return t;
    }

    boxTime(t, min, max, bias) {


        var tt = toUnix(t);
        var wasDate = tt !== t;
        var tmin = toUnix(min);
        var tmax = toUnix(max);

        var unboxed = tt < tmin || tt > tmax;

        if (unboxed) {

            switch (bias) {
                case "min":
                    tt = tmin;
                    break;
                case "max":
                    tt = tmax;
                    break;
                default:
                    if (tt < tmin) {
                        tt = tmin;
                    } else if (tt > tmax) {
                        tt = tmax;
                    }
            }
        }

        if (wasDate) {
            return new Date(tt);
        }
        return tt;
    }

}