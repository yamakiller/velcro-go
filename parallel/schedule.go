package parallel

//Scheduler 调度者
type Scheduler func(func([]interface{}))
