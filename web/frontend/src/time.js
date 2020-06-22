


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
