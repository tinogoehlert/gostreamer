package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"

	"github.com/tinogoehlert/gostreamer"
	"github.com/tinogoehlert/gostreamer/gocv"
	"gocv.io/x/gocv"
)

func main() {
	runtime.LockOSThread()
	if len(os.Args) < 3 {
		fmt.Println("How to run:\ndnn-detection [modelfile] [configfile] ([backend] [device])")
		return
	}

	// parse args
	model := os.Args[1]
	config := os.Args[2]
	backend := gocv.NetBackendDefault
	if len(os.Args) > 3 {
		backend = gocv.ParseNetBackend(os.Args[3])
	}

	target := gocv.NetTargetCPU
	if len(os.Args) > 4 {
		target = gocv.ParseNetTarget(os.Args[4])
	}

	gst, _ := gostreamer.NewGstreamer()
	fdsink := gostreamer.NewFdSink()
	gst.AddStr("avfvideosrc").
		AddStr("videoconvert").
		AddStr("video/x-raw, framerate=25/1,format=BGR").
		AddStr("videoscale").
		AddSink(fdsink).Start(context.Background())

	window := gocv.NewWindow("Simple test")
	defer window.Close()

	// open DNN object tracking model
	net := gocv.ReadNet("frozen_inference_graph.pb", "ssd_mobilenet_v2_coco_2018_03_29.pbtxt")
	if net.Empty() {
		fmt.Printf("Error reading network model from : %v %v\n", model, config)
		return
	}
	defer net.Close()
	net.SetPreferableBackend(gocv.NetBackendType(backend))
	net.SetPreferableTarget(gocv.NetTargetType(target))

	ratio := 2.0
	mean := gocv.NewScalar(104, 177, 123, 0)
	swapRGB := true

	imager := gocvstreamer.NewCvImager(fdsink)

	imager.ProduceMat(context.Background(), func(img gocv.Mat) error {
		// convert image Mat to 300x300 blob that the object detector can analyze
		blob := gocv.BlobFromImage(img, ratio, image.Pt(300, 300), mean, swapRGB, false)
		// feed the blob into the detector
		net.SetInput(blob, "")
		// run a forward pass thru the network
		prob := net.Forward("")
		performDetection(&img, prob)

		prob.Close()
		blob.Close()

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			return fmt.Errorf("key pressed")
		}
		return nil
	})
}

// performDetection analyzes the results from the detector network,
// which produces an output blob with a shape 1x1xNx7
// where N is the number of detections, and each detection
// is a vector of float values
// [batchId, classId, confidence, left, top, right, bottom]
func performDetection(frame *gocv.Mat, results gocv.Mat) {
	for i := 0; i < results.Total(); i += 7 {
		id := int(results.GetFloatAt(0, i+1))
		var class = fmt.Sprintf("%d", id)
		switch id {
		case 1:
			class = "Person"
		case 17:
			class = "Cat"
		case 18:
			class = "Dog"
		case 62:
			class = "Chair"
		case 72:
			class = "Screen"
		}
		confidence := results.GetFloatAt(0, i+2)
		if confidence > 0.8 {

			left := int(results.GetFloatAt(0, i+3) * float32(frame.Cols()))
			top := int(results.GetFloatAt(0, i+4) * float32(frame.Rows()))
			right := int(results.GetFloatAt(0, i+5) * float32(frame.Cols()))
			bottom := int(results.GetFloatAt(0, i+6) * float32(frame.Rows()))
			gocv.PutText(frame, class, image.Point{left, top}, gocv.FontHersheyDuplex, 1.5, color.RGBA{
				R: 255,
				G: 255,
				B: 255,
			}, 3)
			gocv.Rectangle(frame, image.Rect(left, top, right, bottom), color.RGBA{0, 255, 0, 0}, 2)
		}
	}
}
