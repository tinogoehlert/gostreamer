package gostreamer

// GstreamerPlugin gstreamer plugin interface
type GstreamerPlugin interface {
	Cmds() []string
}

// GstreamerGenericPlugin generic gstreamer plugin
type GstreamerGenericPlugin struct {
	cmds []string
}

func (p GstreamerGenericPlugin) Cmds() []string {
	return p.cmds
}

// Queue returns the queue plugin
func Queue() GstreamerPlugin {
	return GstreamerGenericPlugin{
		cmds: []string{"queue"},
	}
}
