package gostreamer

import (
	"context"
	"fmt"
	"image"
	"image/color"
)

// Imager produces go image objects
type Imager struct {
	sink    DataSink
	imgChan chan image.Image
	caps    Caps
}

func NewImager(sink DataSink) *Imager {
	return &Imager{
		sink:    sink,
		imgChan: make(chan image.Image),
	}
}

func (im *Imager) ProduceImage(ctx context.Context, cb func(mat image.Image) error) error {
	caps, err := im.sink.Start(ctx)
	if err != nil {
		return err
	}

	for {
		b := <-im.sink.DataChan()
		img, err := byteToImage(b, caps.Width, caps.Height, caps.Channels)
		if err != nil {
			return err
		}
		if err := cb(img); err != nil {
			return err
		}
	}
}

func byteToImage(b []byte, width, height, channels int) (image.Image, error) {
	switch channels {
	case 1:
		return byteToGrayscaleImage(b, width, height), nil
	case 3, 4:
		return byteToColorImage(b, width, height, channels), nil
	}
	return nil, fmt.Errorf("unsupported channel size")
}

func byteToColorImage(b []byte, width, height, channels int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{
		R: uint8(0),
		G: uint8(0),
		B: uint8(0),
		A: uint8(255),
	}

	var step = width * channels
	for y := 0; y < height; y++ {
		for x := 0; x < step; x = x + channels {
			c.B = uint8(b[y*step+x])
			c.G = uint8(b[y*step+x+1])
			c.R = uint8(b[y*step+x+2])
			if channels == 4 {
				c.A = uint8(b[y*step+x+3])
			}
			img.SetRGBA(int(x/channels), y, c)
		}
	}
	return img
}

func byteToGrayscaleImage(b []byte, width, height int) image.Image {
	img := image.NewGray(image.Rect(0, 0, width, height))
	c := color.Gray{Y: uint8(0)}

	var step = width
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c.Y = uint8(b[y*step+x])
			img.SetGray(x, y, c)
		}
	}
	return img
}
