package teeny

import (
	"fmt"
	"log"
	"sync"
)

type LogicMap map[string]Logic

type WorkerBuilder struct {
	sync.RWMutex

	LogicMap LogicMap
}

func NewWorkerBuilderBuilder() *WorkerBuilder {
	b := new(WorkerBuilder)
	b.LogicMap = LogicMap{}

	return b
}

func (b *WorkerBuilder) Add(l Logic) error {
	b.Lock()
	defer b.Unlock()

	b.LogicMap[l.Name()] = l

	return nil
}

func (b *WorkerBuilder) GetBuilder() func() Worker {
	b.RLock()
	defer b.RUnlock()

	return func() Worker {
		return b
	}
}

func (b *WorkerBuilder) BlockUntilReady() {
	for name, f := range b.LogicMap {
		log.Printf("Waiting for Logic: %v", name)
		if f.GetBlockUntilReadyFunc() != nil {
			f.GetBlockUntilReadyFunc()()
		}
	}
}

func (b *WorkerBuilder) Interrupt() {
	for name, f := range b.LogicMap {
		log.Printf("Interruptint Logic: %v", name)
		if f.GetInterruptFunc() != nil {
			f.GetInterruptFunc()()
		}
	}

}
func (b *WorkerBuilder) Terminate() {
	for name, f := range b.LogicMap {
		log.Printf("Terminating Logic: %v", name)
		if f.GetTerminateFunc() != nil {
			f.GetTerminateFunc()()
		}
	}
}

func (b *WorkerBuilder) GetLogic(name string) (Logic, error) {
	b.RLock()
	defer b.RUnlock()

	var err error

	l, ok := b.LogicMap[name]

	if ok != true {
		err = fmt.Errorf("logic %v is not found", name)
	}

	return l, err
}
