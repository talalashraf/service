package service

import (
  "os"
  "os/signal"
  "sync"
  "time"
)

type Service struct {
  wg          sync.WaitGroup
  quitChannel chan bool
}

// use once per application
func NewService() *Service {
  s := Service{
    quitChannel: make(chan bool, 0),
    wg:          sync.WaitGroup{},
  }
  go s.processSignals()
  return &s
}

func (s *Service) processSignals() {
  sigChan := make(chan os.Signal, 0)
  signal.Notify(sigChan, os.Interrupt, os.Kill)
  <-sigChan
  close(s.quitChannel)
}

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

func (s *Service) ChannelReader(c chan interface{}, f func(data *interface{})) {
  quit := false
  for !quit {
    select {
    case <-s.quitChannel:
      quit = true
    case data <- c:
      f(data)
    }
  }
  s.wg.Done()
}

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
  s.wg.Done()
}

func (s *Service) StartLoop(f func()) {
  s.Start(func() {
    s.Loop(f)
  })
}

func (s *Service) Start(f func()) {
  s.wg.Add(1)
  go f()
}

func (s *Service) Wait() {
  s.wg.Wait()
}
