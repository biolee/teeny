package teeny

import (
	"sync"
	"testing"
	"time"
)

func TestPoolSizeAdjustment(t *testing.T) {
	var logicName = "test"
	pool := NewSimplePool(10, logicName, func(interface{}) Output {
		return Output{
			Data: "foo",
		}
	})

	if exp, act := 10, len(pool.workers); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}

	pool.SetSize(10)
	if exp, act := 10, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}

	pool.SetSize(9)
	if exp, act := 9, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}

	pool.SetSize(10)
	if exp, act := 10, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}

	pool.SetSize(0)
	if exp, act := 0, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}

	pool.SetSize(10)
	if exp, act := 10, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}

	// Finally, make sure we still have actual active workers.
	if exp, act := "foo", pool.Process(Input{
		Name: logicName,
		Data: 0,
	}).Data.(string); exp != act {
		t.Errorf("Wrong result: %v != %v", act, exp)
	}

	pool.Close()
	if exp, act := 0, pool.GetSize(); exp != act {
		t.Errorf("Wrong size of pool: %v != %v", act, exp)
	}
}

func TestFuncJob(t *testing.T) {
	var logicName = "test"
	pool := NewSimplePool(10, logicName, func(in interface{}) Output {
		intVal := in.(int)
		return Output{
			Data: intVal * 2,
		}
	})

	defer pool.Close()

	for i := 0; i < 10; i++ {
		ret := pool.Process(Input{
			Data: 10,
			Name: logicName,
		})
		if exp, act := 20, ret.Data.(int); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestFuncJobNotTimeout(t *testing.T) {
	var logicName = "test"
	pool := NewSimplePool(10, logicName, func(in interface{}) Output {
		intVal := in.(int)
		return Output{
			Data: intVal * 2,
		}
	})

	defer pool.Close()

	for i := 0; i < 10; i++ {
		ret := pool.ProcessTimed(Input{
			Data: 10,
			Name: logicName,
		}, time.Millisecond)
		if exp, act := 20, ret.Data.(int); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestFuncJobTimeout(t *testing.T) {
	var logicName = "test"
	pool := NewSimplePool(1, logicName, func(in interface{}) Output {
		intVal := in.(int)
		<-time.After(time.Millisecond)
		return Output{
			Data: intVal * 2,
		}
	})

	defer pool.Close()

	ret := pool.ProcessTimed(Input{
		Data: 10,
		Name: logicName,
	}, time.Millisecond)
	if exp, act := ErrJobTimedOut, ret.Err; exp != act {
		t.Errorf("Wrong result: %v != %v", act, exp)
	}
}

func TestTimedJobsAfterClose(t *testing.T) {
	var logicName = "test"
	pool := NewSimplePool(10, logicName, func(in interface{}) Output {
		intVal := in.(int)
		return Output{
			Data: intVal * 2,
		}
	})
	pool.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Process after Stop() did not panic")
		}
	}()

	pool.Process(Input{
		Data: 10,
		Name: logicName,
	})
}

func TestParallelJobs(t *testing.T) {
	nWorkers := 10

	jobGroup := sync.WaitGroup{}
	testGroup := sync.WaitGroup{}

	var logicName = "test"
	pool := NewSimplePool(10, logicName, func(in interface{}) Output {
		intVal := in.(int)
		return Output{
			Data: intVal * 2,
		}
	})
	defer pool.Close()

	for j := 0; j < 1; j++ {
		jobGroup.Add(nWorkers)
		testGroup.Add(nWorkers)

		for i := 0; i < nWorkers; i++ {
			go func() {
				ret := pool.Process(Input{
					Data: 10,
					Name: logicName,
				})
				if exp, act := 20, ret.Data.(int); exp != act {
					t.Errorf("Wrong result: %v != %v", act, exp)
				}
				testGroup.Done()
			}()
		}

		testGroup.Wait()
	}
}

func BenchmarkFuncJob(b *testing.B) {
	var logicName = "test"
	pool := NewSimplePool(10, logicName, func(in interface{}) Output {
		intVal := in.(int)
		return Output{
			Data: intVal * 2,
		}
	})
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ret := pool.Process(Input{
			Data: 10,
			Name: logicName,
		})
		if exp, act := 20, ret.Data.(int); exp != act {
			b.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func BenchmarkFuncTimedJob(b *testing.B) {
	var logicName = "test"
	pool := NewSimplePool(10, logicName, func(in interface{}) Output {
		intVal := in.(int)
		return Output{
			Data: intVal * 2,
		}
	})
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ret := pool.ProcessTimed(Input{
			Data: 10,
			Name: logicName,
		}, time.Second)
		if ret.Err != nil {
			b.Error(ret.Err)
		}
		if exp, act := 20, ret.Data.(int); exp != act {
			b.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}
