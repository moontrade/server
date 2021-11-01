package app

import "sync"

// An Observer holds a channel that delivers the messages for all commands
// processed by a Service.
type Observer interface {
	Stop()
	C() <-chan Message
}

type observer struct {
	mon  *monitor
	msgC chan Message
}

func (o *observer) C() <-chan Message {
	return o.msgC
}

func (o *observer) Stop() {
	o.mon.obMu.Lock()
	defer o.mon.obMu.Unlock()
	if _, ok := o.mon.obs[o]; ok {
		delete(o.mon.obs, o)
		close(o.msgC)
	}
}

// Monitor represents an interface for sending and consuming command
// messages that are processed by a Service.
type Monitor interface {
	// Send a message to observers
	Send(msg Message)
	// NewObjser returns a new Observer containing a channel that will send the
	// messages for every command processed by the service.
	// Stop the observer to release associated resources.
	NewObserver() Observer
}

type monitor struct {
	s    *service
	obMu sync.Mutex
	obs  map[*observer]struct{}
}

func newMonitor(s *service) *monitor {
	m := &monitor{s: s}
	m.obs = make(map[*observer]struct{})
	return m
}

func (m *monitor) Send(msg Message) {
	if len(msg.Args) > 0 {
		// do not allow monitoring of certain system commands
		switch msg.Args[0] {
		case "raft", "machine", "auth", "cluster":
			return
		}
	}

	m.obMu.Lock()
	defer m.obMu.Unlock()
	for o := range m.obs {
		o.msgC <- msg
	}
}

func (m *monitor) NewObserver() Observer {
	o := new(observer)
	o.mon = m
	o.msgC = make(chan Message, 64)
	m.obMu.Lock()
	m.obs[o] = struct{}{}
	m.obMu.Unlock()
	return o
}
