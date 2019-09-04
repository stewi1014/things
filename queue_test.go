package things

import (
	"errors"
	"sync/atomic"
	"testing"
)

var num uint64

func testFuncs(n int) []func() error {
	funcs := make([]func() error, n)
	for i := range funcs {
		funcs[i] = func() error {
			atomic.AddUint64(&num, 1)
			return nil
		}
	}
	return funcs
}

func BenchmarkQueue(b *testing.B) {
	funcs := testFuncs(b.N)
	q := NewQueue(nil)
	go func() {
		q.Run(0)
	}()

	q.Do(funcs...)
	q.Wait()
	q.Cancel()
}

var errTask = errors.New("task error")

func TestQueue(t *testing.T) {
	testNum, batchSize := 5000, 1000
	funcs := testFuncs(batchSize)

	queue := NewQueue(nil)

	queue.Do(funcs...)
	for i := 1; i < testNum/10; i++ {
		queue.Do(funcs...)
	}

	doAdd := make(chan struct{})
	doneAdd := make(chan struct{})

	ec := make(chan error)
	go func() {
		ec <- queue.Err(true)
	}()

	for i := 20; i >= 0; i-- {
		go queue.Run(0)
	}

	queue.Do(func() error {
		close(doAdd)
		<-doneAdd
		return errTask
	})

	<-doAdd
	for i := testNum / 10; i < testNum; i++ {
		queue.Do(funcs...)
	}
	close(doneAdd)

	if err := queue.Wait(); err != errTask {
		t.Errorf("expected %v from Wait() but got %v", errTask, err)
	}

	queue.SkipErrored()

	_, err := queue.RunQueued(0)
	if err != nil {
		t.Errorf("error running tasks %v", err)
	}

	if num != uint64(testNum*batchSize) {
		t.Errorf("not all functions were ran! Queued %v but only ran %v", testNum*batchSize, num)
	}

	if err := <-ec; err != errTask {
		t.Errorf("expected %v from error channel but got %v", errTask, err)
	}

	queue.Reset(nil)
	if err := queue.Err(false); err != nil {
		t.Errorf("reset didn't clear error, got %v", err)
	}

	num = 0
}
