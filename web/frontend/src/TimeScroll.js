import React from 'react';




export default class TimeScroll extends React.Component{


    constructor(props) {
        super(props);

        this.state =  {
            current: new Date(),
        }
    }

    mouseDown(el) {
        this.anchor = [el.screenX, el.screenY];
    }

    mouseMove(el) {
        if (this.anchor) {
            var deltaX = el.screenX - this.anchor[0];

            var newScroll =  (el.currentTarget.scrollLeft - deltaX);

            newScroll = Math.max(newScroll, 0);
            newScroll = Math.min(newScroll, el.currentTarget.scrollLeftMax)

            el.currentTarget.scrollLeft = newScroll;

            this.anchor = [el.screenX, el.screenY];

            
        }
    }
    
    mouseUp(el) {
        this.anchor = null;
    }

    onScroll(el) {
        var scr = el.currentTarget.scrollLeft;
        var offset = (el.currentTarget.clientWidth * .5) - 10;

        var viewportRect = el.currentTarget.getBoundingClientRect();
        var pos = viewportRect.left + offset;

        var setPos = viewportRect.left+offset;
        var element = document.elementFromPoint(setPos, viewportRect.top + 5);
        setPos += 10;

        var elementViewPort = element.getBoundingClientRect();

        var elementDelta = setPos - elementViewPort.x;

        var elementRatio = elementDelta / element.clientWidth;

        this.onScrollChange(element, elementRatio);

    }

    onScrollChange(el, ratio) {

    }

    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        var start = this.props.startTime.getTime();
        var end = this.props.endTime.getTime();
        const hour = 60*60*1000;
        start -= (start % hour);
        end += (hour - (end % hour));

        const spanSeconds = (end- start) / 1000;

        var items = [];

        const hourWidth = window.innerWidth / 4;



        for (var i = 0; i < spanSeconds; i += 3600) {
            var t = new Date(start + (1000*i));
            var label = t.getHours();

            if (label > 12) {
                label = (label %12 )+ "p";
            } else {
                label = (label || 12) + "a";
            }
            var timeItem = <div key={i} style={{
                display: "inline-block",
                height: "100%",
                width: hourWidth + "px",
                borderLeft:"thin white solid",
                color: "white",
                background:"navy",
                padding: "2px"
            }}>{label}</div>;

            items.push(timeItem);

        }
       return <div 
        onMouseDown={this.mouseDown.bind(this)}
        onMouseUp={this.mouseUp.bind(this)}
        onMouseMove={this.mouseMove.bind(this)}
        onScroll={this.onScroll.bind(this)}
        style={{
           width:"100%",
           height:"50px", 
           background:"darkblue", 
           overflowX:"auto",
           overflowY: "hidden",
           whiteSpace: "nowrap",
          msOverflowStyle:"none"
         
        }}>
        <div style={{
            position:"absolute",
            width: "5px",
            background: "yellow",
            height:"50px",
            left: "50%"
        }}></div>

            {items}

        </div>;
    }

}