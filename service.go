package service

import (
  "os"
  "os/signal"
  "sync"
  "time"
)

type Service struct {
  wg          sync.WaitGroup
  quitChannel chan struct{}
}

// use once per application
func NewService() *Service {
  s := Service{
    quitChannel: make(chan struct{}, 0),
    wg:          sync.WaitGroup{},
  }
  go s.processSignals()
  return &s
}

//
func (s *Service) processSignals() {
  sigChan := make(chan os.Signal, 0)
  signal.Notify(sigChan, os.Interrupt, os.Kill)
  <-sigChan
  s.Quit()
}

// Request that the service exits. Usually not needed. Signals are processed for you.
func (s *Service) Quit() {
  close(s.quitChannel)
}

// call a callback function every time.Duration. If the function takes longer than time.Duration to execute, it will not be called while it's executing, and multiple ticks will be deduped. (default time.Ticker behavior)
func (s *Service) Timer(duration time.Duration, f func()) {
  t := time.NewTicker(duration)
  quit := false
  for !quit {
    select {
    case <-s.quitChannel:
      t.Stop()
      quit = true
    case <-t.C:
      f()
    }
  }
  s.wg.Done()
}

// call a callback function when there is data ready to be read from the channel
func (s *Service) ChannelReader(c chan interface{}, f func(*interface{}, bool)) {
  var data interface{}
  var ok bool
  quit := false
  for !quit {
    select {
    case <-s.quitChannel:
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
    <-s.quitChannel
    f()
    s.wg.Done()
  }()
}

// do this over and over as fast as possible.
func (s *Service) Loop(f func()) {
  quit := false
  for !quit {
    select {
    case <-s.quitChannel:
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
