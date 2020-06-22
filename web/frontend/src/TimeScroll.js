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

        var item_id = Number(el.attributes.item_id && el.attributes.item_id.value);
        var item = this.props.items.find(i => i.id === item_id);

        if (this.props.onTimeChange) {
            this.props.onTimeChange(newTime, item);
        }
        console.log(`Scrolled to ${this.myRef.current.scrollLeft}`);
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

        var scrollLeft = this.myRef.current.scrollLeft;
        var item;

        for (var i = 0; i < this.mappedItems.scrollMap.length; i++) {
            item = this.mappedItems.scrollMap[i];
            if (!item) {
                continue;
            }


            if (item.scrollLeft <= scrollLeft && item.scrollLeft + item.width >= scrollLeft) {
                break;
            }
        }

        if (!item) {
            return null;
        }

        var ratio = ((item.scrollLeft + item.width) - scrollLeft) / item.width;

        return {
            element: item.element,
            ratio: ratio
        }


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


            var id = this.getMediaItemId(mi);
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

            var chunkIndex = Math.trunc((info.unixStart - ts.start) / (1000 * chunkSeconds));

            var list = chunks[chunkIndex] || [];
            list.push(info);
            chunks[chunkIndex] = list;

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

    getMediaItemId(mi) {
        return `ts-${this.idBase}-${mi.id}`;
    }

    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        const span = this.totalSeconds();

        var items = [];

        const hourWidth = window.innerWidth / 4;
        const chunkSeconds = 3600;


        for (var i = 0; i < span.seconds; i += chunkSeconds) {
            var t = new Date(span.start + (1000 * i));

            var iEnd = new Date(t.getTime() + chunkSeconds * 1000);
            var label = t.getHours();

            var mediaItems = this.props.items.filter(mi => mi.start >= t && mi.start < iEnd);

            var w = hourWidth;

            if (mediaItems && mediaItems.length) {
                w = 20;
            }

            if (label > 12) {
                label = (label % 12) + "p";
            } else {
                label = (label || 12) + "a";
            }
            var hourItem = <div key={"file" + i} time={this.toUnix(t)} seconds={chunkSeconds} style={{
                display: "inline-block",
                height: "100%",
                width: w + "px",
                borderLeft: "thin white solid",
                color: "white",
                background: "navy",
                padding: "2px"
            }}>{label}</div>;

            items.push(hourItem);



            mediaItems.forEach(mi => {
                var seconds = (mi.end.getTime() - mi.start.getTime()) / 1000;

                var color = mi.video ? "green" : "gold";

                if (!mi.video) {
                    return;
                }


                var motionItem = <div id={this.getMediaItemId(mi)}
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

                items.push(motionItem);

            });

        }


        return <div ref={this.myRef}
            onMouseDown={this.mouseDown.bind(this)}
            onMouseUp={this.mouseUp.bind(this)}
            onMouseMove={this.mouseMove.bind(this)}
            onScroll={this.onScroll.bind(this)}
            style={{
                width: "100%",
                height: "50px",
                background: "darkblue",
                overflowX: "auto",
                overflowY: "hidden",
                whiteSpace: "nowrap",
                msOverflowStyle: "none"

            }}>
            <div style={{
                position: "absolute",
                width: "5px",
                background: "yellow",
                height: "50px",
                left: "50%"
            }}></div>

            {items}

        </div>;
    }

}