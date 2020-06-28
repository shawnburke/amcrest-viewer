import React from 'react';
import { boxTime, toUnix, hour } from './time';


var tsCounter = 0;
const dividerWidth = 5;

export default class TimeScroll extends React.Component {


    constructor(props) {
        super(props);

        this.myRef = React.createRef();
        this.idBase = tsCounter++;
        this.renderCount = 0;
        this.state = {
            current: new Date(),
        }
    }

    log(s) {
        //console.log(s);
    }

    mouseDown(el) {
        this.mouseAnchor = [el.screenX, el.screenY];
    }

    mouseMove(el) {
        if (this.mouseAnchor) {
            var deltaX = el.screenX - this.mouseAnchor[0];

            var newScroll = (el.currentTarget.scrollLeft - deltaX);

            newScroll = Math.max(newScroll, 0);
            newScroll = Math.min(newScroll, el.currentTarget.scrollLeftMax)

            el.currentTarget.scrollLeft = newScroll;

            this.mouseAnchor = [el.screenX, el.screenY];


        }
    }

    mouseUp() {
        this.mouseAnchor = null;
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
        const sec = el.attributes.seconds.value;
        const time = this.getElementTime(el);

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

    componentDidUpdate(prevProps, prevState, snapshot) {

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

    // given an element and a time, return
    // the file item at that time.
    getElementItem(el, unixTime) {

        var item_ids = el.attributes.item_ids.value;

        var ids = item_ids.split(",");

        var items = this.props.items;

        while (ids.length > 0) {
            var candidateIndex = items.findIndex(mi => mi.id === Number(ids[0]));

            ids = ids.slice(1);

            if (candidateIndex === -1) {
                continue;
            }
            var c = items[candidateIndex];

            if (this.fileContains(c, unixTime)) {
                return c;
            }
        }
        return null;
    }

    fileContains(f, time) {
        var unixStart = toUnix(f.start);
        var unixEnd = toUnix(f.end);
        return time >= unixStart && time < unixEnd;
    }

    getElementTime(el) {
        var childTime = el.attributes.time && el.attributes.time.value;

        if (!childTime) {
            return null;
        }

        return this.toUnix(childTime);
    }

    getElementDuration(el) {
        var s = el.attributes.seconds && el.attributes.seconds.value;

        if (!s) {
            return 0;
        }

        return Number(s) * 1000;


    }




    scrollToTime(t) {

        // snap time to hour
        //
        var hourTime = t - (t % hour);
        var nextHour = hourTime + hour;

        var hours = document.getElementsByClassName("hour");

        // walk the hours looking for the one that matches this hour.

        for (var i = 0; i < hours.length; i++) {

            var hourEl = hours[i];
            var et = this.getElementTime(hourEl);

            if (et === hourTime) {

                // now walk siblings.
                //
                while (hourEl) {
                    et = this.getElementTime(hourEl);

                    if (et > nextHour) {
                        return;
                    }

                    var d = this.getElementDuration(hourEl);
                    var elEnd = et + d;

                    if (t < elEnd) {
                        var ratio = ((t - et)) / d;
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
            if (el) {
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

        if (cur && el === cur.element) {
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

    onMotionItemMouseUp(ev) {
        ev.preventDefault();
        var target = ev.target;
        var delta = this.myRef.current.scrollLeft - target.scrollAnchor;
        delete target.scrollAnchor;
        if (Math.abs(delta) > 10) {
            return;
        }


        this.scrollToElement(target);
    }

    onMotionItemMouseDown(ev) {
        ev.preventDefault();
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

    renderMediaItem(mi) {
        var seconds = (mi.end.getTime() - mi.start.getTime()) / 1000;

        var color = mi.video ? "navy" : "gold";

        if (!mi.video) {
            return;
        }

        // 5 minute video is full width
        var perSecond = window.innerWidth / (60 * 5);

        var w = Math.max(50, perSecond * seconds);

        var motionItem = <div id={this.getItemId(mi.id)}
            onMouseDown={this.onMotionItemMouseDown.bind(this)}
            onMouseUp={this.onMotionItemMouseUp.bind(this)}
            title={mi.id}
            key={`mi-${this.renderCount}-${mi.id}`} item_id={mi.id} item_ids={mi.id} time={mi.start.getTime()} seconds={seconds} style={{
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
            }}>{seconds}s</div>;
        return motionItem;
    }

    renderTimeItem(unixStart, ms, fileItems) {

        var topOfHour = unixStart % 3600 === 0;
        var halfHour = unixStart % 1800 === 0;
        var quarterHour = unixStart % 900 === 0;

        var seconds = ms / 1000;
        const hourWidth = window.innerWidth / 4;
        var w = seconds / 3600 * hourWidth;

        var label = <span>&nbsp;</span>;

        var borderLeft = "";

        var cls = "";


        if (topOfHour) {
            label = new Date(unixStart).getHours();
            if (label > 12) {
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


        if (this.seenIds[id]) {
            console.error("dupe");
        }
        this.seenIds[id] = true;


        var hourItem = <div key={id} className={cls} id={id} item_ids={itemIds.join(',')} time={unixStart} seconds={seconds} style={{
            display: "inline-block",
            height: "100%",
            width: w + "px",
            borderLeft: borderLeft,
            color: "white",
            background: "black",
            padding: "2px",
            textAlign: "left"
        }}>{label}</div>;


        return hourItem;
    }

    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        this.seenIds = {};

        this.renderCount++;


        var items = [];
        var mediaItems = this.props.items || [];


        // walk the items, creating sections as we go.
        //
        var startTime = toUnix(this.props.startTime);
        var endTime = toUnix(this.props.endTime);
        var curTime = startTime;



        while (curTime < endTime) {


            var untilNextHour = hour - (curTime % hour);
            var nextHour = curTime + untilNextHour;

            // get all of the items in the current hour
            // and get the next video

            var hourItems = mediaItems.filter(mi => toUnix(mi.start) >= curTime && toUnix(mi.start) < nextHour);

            var nextVideoIndex = hourItems.findIndex(mi => mi.video);


            var timeItemSpan = untilNextHour;

            if (nextVideoIndex !== -1) {
                timeItemSpan = toUnix(hourItems[nextVideoIndex].start) - curTime;
            }

            var timeItem = this.renderTimeItem(curTime, timeItemSpan, hourItems.slice(0, nextVideoIndex));
            items.push(timeItem);

            // remove the items
            mediaItems = mediaItems.slice(nextVideoIndex === -1 ? hourItems.length : nextVideoIndex);
            curTime += timeItemSpan;

            // render the media item
            if (nextVideoIndex !== -1) {
                var videoItem = hourItems[nextVideoIndex];
                items.push(this.renderMediaItem(videoItem));
                curTime += (videoItem.file.duration_seconds || 5) * 1000;
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
                overflowX: "hidden",
                overflowY: "hidden",
                whiteSpace: "nowrap",
                msOverflowStyle: "none",
                userSelect: "none",

            }}>
            <div id={this.getItemId("divider")} style={{
                position: "absolute",
                width: dividerWidth + "px",
                background: "yellow",
                height: "75px",
                left: "50%"
            }}></div>

            {items}

        </div>;
    }

}