package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/tinogoehlert/gostreamer/gocv"

	"github.com/tinogoehlert/gostreamer"
	"gocv.io/x/gocv"
)

func main() {
	runtime.LockOSThread()
	gst, _ := gostreamer.NewGstreamer()
	fdsink := gostreamer.NewFdSink()
	gst.AddStr("avfvideosrc").
		AddStr("videoconvert").
		AddStr("video/x-raw, framerate=25/1,format=BGR").
		AddStr("videoscale").
		AddSink(fdsink)

	imager := gocvstreamer.NewCvImager(fdsink)
	window := gocv.NewWindow("Simple test")
	defer window.Close()

	imager.ProduceMat(context.Background(), func(mat gocv.Mat) error {
		window.IMShow(mat)
		if window.WaitKey(1) >= 0 {
			return fmt.Errorf("interrupted")
		}
		return nil
	})
}
