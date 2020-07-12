


import React from 'react';
import ReactPlayer from 'react-player';
import { second } from './time';

export default class Player extends React.Component {

    onVideoReady() {
        // this gets called on seek
        if (this.videoReady) {
            return;
        }
        this.videoReady = true;
        this.setVideoPosition();
    }

    setVideoPosition() {
        if (!this.props.file || !this.props.file.path) {
            return;
        }

        if (!this.videoReady || this.live || !this.props.position) {
            return;
        }

        var delta = this.props.position.delta(this.props.file.start, second);
        this.player.seekTo(delta, "seconds");
    }


    playerRef = player => {
        this.player = player
    }

    shouldComponentUpdate(nextProps, _nextState) {

        var oldFile = this.props.file && this.props.file.id;
        var newFile = nextProps.file && nextProps.file.id;

        if (oldFile === newFile && this.props.position !== nextProps.position) {
            this.setVideoPosition();
            return false;
        }
        this.videoReady = false;

        return true;
    }

    renderVideoPlayer(val, live) {


        this.live = live;

        return <ReactPlayer
            controls={!live}
            ref={this.playerRef}
            url={val.path}
            width="100%"
            height="100%"
            playsinline={true}
            playing={true}
            volume={0}
            muted={true}
            style={{
                height: "100%"
            }}
            onReady={this.onVideoReady.bind(this)}
        />;
    }

    renderImage(val) {
        return <img alt="view" src={val.path} style={{
            height: "100%"
        }} />;
    }

    render() {

        if (this.props.file == null) {
            return <div></div>;
        }


        var val = this.props.file;

        switch (val.type) {
            case 1:
            case 2:
                return this.renderVideoPlayer(val, val.type === 2)

            case 0:
                return this.renderImage(val);

            default:
                console.error(`Unknown media file type ${val.type}`);
                return <div>Bad file type</div>
        }

    }
}

