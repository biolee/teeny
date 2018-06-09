package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/biolee/teeny"
)

const ClosureLogicName = "simple closure Logic"

type ClosureLogic struct {
	sync.Mutex

	num int
	f   func(interface{}) interface{}
}

func (c *ClosureLogic) Name() string {
	return ClosureLogicName
}

func (c *ClosureLogic) Process(input teeny.Input) teeny.Output {
	switch input.Data.(type) {
	case string:
		log.Print(input.Data)
		return teeny.Output{}
	default:
		return teeny.Output{
			Err: fmt.Errorf("Logic %v get unintend type\n", c.Name()),
		}
	}
}

func (c *ClosureLogic) GetBlockUntilReadyFunc() func() {
	return func() {
		c.Lock()
		defer c.Unlock()

		c.num++

		log.Println(c.num)
	}
}

func (c *ClosureLogic) GetInterruptFunc() func() {
	return nil
}

func (c *ClosureLogic) GetTerminateFunc() func() {
	return nil
}
