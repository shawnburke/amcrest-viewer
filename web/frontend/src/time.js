

export const second = 1000;
export const minute = second * 60;
export const hour = 60 * minute;
export const day = hour * 24;

export const month = day * 30;
export const year = day * 365;

export class Range{
    constructor(start, end) {
        this.start = new Time(start);
        this.end = new Time(end);
    }

    contains(t) {
        t = new Time(t);
        return this.start.before(t, true) && this.end.after(t);
    }
}


export class Time {
    constructor(val) {

        if (!val) {
            val = new Date();
        }

        var type = typeof val;

        switch (type) {
            case "number":
                this.unix = val;
                this.date = new Date(val);
                break;
            case "string":
                this.date = new Date(val);
                this.unix = this.date.getTime();
                break;
            case "object":
                if (val instanceof Date) {
                    this.date = val;
                    this.unix = this.date.getTime();
                    break;
                }
                if (val instanceof Time) {
                    this.date = val.date;
                    this.unix = val.unix;
                    break;
                }
                // fall through
            default:
                throw new Error(`Unknown time value: ${val} (${typeof val} / ${val.constructor && val.constructor.name})`);
        }

        if (this.unix === 0 || 
            this.date == null || 
            Number.isNaN(this.unix) ||
             Number.isNaN(this.date.getTime()) ) {
            throw new Error(`Invalid Time init: ${val}`)
        }
    }

    iso() {
        return this.date.toISOString()
    }

    locale() {
        return this.date.toLocaleString()
    }

    localeTime() {
        return this.date.toTimeLocaleString();
    }

    after(t, inclusive) {
        t = new Time(t);

        if (inclusive) {
            return this.unix >= t.unix;
        }
        return this.unix > t.unix;
    }

    before (t, inclusive) {

        t = new Time(t);

        if (inclusive) {
            return this.unix <= t.unix;
        }
        
        return this.unix < t.unix;
    }

    same(t) {
        t = new Time(t);
        return t.unix === this.unix;
    }

    add(n, type) {
        return this.offset(n, type);
    }

    offset(n, type) {

        if (!n) {
            return this;
        }

        if (!type) {
            type = 1;
        }
        var start = this.unix;
    
        var amount = 0;

        switch (type) {
            case "hour":
                amount = hour;
                break;
            case "minute":
                amount = minute;
                break;
            case "second":
                amount = second;
                break;
            case "day":
                amount = day;
                break; 
            default:
                amount = Number(type);
                break;
        }

        start += n * amount;
        return new Time(start);
    }

    box(min, max) {
        min = new Time(min);
        max = new Time(max);

        if (this.before(min)) {
            return min;
        }

        if (this.after(max)) {
            return max;
        }
        return this;
    }

    round(type) {

        if (!type) {
            return this;
        }

        var half = type / 2;
        var floor = this.floor(type).unix;
        var mod = this.unix % type;
        
        if (mod <= half) {
            return new Time(floor);
        }
        return new Time(floor+type);
    }

    _getTimezoneOffset(tzMins) {

       return tzMins || this.date.getTimezoneOffset();
    }

    floor(type, tzOffsetMins) {
        if (!type) {
            return this;
        }

        var tz = this._getTimezoneOffset(tzOffsetMins) * minute;
        var unix = this.unix - tz;

        var mod = unix % type;
        var floor = unix - mod;
        return new Time(floor + tz);
    }

    ceil(type) {

        if (!type) {
            return this;
        }

        var floor = this.floor(type);

        // is there any delta?
        if (floor.unix < this.unix) {

            return floor.add(1, type);
        }

        return floor;
    }

    delta(t, type) {
        var d = this.unix - t.unix;

        if (type) {
            d /= type;
        }
        return Math.round(d);
    }
}

Time.now = function() { return new Time();}
Time.same = (d1, d2) => {

    var d1ms = d1 && d1.unix;
    var d2ms = d2 && d2.unix;

    return d1ms === d2ms;
}

Time.wrap = function(d) {
    if (d instanceof Time) {
        return d;
    }
    return new Time(d);
}



export function boxTime(t, min, max) {


    var tt = toUnix(t);
    var tmin = toUnix(min);
    var tmax = toUnix(max);

    tt = Math.max(tt, tmin);
    tt = Math.min(tt, tmax);

    return tt;
}

export function toUnix(t) {
    if (t.getTime) {
        return t.getTime()
    }

    return Number(t);
}

export function iso(t) {
    var d = new Date(t);
    return d.toISOString();
}

export function snapTime(t, unit, bias) {

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
            break;
    }

    t += offset;

    if (wasDate) {
        t = new Date(t);
    }
    return t;
}
