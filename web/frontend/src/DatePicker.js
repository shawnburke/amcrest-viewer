import React from 'react';

import { snapTime } from './time';

const daysFirst = -1000000;

class DatePicker extends React.Component {
    constructor(props) {
        super(props);


        this.state = {
            date: null,
        }
        this.state.date = this.setDate(0, true);
    }

    sameDay(d1, d2) {
        return d1 && d2 && (d1.toDateString() === d2.toDateString());
    }



    getSnapshotBeforeUpdate(_prevProps, prevState) {
        if (this.state.date !== prevState.date && this.props.onChange) {
            this.props.onChange(this.state.date);
        }
        return null;
    }

    boxDate(d) {
        if (this.props.minDate && this.props.minDate > d) {
            d = this.props.minDate;
        }

        if (this.props.maxDate && this.props.maxDate < d) {
            d = this.props.maxDate;
        }

        return d;
    }

    setDate(n, noset) {

        var d = this.state.date ? this.state.date.getTime() : new Date();
        if (!n) {
            d = new Date();

            if (this.props.maxDate && this.props.maxDate < d) {
                d = this.props.maxDate;
            }
            d = d.getTime();

        } else if (n === daysFirst) {
            d = this.props.minDate;
        } else {
            d += (24 * 60 * 60 * 1000 * n);
        }

        // box the values.
        var final = snapTime(this.boxDate(new Date(d)), "day", "-1");

        if (noset !== true) {

            this.setState({
                date: final,
            }
            )
        }
        return final;

    }

    dayStart(d) {
        if (!d) {
            d = new Date();
            if (d > this.props.maxDate) {
                d = this.props.maxDate;
            }
        }
        return new Date(d.toDateString())
    }

    render() {
        var d = this.state.date;
        if (!d) {
            return <div></div>;
        }
        d = this.boxDate(d);
        var minDate = this.props.minDate;
        var maxDate = this.props.maxDate;

        var firstEnabled = minDate && this.dayStart(d) > this.dayStart(minDate);
        var prevEnabled = !minDate || firstEnabled;
        var nextEnabled = this.dayStart(d) < this.dayStart() && (!maxDate || this.dayStart(d) < this.dayStart(maxDate))


        return <div>
            <button style={{ visibility: firstEnabled ? "visible" : "hidden" }} onClick={this.setDate.bind(this, daysFirst)} disabled={!firstEnabled}><span role="img" aria-label="1">⏮</span></button>
            <button style={{ visibility: prevEnabled ? "visible" : "hidden" }} disabled={!prevEnabled} onClick={this.setDate.bind(this, -1)}><span role="img" aria-label="1"> ⏪</span></button>
            <button onClick={this.setDate.bind(this, 0)}>{d.toLocaleDateString()}</button>
            <button style={{ visibility: nextEnabled ? "visible" : "hidden" }} disabled={!nextEnabled} onClick={this.setDate.bind(this, 1)}><span role="img" aria-label="1">⏩</span></button>
        </div>

    }

}


export default DatePicker;