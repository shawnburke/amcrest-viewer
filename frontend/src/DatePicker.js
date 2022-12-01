import React from 'react';

import { Time, day, minute } from './time';

class DatePicker extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            date: Time.now(),
        }
    }

  
    shouldComponentUpdate(nextProps, _nextState) {
        if (!Time.same(nextProps.minDate, this.props.maxDate) || !Time.same(nextProps.maxDate, this.props.maxDate)) {
            this._controller = null;
        }
        return true;
    }
   

    
    ensureController() {

        var d =  this.state.date || this.props.date || Time.now();
        d = Time.wrap(d);
        if (!this._controller) {
            var minDate = this.props.minDate;
            var maxDate = this.props.maxDate;
            this._controller = new datePickerController(minDate, maxDate,d);
            this._controller.change = (date) => {
                setTimeout( () => {
                    this.setDate(date);
                }, 1);
                }
        }
        this._controller.date(d);
        return this._controller;
    }

    setDate(d) {

        if (d.same(this.state.date)) {
            return;
        }

        this.setState({
            date: d
        });

        this.props.onChange && this.props.onChange(d);
       
    }

   
    render() {
        var d = this.state.date;
        if (!d) {
            return <div></div>;
        }

       var c=  this.ensureController();
       
       
        var firstEnabled =c.canDecrement();
        var prevEnabled = c.canDecrement();
        var nextEnabled =c.canIncrement();

        return <div>
            <button style={{ visibility: firstEnabled ? "visible" : "hidden" }} onClick={()=>c.date(c.min())} disabled={!firstEnabled}><span role="img" aria-label="1">⏮</span></button>
            <button style={{ visibility: prevEnabled ? "visible" : "hidden" }} disabled={!prevEnabled} onClick={()=>c.advance(-1)}><span role="img" aria-label="1"> ⏪</span></button>
            <button onClick={()=>c.date(Time.now())}>{c.date().date.toLocaleDateString()}</button>
            <button style={{ visibility: nextEnabled ? "visible" : "hidden" }} disabled={!nextEnabled} onClick={()=>c.advance(1)}><span role="img" aria-label="1">⏩</span></button>
        </div>

    }

}


// datePickerController manages date data for the DatePicker.
// It takes min/max dates and converts them to midnight for
// each represented day with the range being min <= cursor <= max,
// e.g. inclusive.
//
export class datePickerController {
    constructor(minDate, maxDate, currentDate, tzOffsetMinutes) {

        this.tzOffsetMinutes = tzOffsetMinutes;

        // as utc.
        if (minDate) {
            this._min = this.toDate(minDate);
        }

        if (maxDate) {
            this._max = this.toDate(maxDate);
        }


        var cur = currentDate || Time.now();
        this._cursor = this.toDate(cur.box(this._min, this._max));
    }

    toDate(d) {

        var t = Time.wrap(d);

        var date = new Time(t.date.toDateString());
        if (this.tzOffsetMinutes) {
            date = date.add(this.tzOffsetMinutes, minute);
        }
        return Time.wrap(date);
    }


    min() {
        return this._min;
    }

    max() {
        return this._max;
    }

    date(d) {

        
        if (d === undefined) {
            return this._cursor;
        }

        d = this.toDate(d);
        if (Time.same(d, this._cursor)) {
            return this._cursor;
        }

       
        d = d.box(this._min, this._max);

        this._cursor = d;
        if (this.change) {
          this.change(d);
        }
        return d;
    }

    advance(nDays) {
        if (nDays === 0) {
            return this.date(this._max);
        }

        var newDate = this._cursor.add(nDays, day);
        newDate = newDate.box(this._min, this._max);
        return this.date(newDate);
    }

    canDecrement() {
        if (!this._min) {
            return true;
        }
        return this._cursor.after(this._min);
    }

    canIncrement() {

        if (!this._max) {
            return true;
        }
        return this._cursor.before(this._max);
    }

}



export default DatePicker;