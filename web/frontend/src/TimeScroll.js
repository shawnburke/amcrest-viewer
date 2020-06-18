import React from 'react';




export default class TimeScroll extends React.Component{


    constructor(props) {
        super(props);

        this.myRef = React.createRef();
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
        var offset = (el.currentTarget.clientWidth * .5) - 10;

        var viewportRect = el.currentTarget.getBoundingClientRect();
      
        var setPos = viewportRect.left+offset;
        var element = document.elementFromPoint(setPos, viewportRect.top + (viewportRect.bottom - viewportRect.top)/2);
        setPos += 10;

        var elementViewPort = element.getBoundingClientRect();

        var elementDelta = setPos - elementViewPort.x;

        var elementRatio = elementDelta / element.clientWidth;

        this.onScrollChange(element, elementRatio);

    }

    onScrollChange(el, ratio) {

        if (!el.attributes.time) {
            return;
        }
        const sec = el.attributes.seconds.value;
        const time =  Number(el.attributes.time.value);

        var newTime = time + (1000*sec*ratio);

        if (this.props.onTimeChange) {
            this.props.onTimeChange(newTime, el.attributes.item_id && el.attributes.item_id.value);
        }
        console.log(`Scrolled to ${el.scrollLeft}`);
    }

    boxTime(t, min, max) {
        

        var tt = this.toUnix(t);
        var tmin = this.toUnix(min);
        var tmax = this.toUnix(max);

        tt = Math.max(tt, tmin);
        tt = Math.min(tt, tmax);

        return tt;
    }

    toUnix(t) {
        if (t.getTime) {
            return t.getTime()
        }
        return t;
    }

    shouldComponentUpdate(nextProps, nextState) {


        if (!this.anchor) {

            if (nextProps.startTime === this.props.startTime &&
                nextProps.endTime === this.props.endTime && 
                nextProps.position !== this.props.position ) {
                    var t = this.toUnix(nextProps.position);
                    this.scrollToTime(t);
                    return false;
                }
        }


        return true;
    }

    scrollToTime(t) {
        var left = 0;
        const node = this.myRef.current;
        var indicator = node.clientWidth / 2;

        var bestFit;
        var bestFitDuration;

        for (var i = 0 ; i < node.childNodes.length;i++) {
            var child = node.childNodes[i];

            var childTime = child.attributes.time && child.attributes.time.value;

            if (childTime) {
            
                var seconds = Number(child.attributes.seconds.value);
                var childEnd = Number(childTime) + (1000*seconds);

                // // went too far
                // if (childEnd > t) {
                //     break;
                // }

                if (t > childTime && t < childEnd) {
                    if (!bestFitDuration || bestFitDuration > seconds) {
                        bestFit = child;
                    }
                }
            }

            left += child.getClientRects().width;
        }

        if (bestFit) {
            this.select(bestFit);
        }
    }

    
     select(e) {
        var p = e.offsetParent; 
        var s = this.myRef.current;
        var center = s.clientWidth / 2;
        var newScroll = ((e.offsetLeft - p.offsetLeft) - center) + p.offsetLeft;
        console.log(`Scrolling ${s.scrollLeft} => ${newScroll}`);
        s.scrollLeft = newScroll;
      }

    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        const hour = 60*60*1000;
        const month = hour * 24 * 30;
       
        var start = this.boxTime(this.props.startTime, new Date().getTime() - month, new Date());
        var end = this.boxTime(this.props.endTime, new Date().getTime() - month, new Date());

        // buffer an hour on either end, snap
        // to hour boundaries
        start -= (start % hour) + hour;
        end += (hour - (end % hour)) + hour;

        const spanSeconds = (end- start) / 1000;

        var items = [];

        const hourWidth = window.innerWidth / 4;
        const chunkSeconds = 3600;


        for (var i = 0; i < spanSeconds; i += chunkSeconds) {
            var t = new Date(start + (1000*i));

            var iEnd = new Date(t.getTime() + chunkSeconds*1000);
            var label = t.getHours();

            var mediaItems = this.props.items.filter(mi => mi.start >= t && mi.start < iEnd);
            
            var w = hourWidth;

            if (mediaItems && mediaItems.length) {
                w = 20;
            }

            if (label > 12) {
                label = (label %12 )+ "p";
            } else {
                label = (label || 12) + "a";
            }
            var hourItem = <div key={"file" + i} time={this.toUnix(t)} seconds={chunkSeconds} style={{
                display: "inline-block",
                height: "100%",
                width: w + "px",
                borderLeft:"thin white solid",
                color: "white",
                background:"navy",
                padding: "2px"
            }}>{label}</div>;

            items.push(hourItem);

            

            mediaItems.forEach(mi => {
                var seconds = (mi.end.getTime() - mi.start.getTime())/1000;

                var color = mi.video ? "green" : "gold";

                
                var motionItem = <div onClick={ev => this.select(ev.target)}
                    key={mi.id} item_id={mi.id} time={mi.start.getTime()} seconds={seconds} style={{
                    display: "inline-block",
                    position:"relative",
                    height: "60%",
                    width: "50px",
                    top:"15px",
                    borderLeft:"thin white solid",
                    color: "white",
                    background:color, 
                    padding: "2px",
                    MozBorderRadius:"5px",
                    WebkitBorderRadius:"5px",
                    border:"1px white solid",
                }}>{seconds}s</div>;

                items.push(motionItem);
                
            });
            
        }

       
       return <div ref={this.myRef}
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