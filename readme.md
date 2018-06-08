> Goroutine or "Thread" pool


# motivation

Inspired by [tunny](https://github.com/Jeffail/tunny)

- add os thread locked goroutine pool
- add process logic for same input type

# usage

1. add logic that implement [`Logic interface`](./logic.go#5) 

```go
b := teeny.NewWorkerBuilderBuilder()

b.Add(&ClosureLogic{})
```

2. build pull 

```go
p, _ := teeny.New(teeny.PoolTypeThread, 100, b.GetBuilder())
```

3. process your data

```go
p.Process(teeny.Input{
	Name: ClosureLogicName,
	Data: "hello word",
})
```
