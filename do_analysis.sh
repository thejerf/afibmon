#!/bin/bash

set -v -e

pwd

HEARTDATA="$1"

mkdir -p amp_frames freq_frames
~/syncdir/projects/githubgo/src/github.com/thejerf/afibmon/analyze "$1"
ffmpeg -r 15 -i freq_frames/frame%05d.png -vcodec libx264 -crf 25 frequency.mp4
ffmpeg -r 15 -i amp_frames/frame%05d.png -vcodec libx264 -crf 25 amplitude.mp4
