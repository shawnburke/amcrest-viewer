import React from 'react';
import { boxTime, toUnix } from './time';


var tsCounter = 0;

export default class TimeScroll extends React.Component {


    constructor(props) {
        super(props);

        this.myRef = React.createRef();
        this.idBase = tsCounter++;
        this.state = {
            current: new Date(),
        }
    }

    mouseDown(el) {
        this.anchor = [el.screenX, el.screenY];
    }

    mouseMove(el) {
        if (this.anchor) {
            var deltaX = el.screenX - this.anchor[0];

            var newScroll = (el.currentTarget.scrollLeft - deltaX);

            newScroll = Math.max(newScroll, 0);
            newScroll = Math.min(newScroll, el.currentTarget.scrollLeftMax)

            el.currentTarget.scrollLeft = newScroll;

            this.anchor = [el.screenX, el.screenY];


        }
    }

    mouseUp() {
        this.anchor = null;
    }


    onScroll() {

        var efs = this.elementFromScroll(10);

        if (!efs) {
            console.warn(`Can't find element from scroll position!`);
            return;
        }

        this.onScrollChange(efs.element, efs.ratio);

    }

    onScrollChange(el, ratio) {

        if (!el.attributes.time) {
            return;
        }
        const sec = el.attributes.seconds.value;
        const time = this.getElementTime(el);

        var newTime = time + (1000 * sec * ratio);

        var item = this.getTimeMapItem(newTime);

        if (this.props.onTimeChange) {
            this.props.onTimeChange(newTime, item);
        }
        console.log(`Scrolled to ${this.myRef.current.scrollLeft}, Item=${item && item.item.id}`);
    }

    boxTime(t, min, max) {
        return boxTime(t, min, max);
    }

    toUnix(t) {
        return toUnix(t);
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        this.getTimeMap();
    }

    shouldComponentUpdate(nextProps, nextState) {


        var firstItemIdNext = nextProps.items && nextProps.items.length && nextProps.items[0].id;
        var firstItemIdProps = this.props.items && this.props.items.length && this.props.items[0].id;

        const positionOnlyChange = firstItemIdNext === firstItemIdProps &&
            nextProps.startTime === this.props.startTime &&
            nextProps.endTime === this.props.endTime &&
            nextProps.position !== this.props.position;

        if (positionOnlyChange) {
            var t = this.toUnix(nextProps.position);
            this.scrollToTime(t);
            return false;
        }
        delete this.mappedItems;
        return true;
    }

    getElementTime(el) {
        var childTime = el.attributes.time && el.attributes.time.value;

        if (!childTime) {
            return null;
        }

        return this.toUnix(childTime);
    }


    getScrollTime() {
        const span = this.totalSeconds();

        const parent = this.myRef.current;
        var center = parent.clientWidth / 2;

        var scrollPerSecond = (parent.scrollWidth - parent.clientWidth) / span.seconds;

        var scrollTime = parent.scrollLeft / scrollPerSecond;

        return Math.trunc(span.start + (scrollTime * 1000));
    }

    getTimeMapItem(t) {
        const span = this.totalSeconds();


        const seconds = (t - span.start) / 1000;

        const mapIndex = Math.trunc(seconds / this.mappedItems.chunkSeconds);

        var map = this.mappedItems.timeMap;
        var itemList = map[mapIndex];

        if (!itemList || !itemList.length) {
            return;
        }

        var item = itemList[0];
        return item;
    }

    scrollToTime(t) {

        const item = this.getTimeMapItem(t);

        if (!item) {
            return;
        }

        var childTime = item.unixStart;
        var childDuration = item.item.duration_seconds;

        var ratio = ((t - childTime) / 1000) / childDuration;
        this.scrollToElement(item.element, ratio);

    }

    elementFromScroll(buffer) {

     

        // see what's under the divider
        var div = document.getElementById(this.getItemId('divider'));
        if (div) {
            var divRect = div.getBoundingClientRect();
            var el = document.elementFromPoint(divRect.x-divRect.width, divRect.y+(divRect.height/2));
            if (el) {
                var elRect = el.getBoundingClientRect();
                var ratio = (divRect.x - elRect.x) / elRect.width;
                return {
                    element: el,
                    ratio: ratio,
                }
            }
        }
        return null;
       
    }


    getElementX(el) {
        var parent = this.myRef.current;
        var center = parent.clientWidth / 2;
        var left = el.offsetLeft - parent.offsetLeft;
        return left - center;
    }

    scrollToElement(el, ratio) {

        var cur = this.elementFromScroll(10);

        if (cur && el == cur.element) {
            return;
        }

        var left = this.getElementX(el);

        var extra = 0;

        if (ratio) {
            extra = el.clientWidth * ratio;
        }

        var parent = this.myRef.current;
        var newScroll = left + extra;
        console.log(`Scrolling ${parent.scrollLeft} => ${newScroll}`);
        parent.scrollLeft = newScroll;
    }

    getTimeMap() {
        if (this.mappedItems) {
            return
        }

        const ts = this.totalSeconds();

        const chunkSeconds = 5;

        var chunks = new Array(ts.seconds / chunkSeconds);
        var count = 0;

        var itemsByScroll = new Array(this.props.items.length);

        this.props.items.forEach((mi, i) => {


            var id = this.getItemId(mi.id);
            var el = document.getElementById(id);
            if (!el) {
                console.error(`Couldn't find element for id ${id} (mi-id: ${mi.id})`);
                return;
            }

            var info = {
                element: el,
                width: el.clientWidth,
                item: mi,
                scrollLeft: this.getElementX(el),
                unixStart: this.toUnix(mi.start),
                unixEnd: this.toUnix(mi.end),
            };

            for (var ct = info.unixStart; ct < info.unixEnd;ct+=chunkSeconds*1000) {

                var chunkIndex = Math.trunc((ct- ts.start) / (1000 * chunkSeconds));
                var list = chunks[chunkIndex] || [];
                list.push(info);
                chunks[chunkIndex] = list;
            }
            itemsByScroll[i] = info;

            count++;
        });
        this.mappedItems = {
            timeMap: chunks,
            scrollMap: itemsByScroll,
            chunkSeconds: chunkSeconds,
        }

        console.log(`Build new time map with ${chunks.length} slots, ${count} items`)
    }

    onMotionItemMouseUp(ev) {

        var target = ev.target;
        var delta = this.myRef.current.scrollLeft - target.scrollAnchor;
        delete target.scrollAnchor;
        if (Math.abs(delta) > 10) {
            return;
        }

        this.scrollToElement(target);
    }

    onMotionItemMouseDown(ev) {

        // snap the current scroll offset
        ev.target.scrollAnchor = this.myRef.current.scrollLeft;
    }

    totalSeconds() {
        const hour = 60 * 60 * 1000;
        const month = hour * 24 * 30;

        var start = this.boxTime(this.props.startTime, new Date().getTime() - month, new Date());
        var end = this.boxTime(this.props.endTime, new Date().getTime() - month, new Date());

        // buffer an hour on either end, snap
        // to hour boundaries
        start -= (start % hour) + hour;
        end += (hour - (end % hour)) + hour;

        const spanSeconds = (end - start) / 1000;
        return {
            seconds: spanSeconds,
            start: start,
            end: end,
        }
    }

    getItemId(key) {
        return `ts-${this.idBase}-${key}`;
    }

    renderItem(mi){
        var seconds = (mi.end.getTime() - mi.start.getTime()) / 1000;

        var color = mi.video ? "green" : "gold";

        if (!mi.video) {
            return;
        }


        var motionItem = <div id={this.getItemId(mi.id)}
            onMouseDown={this.onMotionItemMouseDown.bind(this)}
            onMouseUp={this.onMotionItemMouseUp.bind(this)}
            key={mi.id} item_id={mi.id} time={mi.start.getTime()} seconds={seconds} style={{
                display: "inline-block",
                position: "relative",
                height: "60%",
                width: "50px",
                top: "15px",
                borderLeft: "thin white solid",
                color: "white",
                background: color,
                padding: "2px",
                MozBorderRadius: "5px",
                WebkitBorderRadius: "5px",
                border: "1px white solid",
            }}>{seconds}s</div>;
        return motionItem;
    }

    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        const span = this.totalSeconds();

        var items = [];
        var mediaItems = this.props.items || [];

        const hourWidth = window.innerWidth / 4;
        const chunkSeconds = 600;

        var itemPos = 0;


        for (var i = 0; i < span.seconds; i += chunkSeconds) {

            // create background items for every n minutes.
            // if the item is an hour boundary, add the label.
            //

            var tStart = new Date(span.start + (1000 * i));
            var tEnd = new Date(toUnix(tStart) + (chunkSeconds*1000));
            var curHour = tStart.getHours()
           


            var itemDistanceSeconds = chunkSeconds;
            
            if (mediaItems[itemPos] && mediaItems[itemPos].start.getHours() === curHour) {
                itemDistanceSeconds = (toUnix(mediaItems[itemPos].start) - toUnix(tStart)) / 1000;
            }

            var topOfHour = i % 3600 === 0;

            var secondsWidth = itemDistanceSeconds;

            var w = secondsWidth / 3600 * hourWidth;

        

            var label = <span>&nbsp;</span>;
            if (topOfHour) {
                label = curHour;
                if (label > 12) {
                    label = (label % 12) + "p";
                } else {
                    label = (label || 12) + "a";
                }

                w = Math.max(w, 25);
            }
            
           
           
            var hourItem = <div key={toUnix(tStart)} time={this.toUnix(tStart)} seconds={secondsWidth} style={{
                display: "inline-block",
                height: "100%",
                width: w + "px",
                borderLeft: topOfHour ? "thin white solid": "",
                color: "white",
                background: "navy",
                padding: "2px"
            }}>{label}</div>;

            items.push(hourItem);

            if (!mediaItems[itemPos]) {
                continue;
            }

            var lastItemEnd = mediaItems[itemPos].unixStart;

            while (true) {

                var nextItem = mediaItems[itemPos];
                
                if (!nextItem) {
                    break;
                }

                // if more than 5 mins between items
                if (nextItem.unixStart - lastItemEnd > chunkSeconds*1000) {
                    break;
                }

                // if in next hour
                if (nextItem.start.getHours() != curHour) {
                    break;
                }

                // render this item
                var timeItem = this.renderItem(nextItem);
                if (timeItem) {
                    items.push(timeItem);
                    lastItemEnd = timeItem.unixEnd;

                    // if we have advanced more than half way into a chunk, update
                    // index
                    var dist = lastItemEnd - tStart.getTime();

                    if (dist > (chunkSeconds * 1000 / 2)) {
                        i++;
                        tStart = new Date(tStart.getTime() + (chunkSeconds*1000))
                    }
                }
                itemPos++;
            }

        }


        return <div ref={this.myRef}
            onMouseDown={this.mouseDown.bind(this)}
            onMouseUp={this.mouseUp.bind(this)}
            onMouseMove={this.mouseMove.bind(this)}
            onScroll={this.onScroll.bind(this)}
            style={{
                width: "100%",
                height: "75px",
                background: "darkblue",
                overflowX: "auto",
                overflowY: "hidden",
                whiteSpace: "nowrap",
                msOverflowStyle: "none"

            }}>
            <div id={this.getItemId("divider")} style={{
                position: "absolute",
                width: "5px",
                background: "yellow",
                height: "75px",
                left: "50%"
            }}></div>

            {items}

        </div>;
    }

}