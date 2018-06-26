package closure

import (
	"fmt"

	"github.com/biolee/teeny"
)

const closureLogicName = "simple closure Logic"

type Logic struct {
}

func (c *Logic) Name() string {
	return closureLogicName
}

func (c *Logic) Process(input teeny.Input) teeny.Output {
	closureFunc, ok := input.Data.(func() teeny.Output)
	if !ok {
		return teeny.Output{
			Err: fmt.Errorf("Logic %v get unintend type\n", c.Name()),
		}
	}
	return closureFunc()

}

func (c *Logic) GetBlockUntilReadyFunc() func() {
	return nil
}

func (c *Logic) GetInterruptFunc() func() {
	return nil
}

func (c *Logic) GetTerminateFunc() func() {
	return nil
}
