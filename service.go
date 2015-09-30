package service

import (
  "os"
  "os/signal"
  "sync"
  "time"
)

type Service struct {
  wg          sync.WaitGroup
  QuitChannel chan struct{}
  IsQuitting  bool
}

// use once per application
func NewService() *Service {
  s := Service{
    QuitChannel: make(chan struct{}, 0),
    wg:          sync.WaitGroup{},
  }
  go s.processSignals()
  return &s
}

// used internally to watch for quit signal
func (s *Service) processSignals() {
  sigChan := make(chan os.Signal, 0)
  signal.Notify(sigChan, os.Interrupt, os.Kill)
  <-sigChan
  s.Quit()
}

// Request that the service exits. Usually not needed. Signals are processed for you.
func (s *Service) Quit() {
  s.IsQuitting = true
  close(s.QuitChannel)
}

// call a callback function every time.Duration. If the function takes longer than time.Duration to execute, it will not be called while it's executing, and multiple ticks will be deduped. (default time.Ticker behavior)
func (s *Service) Timer(duration time.Duration, f func()) {
  t := time.NewTicker(duration)
  quit := false
  for !quit {
    select {
    case <-s.QuitChannel:
      t.Stop()
      quit = true
    case <-t.C:
      t.Stop()
      f()
      t = time.NewTicker(duration)
    }
  }
}

// A dynamic timer. expects a function that returns a time.Duration value, and will sleep that long before calling it again.
// It will be called immediately on load.
// The timer will not be called while the function is running.
func (s *Service) DynamicTimer(f func() time.Duration) {
  t := time.NewTicker(1 * time.Microsecond)
  quit := false
  for !quit {
    select {
    case <-s.QuitChannel:
      t.Stop()
      quit = true
    case <-t.C:
      t.Stop()
      newDuration := f()
      if !quit {
        t = time.NewTicker(newDuration)
      }
    }
  }
}

func (s *Service) StartTimer(duration time.Duration, f func()) {
  s.Start(func() {
    s.Timer(duration, f)
  })
}

// call a callback function when there is data ready to be read from the channel
// Todo: figure out how to make this work with any kind of channel...
func (s *Service) ChannelReader(c chan interface{}, f func(*interface{}, bool)) {
  var data interface{}
  var ok bool
  quit := false
  for !quit {
    select {
    case <-s.QuitChannel:
      quit = true
    case data, ok = <-c:
      f(&data, ok)
    }
  }
}

// A callback function to happen when a shutdown is triggered. Wait() will not return until this callback has completed.
func (s *Service) OnQuit(f func()) {
  s.wg.Add(1)
  go func() {
    <-s.QuitChannel
    f()
    s.wg.Done()
  }()
}

// do this over and over as fast as possible.
func (s *Service) Loop(f func()) {
  quit := false
  for !quit {
    select {
    case <-s.QuitChannel:
      quit = true
    default:
      f()
    }
  }
}

// start a goroutine and do this function over and over as fast as possible. This is the same as calling Start() with an inner Loop()
func (s *Service) StartLoop(f func()) {
  s.Start(func() {
    s.Loop(f)
  })
}

// Starts a go routine to run the passed function, but with a waitgroup around it.
func (s *Service) Start(f func()) {
  s.wg.Add(1)
  go func() {
    f()
    s.wg.Done()
  }()
}

// wait for all "Start"ed workers to finish before continuing
func (s *Service) Wait() {
  s.wg.Wait()
}
