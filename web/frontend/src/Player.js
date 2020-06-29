


import React from 'react';
import ReactPlayer from 'react-player';
import { toUnix } from './time';

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

        if (!this.videoReady) {
            return;
        }
        var delta = toUnix(this.props.position) - toUnix(this.props.file.start);
        this.player.seekTo(delta / 1000, "seconds");
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


    render() {

        if (this.props.file == null) {
            return <div></div>;
        }


        var val = this.props.file;


        if (val.type === 1) {
            return <ReactPlayer
                controls
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

        if (val.type === 0) {
            return <img alt="view" src={val.path} style={{
                height: "100%"
            }} />;
        }

        return <div></div>;

    }
}

