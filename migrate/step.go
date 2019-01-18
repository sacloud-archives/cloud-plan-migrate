package migrate

import (
	"fmt"
	"sync"
	"time"
)

const (
	statusDisabled = "-"
	statusWaiting  = "(waiting)"
	statusStarted  = "running(%ds)"
	statusDone     = "done"
	statusError    = "error"
)

type step struct {
	needProcess bool
	logPrefix   string
	started     bool
	startTime   time.Time
	done        bool
	err         error

	logger Logger
}

func (s *step) Error() error {
	return s.err
}

func (s *step) Status() string {
	if s.err != nil {
		return statusError
	}

	status := statusDisabled
	if s.needProcess {
		status = statusWaiting
	} else {
		return status
	}
	if s.started {
		status = fmt.Sprintf(statusStarted, int(s.elapsed().Seconds()))
	}
	if s.done {
		status = statusDone
	}
	return status
}

func (s *step) elapsed() time.Duration {
	return time.Since(s.startTime)
}

func (s *step) start() {
	s.startTime = time.Now()
	s.started = true
	s.logStarted()
}

func (s *step) logStarted() {
	if !s.needProcess {
		return
	}

	if s.logger != nil {
		s.logger.Printf("%s started%s", s.logPrefix, newline)
	}
}

func (s *step) finalize() {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	if !s.done {
		s.done = true
		s.logDone()
	}
}

func (s *step) logDone() {
	if !s.needProcess {
		return
	}
	if s.logger != nil {
		s.logger.Printf("%s finished (elapsed:%ds)%s", s.logPrefix, int(s.elapsed().Seconds()), newline)
	}
}

func (s *step) setError(err error) {
	s.err = err
	s.logError()
}

func (s *step) logError() {
	if !s.needProcess {
		return
	}
	if s.logger != nil {
		s.logger.Printf("%s error: %s%s", s.logPrefix, s.err, newline)
	}
}
