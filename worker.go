package teeny

import "runtime"

type workRequest struct {
	// jobChan is used to send the payload to this worker.
	jobChan chan<- Input

	// retChan is used to read the result from this worker.
	retChan <-chan Output

	// interruptFunc can be called to cancel a running job. When called it is no
	// longer necessary to read from retChan.
	interruptFunc func()
}

//------------------------------------------------------------------------------

// workerWrapper takes a Worker implementation and wraps it within a goroutine
// and channel arrangement. The workerWrapper is responsible for managing the
// lifetime of both the Worker and the goroutine.
type workerWrapper struct {
	worker        Worker
	interruptChan chan struct{}

	// reqChan is NOT owned by this type, it is used to send requests for work.
	reqChan chan<- workRequest

	// closeChan can be closed in order to cleanly shutdown this worker.
	closeChan chan struct{}

	// closedChan is closed by the run() goroutine when it exits.
	closedChan chan struct{}
}

func newWorkerWrapper(
	poolType PoolType,
	reqChan chan<- workRequest,
	worker Worker,
) *workerWrapper {
	w := workerWrapper{
		worker:        worker,
		interruptChan: make(chan struct{}),
		reqChan:       reqChan,
		closeChan:     make(chan struct{}),
		closedChan:    make(chan struct{}),
	}

	go w.run(poolType)

	return &w
}

//------------------------------------------------------------------------------

func (w *workerWrapper) interrupt() {
	close(w.interruptChan)
	w.worker.Interrupt()
}

func (w *workerWrapper) run(poolType PoolType) {
	if poolType == PoolTypeThread {
		runtime.LockOSThread()
	}

	jobChan, retChan := make(chan Input), make(chan Output)
	defer func() {
		w.worker.Terminate()
		close(retChan)
		close(w.closedChan)

		if poolType == PoolTypeThread {
			runtime.UnlockOSThread()
		}
	}()

	for {
		// NOTE: Blocking here will prevent the worker from closing down.
		w.worker.BlockUntilReady()
		select {
		case w.reqChan <- workRequest{
			jobChan:       jobChan,
			retChan:       retChan,
			interruptFunc: w.interrupt,
		}:
			select {
			case input := <-jobChan:
				var result Output
				logic, err := w.worker.GetLogic(input.Name)
				if err != nil {
					result = Output{
						Err: err,
					}
				} else {
					result = logic.Process(input)
				}
				select {
				case retChan <- result:
				case <-w.interruptChan:
					w.interruptChan = make(chan struct{})
				}
			case _, _ = <-w.interruptChan:
				w.interruptChan = make(chan struct{})
			}
		case <-w.closeChan:
			return
		}
	}
}

func (w *workerWrapper) stop() {
	close(w.closeChan)
}

func (w *workerWrapper) join() {
	<-w.closedChan
}
