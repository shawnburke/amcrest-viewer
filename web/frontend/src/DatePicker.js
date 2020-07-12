import React from 'react';

import { Time, day } from './time';

const daysFirst = -1000000;

class DatePicker extends React.Component {
    constructor(props) {
        super(props);


        this.state = {
            date: Time.now(),
        }
    }

  

    getSnapshotBeforeUpdate(_prevProps, prevState) {
        if (this.state.date !== prevState.date && this.props.onChange) {
            this.props.onChange(this.state.date);
        }
        return null;
    }


    setDate(n, noset) {


        var d = this.state.date || Time.now().floor(day);

        // if no date is past, snap to end or today
        //
        if (!n) {
            d = Time.now();

            if (this.props.maxDate && this.props.maxDate.before(d)) {
                d = this.props.maxDate;
            }
        }

        if (n === daysFirst) {
            d = this.props.minDate;
        } else {
            d = d.add(n, day);
        }

        const final = d.box(this.props.minDate, this.props.maxDate).floor(day);

      
        if (noset !== true) {

            this.setState(
                {
                 date: final,
                }
            )
        }
        return final;

    }

    dayStart(d) {
        if (!d) {
            d = new Time.now();
            if (d.after(this.props.maxDate)) {
                d = this.props.maxDate;
            }
        }
        return d.date.toDateString();
    }

    render() {
        var d = this.state.date;
        if (!d) {
            return <div></div>;
        }
        var minDate = this.props.minDate;
        var maxDate = this.props.maxDate;

        d = d.box(minDate, maxDate);
       
        var firstEnabled = minDate && this.dayStart(d) > this.dayStart(minDate);
        var prevEnabled = !minDate || firstEnabled;
        var nextEnabled = this.dayStart(d) < this.dayStart() && (!maxDate || this.dayStart(d) < this.dayStart(maxDate))


        return <div>
            <button style={{ visibility: firstEnabled ? "visible" : "hidden" }} onClick={this.setDate.bind(this, daysFirst)} disabled={!firstEnabled}><span role="img" aria-label="1">⏮</span></button>
            <button style={{ visibility: prevEnabled ? "visible" : "hidden" }} disabled={!prevEnabled} onClick={this.setDate.bind(this, -1)}><span role="img" aria-label="1"> ⏪</span></button>
            <button onClick={this.setDate.bind(this, 0)}>{d.date.toLocaleDateString()}</button>
            <button style={{ visibility: nextEnabled ? "visible" : "hidden" }} disabled={!nextEnabled} onClick={this.setDate.bind(this, 1)}><span role="img" aria-label="1">⏩</span></button>
        </div>

    }

}


export default DatePicker;