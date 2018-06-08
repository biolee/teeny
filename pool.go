package teeny

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type PoolType int

const (
	PoolTypeGoroutine PoolType = iota
	PoolTypeThread
)

// Errors that are used throughout the Tunny API.
var (
	ErrPoolNotRunning = errors.New("the pool is not running")
	ErrWorkerClosed   = errors.New("worker was closed")
	ErrJobTimedOut    = errors.New("job request timed out")
)

type Input struct {
	Name string
	Data interface{}
}

type Output struct {
	Err  error
	Data interface{}
}

type Worker interface {
	// Process will synchronously perform a job and return the result.
	GetLogic(name string) (Logic,error)

	// BlockUntilReady is called before each job is processed and must block the
	// calling goroutine until the Worker is ready to process the next job.
	BlockUntilReady()

	// Interrupt is called when a job is cancelled. The worker is responsible
	// for unblocking the Process implementation.
	Interrupt()

	// Terminate is called when a Worker is removed from the processing pool
	// and is responsible for cleaning up any held resources.
	Terminate()
}

type Pool struct {
	ctor    func() Worker
	workers []*workerWrapper
	reqChan chan workRequest

	poolType PoolType

	workerMut  sync.Mutex
	queuedJobs int64
}

// New creates a new Pool of workers that starts with n workers. You must
// provide a constructor function that creates new Worker types and when you
// change the size of the pool the constructor will be called to create each new
// Worker.
func New(poolType PoolType, concurrency int, builder func() Worker) (*Pool, error) {
	pool := &Pool{
		ctor:     builder,
		reqChan:  make(chan workRequest),
		poolType: poolType,
	}
	pool.SetSize(concurrency)

	return pool, nil
}

//------------------------------------------------------------------------------

// Process will use the Pool to process a payload and synchronously return the
// result. Process can be called safely by any goroutines, but will panic if the
// Pool has been stopped.
func (p *Pool) Process(i Input) Output {
	var o Output

	atomic.AddInt64(&p.queuedJobs, 1)

	request, open := <-p.reqChan
	if !open {
		panic(ErrPoolNotRunning)
	}

	request.jobChan <- i

	o, open = <-request.retChan
	if !open {
		panic(ErrWorkerClosed)
	}

	atomic.AddInt64(&p.queuedJobs, -1)
	return o
}

// ProcessTimed will use the Pool to process a payload and synchronously return
// the result. If the timeout occurs before the job has finished the worker will
// be interrupted and ErrJobTimedOut will be returned. ProcessTimed can be
// called safely by any goroutines.
func (p *Pool) ProcessTimed(
	i Input,
	timeout time.Duration,
) (o Output) {
	atomic.AddInt64(&p.queuedJobs, 1)
	defer atomic.AddInt64(&p.queuedJobs, -1)

	tout := time.NewTimer(timeout)

	var request workRequest
	var open bool

	select {
	case request, open = <-p.reqChan:
		if !open {
			return Output{
				Err: ErrPoolNotRunning,
			}
		}
	case <-tout.C:
		return Output{
			Err: ErrJobTimedOut,
		}
	}

	select {
	case request.jobChan <- i:
	case <-tout.C:
		request.interruptFunc()
		return Output{
			Err: ErrJobTimedOut,
		}
	}

	select {
	case o, open = <-request.retChan:
		if !open {

			return Output{
				Err: ErrWorkerClosed,
			}
		}
	case <-tout.C:
		request.interruptFunc()
		return Output{
			Err: ErrWorkerClosed,
		}
	}

	tout.Stop()
	return Output{
		Data: o,
	}
}

// QueueLength returns the current count of pending queued jobs.
func (p *Pool) QueueLength() int64 {
	return atomic.LoadInt64(&p.queuedJobs)
}

// SetSize changes the total number of workers in the Pool. This can be called
// by any goroutine at any time unless the Pool has been stopped, in which case
// a panic will occur.
func (p *Pool) SetSize(n int) {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()

	lWorkers := len(p.workers)
	if lWorkers == n {
		return
	}

	// Add extra workers if N > len(workers)
	for i := lWorkers; i < n; i++ {
		p.workers = append(p.workers, newWorkerWrapper(p.poolType, p.reqChan, p.ctor()))
	}

	// Asynchronously stop all workers > N
	for i := n; i < lWorkers; i++ {
		p.workers[i].stop()
	}

	// Synchronously wait for all workers > N to stop
	for i := n; i < lWorkers; i++ {
		p.workers[i].join()
	}

	// Remove stopped workers from slice
	p.workers = p.workers[:n]
}

// GetSize returns the current size of the pool.
func (p *Pool) GetSize() int {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()

	return len(p.workers)
}

// Close will terminate all workers and close the job channel of this Pool.
func (p *Pool) Close() {
	p.SetSize(0)
	close(p.reqChan)
}
