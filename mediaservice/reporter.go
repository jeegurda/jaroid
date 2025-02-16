package mediaservice

import (
	"bufio"
	"io"
	"sync"
	"time"
)

// Reporter provides helper for rate-limited diagnostic messages from downloader
// All messages submitted at higher frequency than rate will be lost.
// Caller should read from Messages() on higher or equal frequency than rate specified, otherwise lagged messages
// are going to be lost.
type Reporter interface {
	Messages() <-chan string
	Submit(msg string, force bool)
	Close()
	CanRead() bool
	ReadLine() (string, error)
}

type dummyReporter struct {
}

func (*dummyReporter) Messages() <-chan string {
	return nil
}

func (*dummyReporter) Submit(msg string, force bool) {

}

func (*dummyReporter) Close() {

}

func (*dummyReporter) CanRead() bool {
	return false
}

func (*dummyReporter) ReadLine() (string, error) {
	return "", nil
}

// NewDummyReporter returns new dummy reporter implementation
func NewDummyReporter() Reporter {
	return &dummyReporter{}
}

// NewReporter returns new reporter instance with provided rate limit and buffer size for Messages() channel
func NewReporter(rate time.Duration, buf int, reader io.Reader) Reporter {
	var br *bufio.Reader

	if reader != nil {
		br = bufio.NewReader(reader)
	}

	rep := &reporterImpl{
		m:        &sync.Mutex{},
		messages: make(chan string, buf),
		rate:     rate,
		reader:   br,
	}

	return rep
}

type reporterImpl struct {
	m        *sync.Mutex
	reader   *bufio.Reader
	messages chan string
	rate     time.Duration
	last     int64
	acc      int64
}

func (r *reporterImpl) Messages() <-chan string {
	r.m.Lock()
	defer r.m.Unlock()

	return r.messages
}

func (r *reporterImpl) Close() {
	r.m.Lock()
	defer r.m.Unlock()

	if r.messages == nil {
		return
	}

	close(r.messages)

	r.messages = nil
}

func (r *reporterImpl) Submit(msg string, force bool) {
	now := time.Now().UnixNano()

	r.m.Lock()
	defer r.m.Unlock()

	if r.messages == nil {
		return
	}

	var el int64

	el, r.last = now-r.last, now

	if el < 0 && !force {
		return
	}

	r.acc += el

	if r.acc < int64(r.rate) {
		if !force {
			return
		}
	} else {
		r.acc = 0
	}

	r.messages <- msg
}

func (r *reporterImpl) CanRead() bool {
	return r.reader != nil
}

func (r *reporterImpl) ReadLine() (string, error) {
	if !r.CanRead() {
		return "", nil
	}

	return r.reader.ReadString('\n')
}
