package teeny

type simpleLogic struct {
	name      string
	processor func(interface{}) Output
}

func (c *simpleLogic) SetProcessor(f func(interface{}) Output) {
	c.processor = f
}
func (c *simpleLogic) SetName(n string) {
	c.name = n
}

func (c *simpleLogic) Name() string {
	return c.name
}

func (c *simpleLogic) Process(input Input) Output {
	return c.processor(input.Data)
}

func (c *simpleLogic) GetBlockUntilReadyFunc() func() {
	return nil
}

func (c *simpleLogic) GetInterruptFunc() func() {
	return nil
}

func (c *simpleLogic) GetTerminateFunc() func() {
	return nil
}

func NewSimplePool(n int, logicName string, f func(interface{}) Output) *Pool {
	l := &simpleLogic{
	}

	l.SetName(logicName)

	l.SetProcessor(f)

	b := NewWorkerBuilderBuilder()
	b.Add(l)

	p, _ := New(PoolTypeGoroutine, n, b.GetBuilder())
	return p
}
