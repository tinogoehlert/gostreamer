package gocvstreamer

import (
	"context"

	"github.com/tinogoehlert/gostreamer"
	"gocv.io/x/gocv"
)

// CvImager Object to generate gocv materials
type CvImager struct {
	sink    gostreamer.DataSink
	matChan chan gocv.Mat
	caps    gostreamer.Caps
}

func NewCvImager(sink gostreamer.DataSink) *CvImager {
	return &CvImager{
		sink:    sink,
		matChan: make(chan gocv.Mat),
	}
}

func (cvi *CvImager) ProduceMat(ctx context.Context, cb func(mat gocv.Mat) error) error {

	caps, err := cvi.sink.Start(ctx)
	if err != nil {
		return err
	}

	for {
		b := <-cvi.sink.DataChan()
		img, err := gocv.NewMatFromBytes(
			caps.Height,
			caps.Width,
			gocv.MatTypeCV8UC3,
			b,
		)
		if err != nil || img.Empty() {
			continue
		}
		err = cb(img)
		img.Close()
		if err != nil {
			return err
		}
	}
}
