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
            var selPos = this.getSelectionPoint();
            console.warn(`No element at scroll ${this.myRef.current.scrollLeft}, x=${selPos.x}, y=${selPos.y}`);
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

        var scrollPerSecond = (parent.scrollWidth - parent.clientWidth) / span.seconds;
        var scrollTime = parent.scrollLeft / scrollPerSecond;
        return Math.trunc(span.start + (scrollTime * 1000));
    }

    getTimeMapItem(t) {

        // do binary search in the timeMap for the closest item
        //

        var item = this.findNearestItem(t, this.mappedItems.startTimeMap)

        if (item) {
            // direct hit!
            if (item.found) {
                return item.item;
            }
            console.log(`Found nearest item ${item.item.id} but time was not contained.`)
            return null;
        }


    }

    // findNearestItem does a binary search for the nearest item to the given time
    findNearestItem(t, items, offset) {

        var log = s => {
            // console.log(s);
        }

        if (!items || !items.length) {
            log("findNearest: Items array is empty")
            return null;
        }


        log(`findNearest: Search ${items.length} for ${new Date(t).toISOString()} from ${new Date(items[0].unixStart).toISOString()} -> ${new Date(items[items.length - 1].unixStart).toISOString()}`);


        offset = offset || 0;

        var createResult = (i, o) => {
            var r = {
                item: i,
                index: o,
            };

            if (i.unixStart <= t && i.unixEnd >= t) {
                r.ratio = (t - i.unixStart) / (i.unixEnd - i.unixStart);
                r.found = true;
            }

            log(`Found item ${i.item.id} (${i.item.duration_seconds}), Start=${new Date(i.unixStart).toISOString()}, ID=${i.item.id}, ratio=${r.ratio}, Delta seconds=${(i.unixStart - t) / 1000}`)
            return r;
        }

        if (items.length === 1) {
            log("findNearest: Items array 1");
            return createResult(items[0], offset);
        }

        var midPoint = Math.trunc(items.length / 2);
        var midPointItem = items[midPoint];

        if (midPointItem.unixStart <= t && midPointItem.unixEnd >= t) {
            return createResult(midPointItem, offset);
        }

        // we want to find the item nearest BEFORE the current position
        // so when we bisect, we always include the prior item
        if (t > midPointItem.unixStart) {
            log(`findNearest: Split up ${items.length} => ${midPoint}-`);
            return this.findNearestItem(t, items.slice(midPoint), offset + midPoint)
        }

        // otherwise, if the time is before the midpoint, we just to the lower half
        log(`findNearest: Split down ${items.length} => 0-${midPoint}`);
        return this.findNearestItem(t, items.slice(0, midPoint - 1), offset);
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
        var itemsByStartTime = [];

        this.props.items.forEach((mi) => {


            var id = this.getItemId(mi.id);
            var el = document.getElementById(id);
            if (!el) {
                // this can happen if we are re-rendering.
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

            itemsByStartTime.push(info);


        });
        this.mappedItems = {
            startTimeMap: itemsByStartTime,
        }
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

    renderItem(mi) {
        var seconds = (mi.end.getTime() - mi.start.getTime()) / 1000;

        var color = mi.video ? "green" : "gold";

        if (!mi.video) {
            return;
        }

        var w = Math.max(50, 10 * seconds);

        var motionItem = <div id={this.getItemId(mi.id)}
            onMouseDown={this.onMotionItemMouseDown.bind(this)}
            onMouseUp={this.onMotionItemMouseUp.bind(this)}
            title={mi.id}
            key={mi.id} item_id={mi.id} time={mi.start.getTime()} seconds={seconds} style={{
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

    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        const span = this.totalSeconds();

        var items = [];
        var mediaItems = this.props.items || [];

        const hourWidth = window.innerWidth / 4;
        const chunkSeconds = 300;

        var itemPos = 0;


        for (var i = 0; i < span.seconds; i += chunkSeconds) {

            // create background items for every n minutes.
            // if the item is an hour boundary, add the label.
            //

            var tStart = new Date(span.start + (1000 * i));

            var curHour = tStart.getHours()



            var itemDistanceSeconds = chunkSeconds;

            if (mediaItems[itemPos] && mediaItems[itemPos].start.getHours() === curHour) {
                itemDistanceSeconds = (toUnix(mediaItems[itemPos].start) - toUnix(tStart)) / 1000;
            }

            var topOfHour = i % 3600 === 0;
            var quarterHour = i % 900 === 0;

            var secondsWidth = Math.min(chunkSeconds, itemDistanceSeconds);

            var w = secondsWidth / 3600 * hourWidth;



            var label = <span>&nbsp;</span>;

            var borderLeft = "";
            if (topOfHour) {
                label = curHour;
                if (label > 12) {
                    label = (label % 12) + "p";
                } else {
                    label = (label || 12) + "a";
                }

                w = Math.max(w, 25);
                borderLeft = "thin white solid";
            } else if (quarterHour) {
                borderLeft = "thin silver dotted";
            }

            var hourItem = <div key={toUnix(tStart)} time={this.toUnix(tStart)} seconds={secondsWidth} style={{
                display: "inline-block",
                height: "100%",
                width: w + "px",
                borderLeft: borderLeft,
                color: "white",
                background: "navy",
                padding: "2px",
                textAlign: "left"
            }}>{label}</div>;

            items.push(hourItem);

            if (!mediaItems[itemPos]) {
                continue;
            }

            var lastItemEnd = toUnix(mediaItems[itemPos].start);

            // render time items when they are within the next chunk size.
            //
            var nextChunkEnd = tStart.getTime() + (chunkSeconds * 1000 * 2);

            while (true) {

                var nextItem = mediaItems[itemPos];

                if (!nextItem || nextItem.start.getHours() !== curHour) {
                    break;
                }

                // if the next item starts before the end of this chunk + 1,

                if (toUnix(nextItem.start) >= nextChunkEnd) {
                    break;
                }

                // then render this item
                var timeItem = this.renderItem(nextItem);
                if (timeItem) {
                    items.push(timeItem);
                }

                if (toUnix(nextItem.end) > nextChunkEnd) {
                    // consume the next chunk.
                    i += chunkSeconds;
                    nextChunkEnd += (1000 * chunkSeconds);
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
                msOverflowStyle: "none",
                userSelect: "none",

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