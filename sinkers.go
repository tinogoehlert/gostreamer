package gostreamer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
)

var (
	capsRe = regexp.MustCompile(`(?m)/GstPipeline:pipeline(?P<id>[0-99])/(?P<name>.*?):.*?caps = video/(?P<type>.*?)|width=\(int\)(?P<width>[0-9]{1,5})|height=\(int\)(?P<height>[0-9]{1,5})|framerate=\(fraction\)(?P<framerate>[0-9]{1,3})/[0-9]|format=\(string\)(?P<format>[A-Za-z]{1,5})`)
)

const (
	// GstCapsFilter cap we will rely on
	GstCapsFilter = "GstCapsFilter"
)

// Sink generic sink interface
type Sink interface {
	bind(gst *Gstreamer)
	Gst() *Gstreamer
	Cmds() []string
	Run()
	Start(ctx context.Context) (*Caps, error)
}

// DataSink represents a sink which returns data
type DataSink interface {
	Sink
	DataChan() chan []byte
}

// Caps represents caps
type Caps struct {
	Width       int
	Height      int
	Framerate   int
	PixelFormat string
	Type        string
	Name        string
}

// Sinker generic sinker
type Sinker struct {
	gst         *Gstreamer
	cmds        []string
	caps        map[string]Caps
	mut         sync.Mutex
	defaultCaps Caps
	dataChan    chan []byte
}

func (s *Sinker) bind(gst *Gstreamer) {
	s.caps = make(map[string]Caps)
	s.gst = gst
}

// Gst returns the parent Gstreamer instance
func (s *Sinker) Gst() *Gstreamer {
	return s.gst
}

// Cmds returns list of given commands
func (s *Sinker) Cmds() []string {
	return s.cmds
}

func (s *Sinker) setCmds(cmds []string) {
	s.cmds = cmds
}

// DataChan returns a channel to receive the data from the sink.
func (s *Sinker) DataChan() chan []byte {
	return s.dataChan
}

// CapsAll returns all captured caps
func (s *Sinker) CapsAll() *map[string]Caps {
	return &s.caps
}

func (s *Sinker) parseCaps(line string) (*Caps, bool) {
	var caps Caps
	names := capsRe.SubexpNames()
	matches := capsRe.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return nil, false
	}
	for _, match := range matches {
		for i, n := range match {
			if i == 0 || n == "" {
				continue
			}
			switch names[i] {
			case "name":
				caps.Name = n
			case "type":
				caps.Type = n
			case "width":
				caps.Width, _ = strconv.Atoi(n)
			case "height":
				caps.Height, _ = strconv.Atoi(n)
			case "format":
				caps.PixelFormat = n
			}
		}
	}
	if caps.Name != "" {
		s.mut.Lock()
		s.caps[caps.Name] = caps
		s.mut.Unlock()
	}
	return &caps, caps.Name != ""
}

// Run runs the pipeline with the given sink
func (s *Sinker) Run() {
	cmd := s.Gst().createCmd(context.Background())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// CommonSink represents a common sink (e.g. videosink, autovideosink or udpsink)
type CommonSink struct {
	Sinker
}

// NewCommonSink creates a new common sink (e.g. videosink, autovideosink or udpsink)
func NewCommonSink(cmds ...string) *CommonSink {
	vs := &CommonSink{}
	vs.setCmds(cmds)
	return vs
}

// Start starts the pipeline with the given sink (blocking)
func (s *CommonSink) Start(ctx context.Context) (*Caps, error) {
	cmd := s.Gst().createCmd(ctx)
	cmd.Start()
	return nil, nil
}

// FdSink represents an fdsink element
type FdSink struct {
	Sinker
}

// NewFdSink creates a new fdsink
func NewFdSink() *FdSink {
	fds := &FdSink{}
	fds.dataChan = make(chan []byte)
	fds.setCmds([]string{"fdsink", "fd=3"})
	return fds
}

// Start starts the pipeline with the given fdsink.
// Blocks until GstCapsFilter was found.
// use DataChan() to receive image buffers
func (s *FdSink) Start(ctx context.Context) (*Caps, error) {
	cmd := s.Gst().createCmd(ctx)

	r, w, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("could not create pipe: %s", err.Error())
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not fetch stdout: %s", err.Error())
	}
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{w}

	stdoutScanner := bufio.NewScanner(stdout)
	cmd.Start()
	var capsChan = make(chan *Caps)
	go func() {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if caps, ok := s.parseCaps(line); ok {
				if caps.Name == GstCapsFilter {
					capsChan <- caps
					s.startBuffReader(r, caps)
				}
			}
		}
	}()

	return <-capsChan, nil
}

func (s *FdSink) startBuffReader(fd *os.File, caps *Caps) {
	var depth int
	switch caps.PixelFormat {
	case "RGB", "BGR":
		depth = 3
	case "RGBA", "BGRA":
		depth = 4
	default:
		return
	}
	buff := make([]byte, (caps.Width*depth)*caps.Height)
	for sz := 0; ; sz = 0 {
		for sz < (caps.Width*depth)*caps.Height {
			n, _ := fd.Read(buff[sz:])
			sz += n
		}
		if len(buff) > 0 {
			s.dataChan <- buff[:sz]
		}
	}
}
