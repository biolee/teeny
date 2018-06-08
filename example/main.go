package main

import (
	"github.com/biolee/teeny"
)

func main() {
	b := teeny.NewWorkerBuilderBuilder()

	b.Add(&ClosureLogic{})

	p, _ := teeny.New(teeny.PoolTypeThread, 100, b.GetBuilder())

	for i := 0; i < 100000; i++ {
		p.Process(teeny.Input{
			Name: ClosureLogicName,
			Data: "hello word",
		})
	}

}
