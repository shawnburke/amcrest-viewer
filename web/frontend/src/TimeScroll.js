import React from 'react';



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

        return Number(t);
    }

    shouldComponentUpdate(nextProps, nextState) {


        if (!this.anchor) {

            if (nextProps.startTime === this.props.startTime &&
                nextProps.endTime === this.props.endTime &&
                nextProps.position !== this.props.position) {
                var t = this.toUnix(nextProps.position);
                this.scrollToTime(t);
                return false;
            }
        }


        return true;
    }

    getElementTime(el) {
        var childTime = el.attributes.time && el.attributes.time.value;

        if (!childTime) {
            return null;
        }

        return this.toUnix(childTime);
    }

    scrollToTime(t) {
        var left = 0;
        const node = this.myRef.current;
        var indicator = node.clientWidth / 2;

        var bestFit;
        var bestFitDuration;

        for (var i = 0; i < node.childNodes.length; i++) {
            var child = node.childNodes[i];

            var childTime = this.getElementTime(child);

            if (childTime) {

                var seconds = Number(child.attributes.seconds.value);
                var childEnd = Number(childTime) + (1000 * seconds);

                if (t >= childTime && t <= childEnd) {
                    if (!bestFitDuration || bestFitDuration > seconds) {
                        bestFit = child;
                    }
                }
            }

            left += child.getClientRects().width;
        }

        if (bestFit) {
            var childTime = this.getElementTime(bestFit);
            var seconds = Number(child.attributes.seconds.value);

            var ratio = ((t - childTime) / 1000) / seconds;
            this.scrollToElement(bestFit, ratio);
        }
    }


    elementFromScroll(buffer) {
        var parent = this.myRef.current;
        buffer = buffer || 2;
        var posCenter = parent.scrollLeft + (parent.clientWidth / 2);
        var posLeft = posCenter - buffer / 2;
        var posRight = posCenter + buffer / 2;

        for (var i = 0; i < parent.childNodes.length; i++) {
            var child = parent.childNodes[i];
            var childLeft = child.offsetLeft - parent.offsetLeft;
            var childRight = childLeft + child.clientWidth;
            if (childLeft <= posCenter && childRight >= posCenter) {
                var ratio = (posCenter - childLeft) / child.clientWidth;
                return {
                    element: child,
                    ratio: ratio
                };
            }
        }
        return null;
    }

    scrollToElement(el, ratio) {

        var cur = this.elementFromScroll(10);

        if (cur && el == cur.element) {
            return;
        }

        var parent = this.myRef.current;
        var center = parent.clientWidth / 2;
        var left = el.offsetLeft - (parent.offsetLeft);

        var extra = 0;

        if (ratio) {
            extra = el.clientWidth * ratio;
        }

        var newScroll = (left + extra) - center;
        console.log(`Scrolling ${parent.scrollLeft} => ${newScroll}`);
        parent.scrollLeft = newScroll;
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
    render() {

        if (!this.props.startTime || !this.props.endTime) {
            return <div>XXX</div>
        }

        const hour = 60 * 60 * 1000;
        const month = hour * 24 * 30;

        var start = this.boxTime(this.props.startTime, new Date().getTime() - month, new Date());
        var end = this.boxTime(this.props.endTime, new Date().getTime() - month, new Date());

        // buffer an hour on either end, snap
        // to hour boundaries
        start -= (start % hour) + hour;
        end += (hour - (end % hour)) + hour;

        const spanSeconds = (end - start) / 1000;

        var items = [];

        const hourWidth = window.innerWidth / 4;
        const chunkSeconds = 3600;


        for (var i = 0; i < spanSeconds; i += chunkSeconds) {
            var t = new Date(start + (1000 * i));

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


                var motionItem = <div id={`ts-${this.idBase}-${mi.id}`}
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