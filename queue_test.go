package things

import (
	"errors"
	"fmt"
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
	for i := 1; i < testNum/2; i++ {
		queue.Do(funcs...)
	}

	doAdd := make(chan struct{})
	doneAdd := make(chan struct{})

	queue.Do(func() error {
		doAdd <- struct{}{}
		<-doneAdd
		fmt.Println(atomic.LoadUint64(&num))
		return errTask
	})

	for i := 20; i >= 0; i-- {
		go queue.Run(0)
	}

	<-doAdd
	for i := testNum / 2; i < testNum; i++ {
		queue.Do(funcs...)
	}
	doneAdd <- struct{}{}

	if err := queue.Wait(); err != errTask {
		t.Logf("expected %v but got %v", errTask, err)
		t.Fail()
	}

	queue.SkipErrored()

	_, err := queue.RunQueued(0)
	if err != nil {
		t.Errorf("error running tasks %v", err)
	}

	if num != uint64(testNum*batchSize) {
		t.Logf("not all functions were ran! Queued %v but only ran %v", testNum*batchSize, num)
		t.Fail()
	}
	fmt.Println(num)
	num = 0
}
