# gostreamer

go package for playing with the gstreamer (gstlaunch) cli.

## supported image formats (DataSinks):
- RGB
- BGR
- RGBA
- BGRA

## opencv support:
- you can also produce gocv Mat objects from video sinks

## TODO:
- audiosink support (DataSink)
- better documentation
- go image support
- sdl image support
- support more image formats for DataSinks

## NOTES:
- FdSink does not work on windows

example: 
```go
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
```
