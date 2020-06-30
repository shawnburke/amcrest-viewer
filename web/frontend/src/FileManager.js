import { hour, month, day, toUnix } from './time';

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

        // initialize info
        var today = new Date();
        this.range = {
            min: this.dateAdd(today, -1, "month"),
            max: today,
        }

        this.window = {
            start: this.dateAdd(today, -1, "day"),
            end: today,
        }

        this.position = this.dateAdd(today, -1, "hour");
    }

    log(s) {
        console.log(s);
    }

    start() {
        this.camerasServer.getStats(this.camid).then(
            s => {
                this.setRange(new Date(s.min_date), new Date(s.max_date));
            }
        );
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
            this._onchange(this._batch.value);
            delete this._batch.value;
            delete this._batch.promise;
            resolve();
        }
        return promise;
    }

    _onchange(value) {

        if (!value) {
            return;
        }

        var batching = this._batch && this._batch.count > 0;

        var rangeChange = value.range;
        if (rangeChange && !batching) {
            this.log(`Changing range to ${value.range.min} => ${value.range.max}`);
        }
        var windowChange = value.window;
        if (windowChange && !batching) {
            this.log(`Changing window to ${value.window.start} => ${value.window.end}`)
        }
        var positionChange = value.position;
        if (positionChange && !batching) {
            this.log(`Setting position to ${value.position}`);
        }

        var fileChange = value.file;
        if (fileChange && !batching) {
            this.log(`Setting file to ${value.file.id} ${value.file.path}`)
        }

        Object.assign(this, value);

        if (batching) {
            Object.assign(this._batch.value, value);
            return;
        }

        if (this.onChange) {
            this.onChange(value);
        };
    }



    setRange(min, max) {
        if (max < min) {
            throw new Error("bad range");
        }

        if (min === this.range.min && max === this.range.max) {
            return;
        }

        try {
            this._startBatch();

            this._onchange({
                range: {
                    min: min,
                    max: max,
                },
            });

            this.setWindow(this.window.start, this.window.end);
        } finally {
            this._endBatch();
        }
    }

    setWindow(start, end) {

        if (end < start) {
            throw new Error("bad window");
        }

        var boxedStart = this.boxTime(start, this.range.min, this.range.max, "min");
        var boxedEnd = this.boxTime(end, this.range.min, this.range.max, "max");

        boxedStart = this.snapTime(boxedStart, "day", -1);
        boxedEnd = this.snapTime(boxedEnd, "day", 1);

        if (boxedStart === this.window.start && boxedEnd === this.window.end) {
            return true;
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
                this.refreshFiles(boxedStart, boxedEnd);
            });

        } finally {
            this._endBatch();
        }

    }

    isInFile(time, file) {
        time = toUnix(time);
        return time >= toUnix(file.start) && time <= toUnix(file.end);
    }

    setPosition(time, file) {

        var boxed = this.boxTime(time, this.window.start, this.window.end, 'min');

        if (this.timeEqual(boxed, this.position)) {
            return true;
        }

        try {
            this._startBatch();

            this._onchange({ position: new Date(boxed) });

            // find the file
            if (!file && this.files) {
                file = this.files.find(f => this.isInFile(time, f));
            }

            if (file) {
                this.setCurrentFile(file);
            }
        }
        finally {
            this._endBatch();
        }
    }

    refreshFiles(start, end) {
        this.log(`Loading files for range ${start} => ${end}`)
        this.loadFiles(start, end).then(items => {

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

                var pos = this.boxTime(this.position, start, end);
                var lastFile = null;

                // find the last file
                if (pos !== this.position && files.length) {
                    lastFile = files[files.length - 1];

                    pos = lastFile.timestamp;
                }

                this.setPosition(pos, lastFile);
            }
            finally {
                this._endBatch();
            }

        })
    }

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

    setCurrentFile(file) {
        var oldid = this.file && this.file.id;
        var newid = file && file.id;
        if (oldid === newid) {
            return true;
        }

        file = this.files.find(f => f.id === newid);

        if (!file) {
            console.warn(`Can't find file ${newid}`);
            return false;
        }

        // set position if not in file
        var boxed = this.boxTime(file.timestamp, this.window.start, this.window.end);

        if (!this.timeEqual(boxed, file.timestamp)) {
            console.warn(`Selected file timestamp ${file.timestamp} outside of window ${this.window.start} => ${this.window.end}`);
            return false;
        }

        this.setPosition(file.timestamp, file);

        this._onchange({ file: file });
        return true;

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