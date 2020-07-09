import React from 'react';
import { boxTime, toUnix, hour, iso } from './time';


var tsCounter = 0;
const dividerWidth = 5;

export default class TimeScroll extends React.Component {


    constructor(props) {
        super(props);

        this.myRef = React.createRef();
        this.idBase = tsCounter++;
        this.state = {
            current: new Date(),
        }
    }

    log(_s) {

        console.log(_s);
    }

    mouseDown(ev) {
        ev.preventDefault();
        this.mouseAnchor = [ev.screenX, ev.screenY];
    }

    mouseMove(ev) {
        ev.preventDefault();
        if (this.mouseAnchor) {
            var deltaX = ev.screenX - this.mouseAnchor[0];

            var newScroll = (ev.currentTarget.scrollLeft - deltaX);

            newScroll = Math.max(newScroll, 0);
            newScroll = Math.min(newScroll, ev.currentTarget.scrollLeftMax || ev.currentTarget.scrollWidth)

            ev.currentTarget.scrollLeft = newScroll;

            this.mouseAnchor = [ev.screenX, ev.screenY];


        }
    }

    mouseUp(ev) {
        ev.preventDefault();
        this.mouseAnchor = null;
    }

    onMotionItemMouseUp(ev) {
        var target = ev.target;
        var delta = this.myRef.current.scrollLeft - target.scrollAnchor;
        delete target.scrollAnchor;
        if (Math.abs(delta) > 10) {
            return;
        }
        ev.preventDefault();
        this.scrollToElement(target);
    }

    onMotionItemMouseDown(ev) {

        // snap the current scroll offset
        ev.target.scrollAnchor = this.myRef.current.scrollLeft;
    }



    onScroll() {
        // when we scroll, we are looking for the element
        // under the selector so we can notify on it.
        var efs = this.elementFromScroll();

        if (!efs) {
            var selPos = this.getSelectionPoint();
            console.warn(`No element at scroll ${this.myRef.current.scrollLeft}, x=${selPos.x}, y=${selPos.y}`);
            return;
        }

        // move on to firing the scroll event
        this.onScrollChange(efs.element, efs.ratio);
    }



    onScrollChange(el, ratio) {

        if (!el.attributes.time) {
            return;
        }

        const range = this.getElementRange(el);
        const sec = range.seconds;
        const time = range.start;

        var newTime = time + (1000 * sec * ratio);

        var item = this.getElementItem(el, newTime);


        if (this.props.onTimeChange) {
            this.props.onTimeChange(newTime, item);
        }
        this.log(`Scrolled to ${this.myRef.current.scrollLeft}, Item=${item && item.id}`);
    }

    boxTime(t, min, max) {
        return boxTime(t, min, max);
    }

    toUnix(t) {
        return toUnix(t);
    }



    shouldComponentUpdate(nextProps, _nextState) {


        var firstItemIdNext = nextProps.items && nextProps.items.length && nextProps.items[0].id;
        var firstItemIdProps = this.props.items && this.props.items.length && this.props.items[0].id;

        const positionOnlyChange = firstItemIdNext === firstItemIdProps &&
            nextProps.startTime === this.props.startTime &&
            nextProps.endTime === this.props.endTime &&
            nextProps.position !== this.props.position;

        if (positionOnlyChange && nextProps.position) {

            if (this.mouseAnchor) {
                return false;
            }

            var t = this.toUnix(nextProps.position);
            this.scrollToTime(t);
            return false;
        }
        return true;
    }

    componentDidUpdate() {
        if (this.props.position) {
            this.scrollToTime(this.toUnix(this.props.position));
        }
    }

    // given an element and a time, return
    // the file item at that time.
    getElementItem(el, unixTime) {

        var item_ids = el.attributes.item_ids.value;

        if (!item_ids) {
            return null;
        }

        var ids = item_ids.split(",").map(id => Number(id));

        ids.sort((a, b) => a - b);

        var items = this.props.items;

        if (!el.item_map) {
            el.item_map = {}

            var now = new Date();
            ids.forEach(id => {
                id = Number(id);

                if (el.item_map[id]) {
                    return;
                }

                // note this might miss items if the indexes aren't in 
                // time order, but it's too slow otherwise.
                var index = items.findIndex(mi => mi.id === id);
                if (index === -1) {
                    // try from beginning
                    index = this.props.items.findIndex(mi => mi.id === id);
                    if (index === -1) {
                        el.item_map[index] = "x";
                        console.error(`Couldn't find index for id ${id}`);
                        return;
                    }
                    items = this.props.items;
                }
                el.item_map[id] = items[index];
                items = items.slice(index);
            });
            console.log(`Map create time for ${item_ids.length} elements out of ${this.props.items.length}: ${new Date().getTime() - now.getTime()}ms`)
        }


        while (ids.length > 0) {

            const target = Number(ids[0]);
            var item = el.item_map[target];

            ids = ids.slice(1);

            if (!item || item === "x") {
                console.error(`Couldn't find ${target}`);
                continue;
            }
            if (this.itemContains(item, unixTime)) {
                this.log(`Found item ${item.id} at ${iso(unixTime)}`)
                return item;
            }
        }

        var elRange = this.getElementRange(el);
        this.log(`No items in ${el.id} between ${iso(elRange.start)} => ${iso(elRange.end)} for ${iso(unixTime)}`)
        return null;
    }

    itemContains(item, time) {
        if (!item.start) {
            throw new Error("bad item:" + JSON.stringify(item))
        }
        var unixStart = toUnix(item.start);
        var unixEnd = toUnix(item.end);
        time = toUnix(time);
        return time >= unixStart && time < unixEnd;
    }


    getElementRange(el) {

        var start = el && el.attributes.time && el.attributes.time.value;

        if (!start) {
            return { invalid: true }
        }

        var seconds = el.attributes.seconds && el.attributes.seconds.value;
        seconds = seconds || 0;

        var ms = seconds * 1000;

        start = toUnix(start)
        var range = {
            start: start,
            seconds: seconds,
            ms: ms,
            end: start + ms,
        };
        range.contains = function (ts) {
            return this.itemContains(range, ts);
        }
        return range;
    }

    scrollToTime(t) {

        // snap time to hour
        //
        var hourTime = t - (t % hour);
        var nextHour = hourTime + hour;

        // make sure they are in order
        var hours = document.getElementsByClassName("hour");
        hours = Array.from(hours).sort((h1, h2) => h1.id.localeCompare(h2.id));

        // walk the hours looking for the one that matches this hour.

        for (var i = 0; i < hours.length; i++) {

            var hourEl = hours[i];
            var elRange = this.getElementRange(hourEl);

            // look for an element that is in the same hour
            // as the target.  once we find that,
            // start walking siblings
            if (elRange.start >= hourTime && elRange.start < nextHour) {

                // now walk siblings.
                //
                while (hourEl) {
                    elRange = this.getElementRange(hourEl);

                    if (elRange.start > nextHour) {
                        return;
                    }

                    if (t < elRange.end) {
                        var ratio = ((t - elRange.start)) / elRange.ms;
                        this.scrollToElement(hourEl, ratio);
                        return;
                    }

                    hourEl = hourEl.nextElementSibling;
                }
                break;
            }
        }
    }

    getSelectionPoint() {
        var div = document.getElementById(this.getItemId('divider'));
        if (div) {
            var divRect = div.getBoundingClientRect();
            return {
                x: divRect.x - 1,
                y: divRect.y + (divRect.height / 2),
            }
        }
        return null;
    }

    elementFromScroll() {

        var selPoint = this.getSelectionPoint();

        if (selPoint) {
            var el = document.elementFromPoint(selPoint.x, selPoint.y);
            while (el !== this.myRef.current) {

                var r = this.getElementRange(el);

                if (r.invalid) {
                    el = el.parentElement;
                    continue;
                }

                var elRect = el.getBoundingClientRect();
                var ratio = (selPoint.x - elRect.x) / elRect.width;
                return {
                    element: el,
                    ratio: ratio,
                }
            }
        }
        return null;
    }


    // the the x position for an element, relative to the parent
    getElementX(el) {
        var parent = this.myRef.current;
        var center = parent.clientWidth / 2;
        var left = el.offsetLeft - parent.offsetLeft;
        return left - center;
    }

    scrollToElement(el, ratio) {

        var cur = this.elementFromScroll();

        if ((cur && el === cur.element) || !cur) {
            return;
        }

        var left = this.getElementX(el) + dividerWidth;

        var extra = 0;

        if (ratio) {
            extra = el.clientWidth * ratio;
        }

        var parent = this.myRef.current;
        var newScroll = left + extra;
        this.log(`Scrolling ${parent.scrollLeft} => ${newScroll}`);
        parent.scrollLeft = newScroll;
    }

    getItemId(key) {
        return `ts-${this.idBase}-${key}`;
    }

    renderMediaItem(mi) {
        var seconds = (toUnix(mi.end) - toUnix(mi.start)) / 1000;

        var color = mi.video ? "navy" : "gold";

        if (!mi.video) {
            return;
        }

        var startTime = new Date(mi.start).toLocaleTimeString();


        var w = this.myRef.current.clientWidth / 2;

        var motionItem = <div id={this.getItemId(mi.id)}
            onMouseDown={this.onMotionItemMouseDown.bind(this)}
            onMouseUp={this.onMotionItemMouseUp.bind(this)}
            title={mi.id}
            key={`mi-${mi.id}`} item_id={mi.id} item_ids={mi.id} time={mi.start.getTime()} seconds={seconds} style={{
                display: "inline-block",
                position: "relative",
                height: "60%",
                width: w + "px",
                top: "15px",
                borderLeft: "thin white solid",
                color: "white",
                background: color,
                padding: "2px",
                MozBorderRadius: "5px",
                WebkitBorderRadius: "5px",
                border: "1px white solid",
                fontSize: ".75em",
                textAlign: "left",
                paddingLeft: "5px",
                paddingTop: "10px",
            }}>
            <span role="img" aria-label="icon">🎥</span>
            <span>{startTime} ({seconds}s)</span>
        </div>;
        return motionItem;
    }

    hourWidth() {
        return window.innerWidth / 4;
    }

    renderTimeItem(unixStart, ms, fileItems, isBuffer) {

        var topOfHour = unixStart % 3600 === 0;
        var halfHour = unixStart % 1800 === 0;
        var quarterHour = unixStart % 900 === 0;

        var seconds = ms / 1000;
        var w = seconds / 3600 * this.hourWidth();

        var label = <span>&nbsp;</span>;

        var borderLeft = "";

        var cls = "";

        var background = "black";

        if (isBuffer) {
            background = "#333";
        }

        if (topOfHour) {
            label = new Date(unixStart).getHours();

            if (label === 12) {
                label += "p";
            } else if (label > 12) {
                label = (label % 12) + "p";
            } else {
                label = (label || 12) + "a";
            }

            w = Math.max(w, 25);
            borderLeft = "thin white solid";
            cls = "hour"
        } else if (halfHour) {
            cls = "half"
        } else if (quarterHour) {
            cls = "quarter"
            borderLeft = "thin silver dotted";
        }

        var id = this.getItemId(`ti-${unixStart}`);
        var itemIds = fileItems.map(i => `${i.id}`);

        var children = []

        fileItems.forEach(fi => {

            var leftPercent = ((toUnix(fi.start) - unixStart) / ms) * 100;
            var widthPercent = 100 * ((toUnix(fi.end) - toUnix(fi.start)) / ms);

            var childItem = <div style={{
                display: "inline-block",
                position: "absolute",
                background: "darkred",
                height: "2px",
                top: "85%",
                left: leftPercent + "%",
                width: Math.max(1, widthPercent) + "%",
            }}>&nbsp;</div>;

            children.push(childItem);

        });

        var timeItem = <div key={id} className={cls} id={id} item_ids={itemIds.join(',')} time={unixStart} seconds={seconds} style={{
            display: "inline-block",
            position: "relative",
            height: "100%",
            width: w + "px",
            borderLeft: borderLeft,
            color: "white",
            background: background,
            padding: "2px",
            textAlign: "left"
        }}>
            <span>{label}</span>
            {children}
        </div>;


        return timeItem;
    }

    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        const minDurationSeconds = 5;
        var items = [];
        var mediaItems = this.props.items || [];


        // walk the items, creating sections as we go.
        //
        var windowStart = toUnix(this.props.startTime);
        var windowEnd = toUnix(this.props.endTime);

        var now = new Date().getTime();

        windowEnd = Math.min(windowEnd, now);

        // buffer the edges to make sure everything
        // can be selected
        // TODO: scale this to be exactly half the 
        // distance.

        var buffer = hour * Math.trunc((window.innerWidth / 2) / this.hourWidth());
        var startTime = windowStart - buffer;
        var endTime = windowEnd + buffer;
        var curTime = startTime;

        while (curTime < endTime) {


            const untilNextHour = hour - (curTime % hour);
            const nextHour = curTime + untilNextHour;
            const cur = curTime;

            // get all of the items in the current hour
            // and get the next video

            var hourItems = mediaItems.filter(mi => toUnix(mi.start) >= cur && toUnix(mi.start) < nextHour);

            var nextVideoIndex = hourItems.findIndex(mi => mi.video);


            var timeItemSpan = untilNextHour;

            var hourItemsCount = hourItems.length;

            if (nextVideoIndex !== -1) {
                timeItemSpan = toUnix(hourItems[nextVideoIndex].start) - curTime;
                hourItemsCount = nextVideoIndex;
            }


            var isBuffer = curTime < windowStart || curTime >= windowEnd;
            var timeItem = this.renderTimeItem(
                curTime, timeItemSpan,
                hourItems.slice(0, hourItemsCount),
                isBuffer);
            items.push(timeItem);

            // remove the items
            mediaItems = mediaItems.slice(nextVideoIndex === -1 ? hourItems.length : nextVideoIndex);
            curTime += timeItemSpan;

            // render the media item
            if (nextVideoIndex !== -1) {
                var videoItem = hourItems[nextVideoIndex];
                items.push(this.renderMediaItem(videoItem));
                curTime += (videoItem.file.duration_seconds || minDurationSeconds) * 1000;
                mediaItems = mediaItems.slice(1);
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
                border: "thin white solid",
                padding: "1px",
                overflowX: "auto",
                overflowY: "hidden",
                whiteSpace: "nowrap",
                msOverflowStyle: "none",
                userSelect: "none",

            }}>
            <div id={this.getItemId("divider")} style={{
                position: "absolute",
                zIndex: "1000",
                width: dividerWidth + "px",
                background: "yellow",
                height: "75px",
                left: "50%"
            }}></div>

            {items}

        </div>;
    }

}