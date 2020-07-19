import {Time, hour, minute, second, day} from "./time";

describe('Time', () => {

    var date = new Date("2020-04-30T01:23:45Z");

    it ("is created correctly", () => {

        var t = new Time(date);
        expect(t.unix).toEqual(date.getTime());

        t = new Time(date.toString());
        expect(t.unix).toEqual(date.getTime());

        t = new Time(date.getTime());
        expect(t.unix).toEqual(date.getTime());

     
        t = new Time(t);
        expect(t.unix).toEqual(date.getTime());
    });

    it("round trips correctly from date",() =>{

        var t = new Time(date);

        var unix = t.unix;

        var d2 = new Date(unix);

        expect(d2.getDate()).toEqual(date.getDate());
    })

    it ("round trips from unix", () => {
        var t= new Time(date.getTime());

        expect(t.iso()).toEqual(date.toISOString());

    });

    if ("offsets correctly", () => {

        var up = t.offset(2, hour);
        var down = t.offset(-2, minute);

        var dup = date.getTime() + (2*hour);

        expect(dup.unix).toEqual(up);

        var ddown = date.getTime() + (-2*minute);

        expect(ddows.unix).toEqual(down);

    });

    if ("compares correctly", () => {

        var t = new Time(date);
        var dafter = t.offset(1, hour);

        expect(dafter.after(t)).toEqual(true);
        expect(dafter.before(t)).toEqual(false);

        expect(t.same(date)).toEqual(true);

    });

    if ("boxes correctly", () => {
        var t = new Time(date);

        var tmin = new Time(new Date("2020-05-01"));
        var tmax = new Time(new Date("2020-05-04"));

        var t2 = t.box(tmin, tmax);

        expect(t2.same(tmin)).toEqual(true);

        tmin = tmin.offset(-7, day);
        tmax = tmax.offset(-7, day);

        t2 = t.box(tmin, tmax);

        expect(t2.same(tmax)).toEqual(true);
    });

    it ("round/floor correctly", () => {
        var d1 = new Date("2020-04-30T07:00:15");

        var t = new Time(d1);

        var tround = t.round(minute);
        expect(tround.date).toEqual(new Date("2020-04-30T07:00:00"));

        t = t.offset(30, second);

        tround = t.round(minute);

        expect(tround.date).toEqual(new Date("2020-04-30T07:01:00"));

        var tfloor = t.floor(day);

        expect(tfloor.date).toEqual(new Date("2020-04-30T00:00:00"));

        var tceil = t.ceil(day);
        expect(tceil.date).toEqual(new Date("2020-05-01T00:00:00"));

        var d2 = new Time("2020-04-30");

        expect(d2.floor(day).date).toEqual(d2.date);
        expect(d2.ceil(day).date).toEqual(d2.date);
    });

    it ("round/floor timezone correctly", () => {
      
        var t = new Time("2020-04-30T07:00:15-07:00");
        var tfloor = t.floor(day, 420);

        expect(tfloor.date).toEqual(new Date("2020-04-30T00:00:00-07:00"));


        var tround = t.round(minute);
        expect(tround.date).toEqual(new Date("2020-04-30T07:00:00-07:00"));

        t = t.offset(30, second);

        tround = t.round(minute);

        expect(tround.date).toEqual(new Date("2020-04-30T07:01:00-07:00"));

       
    });

});