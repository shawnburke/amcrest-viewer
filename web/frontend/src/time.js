

export const hour = 3600 * 1000;
export const day = hour * 24;
export const month = day * 30;

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


        if (val.unix) {
            this.unix = val.unix;
            this.date = new Date(val.unix);
            return;
        }
     

        if (val.getTime) {
            val = val.getTime();
        }
        this.unix = val;
        this.date = new Date(val);
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
            return t.unix >= this.unix;
        }
        return t.unix > this.unix;
    }

    before (t, inclusive) {

        t = new Time(t);

        if (inclusive) {
            return t.unix <= this.unix;
        }
        
        return t.unix < this.unix;
    }

    same(t) {
        t = new Time(t);
        return t.unix === this.unix;
    }

    offset(n, type) {
        var start = this.unix;

        switch (type) {
            case "hour":
                start += n * hour;
                break;
            case "minute":
                start += n * 1000*60;
                break;
            case "second":
                start += n * 1000;
                break;
            case "day":
                start += n * day;
                break; 
            default:
                start += n * Number(type);
                break;
        }
        return new Time(start);
    }
}