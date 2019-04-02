package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/tinogoehlert/gostreamer"
	"github.com/tinogoehlert/gostreamer/gocv"
	"gocv.io/x/gocv"
)

func init() {
	// start udp RTP server
	go func() {
		gst, _ := gostreamer.NewGstreamer()
		gst.AddStr("avfvideosrc").
			AddStr("video/x-raw,framerate=20/1").
			AddStr("videoscale").
			AddStr("videoconvert").
			AddStr("x264enc", "tune=zerolatency", "bitrate=500", "speed-preset=superfast").
			AddStr("rtph264pay").
			AddSink(gostreamer.NewCommonSink("udpsink", "host=127.0.0.1", "port=5000")).
			Run()
	}()
}

func main() {
	runtime.LockOSThread()
	gst, _ := gostreamer.NewGstreamer()
	fdsink := gostreamer.NewFdSink()
	gst.AddStr("udpsrc", "port=5000", "caps=\"application/x-rtp\"").
		AddStr("rtpjitterbuffer").
		AddStr("rtph264depay").
		AddStr("avdec_h264").
		AddStr("videoconvert").
		AddStr("video/x-raw, format=BGR").
		AddSink(fdsink)

	imager := gocvstreamer.NewCvImager(fdsink)
	window := gocv.NewWindow("RTP Test")
	defer window.Close()

	imager.ProduceMat(context.Background(), func(mat gocv.Mat) error {
		window.IMShow(mat)
		if window.WaitKey(1) >= 0 {
			return fmt.Errorf("interrupted")
		}
		return nil
	})
}
