import React from 'react';

const daysFirst = -1000000;

class DatePicker extends React.Component {
constructor(props) {
    super(props);

    var date = this.props.date || new Date();

    if (this.props.maxDate && this.props.maxDate < date) {
        date = this.props.maxDate;
    }
    this.state = {
    	maxDate: this.props.maxDate,
    	date: date ,
        minDate: this.props.minDate
     }
     console.log("Init datepicker: " + JSON.stringify(this.state))
 }
 
 sameDay(d1, d2)  {
 	return d1 && d2 && (d1.toDateString() == d2.toDateString());
 }

 componentDidUpdate() {
     
 }


 getSnapshotBeforeUpdate(prevProps, prevState) {
     if (this.state.date !== prevState.date && this.props.onChange) {
         this.props.onChange(this.state.date);
     }
     return null;
 }
 
 setDate(n, el) {
 		
 	console.log("Set date " + n);
    var d= this.state.date.getTime();
  	if (!n) {
    	 d = new Date().getTime();
    } else if (n === daysFirst) {
    	d = this.props.minDate;
    } else {
    	d += (24*60*60*1000*n);
    }

   this.setState( {
        date: new Date(d),
      }
    )
    
 }

 dayStart(d) {
    if (!d) {
        d = new Date()
    }
     return new Date(d.toDateString())
 }

	render() {
      const d = this.state.date;
      
      
     
      var prevDisabled = this.state.minDate && this.dayStart(this.state.minDate) >= this.dayStart(this.state.date);
      var firstDisabled = !this.state.minDate || prevDisabled;
      var nextDisabled = d >= this.dayStart(this.state.maxDate); 
      var todayDisabled =  this.props.maxDate < this.dayStart(); 

      todayDisabled = prevDisabled = nextDisabled = false;
      
    
    
    var first = <button days={daysFirst} onClick={this.setDate.bind(this, daysFirst)}
    disabled={firstDisabled}><span>⏮</span></button>;
    
    
    var today = <button onClick={this.setDate.bind(this, 0)} disabled={todayDisabled}><span>📆 Today</span></button>
    
    return <div>
      {first}
      <button  disabled={prevDisabled} onClick={this.setDate.bind(this, -1)}><span>⏪</span></button>
      <button onClick={this.setDate.bind(this, 0)}>{d.toDateString()}</button>
      <button days="1" disabled={nextDisabled} onClick={this.setDate.bind(this, 1)}><span>⏩</span></button>
      {today}
    </div>
  
  }

}


export default DatePicker;