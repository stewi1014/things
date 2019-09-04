package things

import (
	"context"
	"sync"
)

const buffSize = 2048 // buffer size for Queue

// NewQueue creates a new queue with the given context.
// Calling new(Queue) or Queue{} is sufficent.
// A nil context is valid.
func NewQueue(ctx context.Context) *Queue {
	q := &Queue{}
	q.Context(ctx)
	return q
}

// Queue is a system for queuing tasks.
// Tasks are executed by calls to Run and RunQueued.
// A Queue cannot be copied.
type Queue struct {
	// Sync
	once  sync.Once
	cond  sync.Cond
	mutex sync.Mutex

	// Runtime
	originalCtx context.Context // Original context.
	ctx         context.Context // Our internal context
	ctxCancel   context.CancelFunc
	exitError   error
	running     int

	// Queue
	queue []func() error
	off   int // queue index for reading
	count int // counter for runners

	// Task after-error recovery.
	recover     [][]func() error
	errCountInd int // recover index of errored runner
	errIndex    int // queue index of errored task
}

// Run executes tasks in the Queue.
// It blocks until n tasks are complete.
// If n is 0 or negative, it will only return on error.
//
// It returns
//	nil if n tasks are complete,
//	the context error if the context finishes,
//	context.Canceled if Cancel is called,
//	or the first error returned by a task executed here.
//
// Subsequent calls to Run after an error will clear the error and attempt to resume execution.
func (q *Queue) Run(n int) error {
	q.mutex.Lock()
	q.init()
	var buff []func() error
	if n <= 0 {
		buff = make([]func() error, buffSize)
	} else {
		buff = make([]func() error, n)
	}

	if q.err() != nil {
		q.resume()
	}

	for {
		for q.len() == 0 && q.ctx.Err() == nil {
			q.cond.Wait()
		}
		if n > 0 && n < len(buff) {
			buff = buff[:n]
		}
		e, err := q.run(buff)
		if err != nil {
			q.mutex.Unlock()
			return err
		}
		if n > 0 {
			n -= e
			if n <= 0 {
				q.mutex.Unlock()
				return nil
			}
		}
	}
}

// RunQueued is the same as Run, except it will not wait for new tasks
// if less than n tasks are queued, however it will execute tasks added during execution.
// If n is 0 or negative, RunQueued returns only when the Queue is empty.
func (q *Queue) RunQueued(n int) (int, error) {
	q.mutex.Lock()

	if q.err() != nil {
		q.resume()
	}

	var buff []func() error
	if n <= 0 {
		if q.len() > buffSize {
			buff = make([]func() error, buffSize)
		} else {
			buff = make([]func() error, q.len())
		}
	} else {
		if n > buffSize {
			buff = make([]func() error, buffSize)
		} else {
			buff = make([]func() error, n)
		}
	}

	var done int

	for q.len() > 0 {
		if n > 0 && n-done < len(buff) {
			buff = buff[:n-done]
		}
		e, err := q.run(buff)
		done += e
		if err != nil {
			q.mutex.Unlock()
			return done, err
		}
		if n > 0 && done >= n {
			q.mutex.Unlock()
			return done, nil
		}
	}

	q.mutex.Unlock()
	return done, nil
}

// mutex must be held
func (q *Queue) run(buff []func() error) (int, error) {
	if err := q.ctx.Err(); err != nil {
		return 0, err
	}

	q.running++
	defer func() {
		q.cond.Broadcast()
	}()

	ctx := q.ctx
	c, index := q.getTasks(buff)

	q.mutex.Unlock()

	var err error
	for i := 0; i < c; i++ {
		if ctx.Err() != nil {
			// Check if the context has just switched.
			q.mutex.Lock()
			ctx = q.ctx
			err = ctx.Err()
			if err != nil {
				q.cancel()
				q.returnTasks(buff[i:c], index)
				q.running--
				if q.running == 0 {
					// We're last
					q.parseRecovered()
				}
				return i, err
			}
			q.mutex.Unlock()
		}
		err = buff[i]()
		if err != nil {
			q.mutex.Lock()
			if q.exitError == nil {
				q.exitError = err
			}
			q.cancel()
			q.failTask(buff[i:c], index)
			q.running--
			if q.running == 0 {
				q.parseRecovered()
			}
			return i, err
		}
	}

	q.mutex.Lock()
	q.running--
	return c, err
}

// Len returns the length of the Queue.
// It does not count currently executing tasks, and might increase without calls to Do as
// Run calls may return unexecuted tasks back to the queue if an error is encountered.
func (q *Queue) Len() int {
	q.mutex.Lock()
	l := q.len()
	q.mutex.Unlock()
	return l
}

// IsIdle returns true if there are no Run calls, or all Run calls are waiting for more tasks.
func (q *Queue) IsIdle() bool {
	q.mutex.Lock()
	if q.running > 0 {
		q.mutex.Unlock()
		return false
	}
	q.mutex.Unlock()
	return true
}

// Do adds tasks to the Queue.
// Sucessful tasks are never executed more than once, are not lost on errors and are always executed in order.
// If execution is halted by an error, all unexecuted tasks including the task that created the error are returned to the queue,
// taking care to respect order. The error can then be handled, and tasks resumed by calling Run again.
// SkipOne is useful for skipping the errored task if needed.
func (q *Queue) Do(f ...func() error) {
	q.mutex.Lock()
	copy(q.queue[q.grow(len(f)):], f)
	q.mutex.Unlock()
	q.cond.Broadcast()
}

// SkipErrored skips the task which produced the error after execution was halted.
// If execution is not halted, it waits until it has.
func (q *Queue) SkipErrored() bool {
	q.mutex.Lock()
	q.waitExit()
	l := q.len()
	if l <= 0 || q.errIndex < 0 {
		q.mutex.Unlock()
		return false
	}

	if q.errIndex > l || q.errIndex <= 0 {
		q.mutex.Unlock()
		return false
	}

	copy(q.queue[q.off+1:q.off+q.errIndex], q.queue[q.off:q.off+q.errIndex-1])
	q.off++
	q.errIndex = 0
	q.mutex.Unlock()
	return true
}

// Cancel stops execution of tasks.
// It does not wait for runners to return, and does not remove any tasks from the queue.
// Calling cancel effectively pauses execution.
func (q *Queue) Cancel() {
	q.mutex.Lock()
	q.cancel()
	q.mutex.Unlock()
}

func (q *Queue) cancel() {
	if q.ctx.Err() == nil {
		q.ctxCancel()
	}
}

// Err returns an error explaining why execution has halted.
// If there is no error, and wait is true, Err will block until an error occours.
// Calling Run again will remove the error and resume execution.
func (q *Queue) Err(wait bool) error {
	q.mutex.Lock()
	q.init()
	err := q.err()
	for ; err == nil && wait; err = q.err() {
		q.cond.Wait()
	}
	q.mutex.Unlock()
	return err
}

// same as Err. mutex must be held.
func (q *Queue) err() error {
	if q.exitError != nil {
		return q.exitError
	}
	return q.ctx.Err()
}

// Reset stops execution of tasks, clears the Queue and sets the Context.
// If a task is executing, it blocks until it returns.
// A nil context is valid.
func (q *Queue) Reset(ctx context.Context) {
	q.mutex.Lock()
	q.cancel()
	q.waitExit()
	q.context(ctx)
	q.exitError = nil
	q.clearQueue()
	q.mutex.Unlock()
}

// Context sets the context for the Queue. It doesn't impact the queue or execution.
// A nil context is valid.
func (q *Queue) Context(ctx context.Context) {
	q.mutex.Lock()
	q.context(ctx)
	q.mutex.Unlock()
}

// Set context
// mutex must be held
func (q *Queue) context(ctx context.Context) {
	q.originalCtx = ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if q.ctxCancel != nil && q.ctx.Err() == nil {
		q.ctxCancel()
	}
	q.ctx, q.ctxCancel = context.WithCancel(ctx)
	go func() {
		<-q.ctx.Done()
		q.cond.Broadcast()
	}()
}

// Wait waits until the Queue is either caught up or errored, returning the error if encountered.
func (q *Queue) Wait() error {
	q.mutex.Lock()
	err := q.wait()
	q.mutex.Unlock()
	return err
}

// mutex must be held
func (q *Queue) wait() error {
	q.init()
	for q.running > 0 || (q.len() > 0 && q.err() == nil) {
		q.cond.Wait()
	}
	return q.err()
}

// wait for runners to exit.
func (q *Queue) waitExit() {
	q.init()
	for q.running > 0 {
		q.cond.Wait()
	}
}

func (q *Queue) resume() {
	q.running--
	q.waitExit()
	q.exitError = nil
	q.context(q.originalCtx)
	q.cond.Broadcast()
}

// Called before using cond.
func (q *Queue) init() {
	q.once.Do(func() {
		q.cond.L = &q.mutex
	})
}

// grows the function buffer by n,
// returning the index where new functions should be written.
// mutex must be held
func (q *Queue) grow(n int) int {
	c, l := cap(q.queue), len(q.queue)
	if c-l >= n {
		q.queue = q.queue[:l+n]
		return l
	}

	if l == 0 {
		q.queue = make([]func() error, n, n*2)
		return 0
	}

	if n <= (c/2)-(l-q.off) {
		u := copy(q.queue, q.queue[q.off:])
		q.queue = q.queue[:u+n]
		q.off = 0
		return u
	}

	nq := make([]func() error, (l-q.off)+n, c*2+n)
	u := copy(nq, q.queue[q.off:])
	q.queue = nq[:u+n]
	q.off = 0
	return u
}

// mutex must be held.
func (q *Queue) clearQueue() {
	q.queue = q.queue[:0]
	q.off = 0
	q.errIndex = 0
	q.recover = q.recover[:0]
}

// unread portion of buffer
func (q *Queue) len() int { return len(q.queue) - q.off }

func (q *Queue) getTasks(buff []func() error) (int, int) {
	c := copy(buff, q.queue[q.off:])
	q.count++
	q.errIndex = 0
	q.off += c
	return c, q.count
}

func (q *Queue) returnTasks(buff []func() error, count int) {
	pos := q.count - count
	if l := len(q.recover); l <= pos {
		nr := make([][]func() error, pos+1)
		copy(nr, q.recover)
		nr[pos] = buff
		q.recover = nr
	} else {
		q.recover[pos] = buff
	}
}

func (q *Queue) failTask(buff []func() error, count int) {
	q.returnTasks(buff, count)
	pos := q.count - count
	q.errCountInd = pos + 1
}

func (q *Queue) parseRecovered() {
	q.errIndex = 0
	for i := range q.recover {
		l := len(q.recover[i])

		copy(q.queue[q.growLeft(l):], q.recover[i])

		if i == q.errCountInd-1 {
			q.errIndex = 1
			q.errCountInd = 0
		}
	}
	q.recover = q.recover[:0]
}

// returns new read offset
func (q *Queue) growLeft(n int) int {
	if q.errIndex > 0 {
		q.errIndex += n
	}

	if q.off >= n {
		q.off -= n
		return q.off
	}

	u, c := q.len(), cap(q.queue)
	var nq []func() error
	if c-u >= n {
		nq = q.queue[:u+n]
	} else {
		nq = make([]func() error, u+n, (c*2)+n)
	}

	copy(nq[n:], q.queue[q.off:])
	q.queue = nq
	q.off = 0
	return q.off
}
