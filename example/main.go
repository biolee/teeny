package main

import (
	"log"

	"github.com/biolee/teeny"
	"github.com/biolee/teeny/closure"
)

func main() {
	b := teeny.NewWorkerBuilderBuilder()
	closureLogic := closure.Logic{}

	b.Add(&closureLogic)

	p, _ := teeny.New(teeny.PoolTypeThread, 100, b.GetBuilder())

	for i := 0; i < 100000; i++ {
		p.Process(teeny.Input{
			Name: closureLogic.Name(),
			Data: func() teeny.Output {
				log.Println("asdfasdfasd")
				return teeny.Output{}
			},
		})
	}

}
