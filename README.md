Every service in Go starts the same way:

1. Start a go routine to run some service
2. Set up signal.Notify to process signals
3. use channels and wait groups to manage closing cleanly

That is, unless you're a fan of abruptly stopping execution and don't care about your goroutines finishing their work. You didn't need that buffer.Flush(), right?

We don't have to rewrite this every time.

    // create a new service (one per app)
    serv := service.NewService()

    // start a new goroutine for it
    serv.Start(func() {
      // setup

      serv.Loop(func() {
        // do this over and over until we're asked to quit.
      })

      // cleaup
    })

    // wait for all started services to finish before exiting
    serv.Wait()

Not all services are dumb loops. Some want to do things every x duration:

    serv.Start(func() {
      serv.Timer(10*time.Second, func() {
        // do this every 10 seconds.
      })
    })

and some just need to receive data from a channel:

    serv.Start(func() {
      serv.ChannelReader(myChan, func(data *interface{}, ok bool) {
        // do something with data
      })
    })
