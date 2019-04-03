package gostreamer

import (
	"context"
	"fmt"
	"os/exec"
)

// Gstreamer represents a gstreamer pipe
type Gstreamer struct {
	cmdName string
	pipe    []GstreamerPlugin
	args    []string
}

const defaultGstBin = "gst-launch-1.0"

// NewGstreamer creates a new Gstreamer instance
func NewGstreamer(args ...string) (*Gstreamer, error) {
	return NewGstreamerFromPath(defaultGstBin, args...)
}

// NewGstreamerFromPath creates a new Gstreamer instance, uses given path to gst-launch binary
func NewGstreamerFromPath(gstBin string, args ...string) (*Gstreamer, error) {
	if err := exec.Command(gstBin, "--version").Run(); err != nil {
		return nil, err
	}
	gst := &Gstreamer{
		cmdName: gstBin,
		args:    []string{"-v"},
	}
	gst.args = append(gst.args, args...)
	return gst, nil
}

// PipeString outputs the current pipe
func (gp *Gstreamer) PipeString() (out string) {
	for _, p := range gp.pipe {
		out += fmt.Sprintf("%s ! ", p)
	}
	return
}

// Add add a GstreamerPlugin to the pipe
func (gp *Gstreamer) Add(p GstreamerPlugin) *Gstreamer {
	gp.pipe = append(gp.pipe, p)
	return gp
}

//AddStr Creates a GstreamerGenericPlugin from given string and adds it to the pipe
func (gp *Gstreamer) AddStr(p ...string) *Gstreamer {
	gp.pipe = append(gp.pipe, GstreamerGenericPlugin{
		cmds: p,
	})
	return gp
}

//AddSink adds a sink to pipe
func (gp *Gstreamer) AddSink(sink Sink) Sink {
	gp.pipe = append(gp.pipe, sink)
	sink.bind(gp)
	return sink
}

func (gp *Gstreamer) createCmd(ctx context.Context) *exec.Cmd {
	var pipeArgs = make([]string, 0, len(gp.pipe)*2)
	pipeArgs = append(pipeArgs, gp.args...)
	for i, p := range gp.pipe {
		for _, cmd := range p.Cmds() {
			pipeArgs = append(pipeArgs, cmd)
		}

		if i != len(gp.pipe)-1 {
			pipeArgs = append(pipeArgs, "!")
		}
	}

	return exec.CommandContext(ctx, "gst-launch-1.0", pipeArgs...)
}
