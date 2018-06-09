package teeny

type Logic interface {
	Name() string
	Process(Input) Output
	GetBlockUntilReadyFunc() func()
	GetInterruptFunc() func()
	GetTerminateFunc() func()
}
