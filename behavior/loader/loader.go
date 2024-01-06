package loader

import (
	"github.com/yamakiller/velcro-go/behavior/actions"
	"github.com/yamakiller/velcro-go/behavior/composites"
	"github.com/yamakiller/velcro-go/behavior/condition"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
	"github.com/yamakiller/velcro-go/behavior/decorators"
	"github.com/yamakiller/velcro-go/behavior/registers"
)

func init() {
	// composites
	registers.Register("Selector", &composites.Selector{})
	registers.Register("RandomSelector", &composites.RandomSelector{})
	registers.Register("Sequenece", &composites.Sequenece{})
	registers.Register("RandomSequence", &composites.RandomSequence{})
	registers.Register("ParallelSequence", &composites.ParallelSequence{})
	registers.Register("ParallelSelector", &composites.ParallelSelector{})
	registers.Register("ifelse", &composites.Ifels{})
	// decorator
	registers.Register("Repeater", &decorators.Repeater{})
	registers.Register("Inverter", &decorators.Inverter{})
	registers.Register("Succeeder", &decorators.Succeeder{})
	registers.Register("UntilFailure", &decorators.UntilFailure{})
	registers.Register("UntilSuccess", &decorators.UntilSuccess{})
	registers.Register("AlwaysFailed", &decorators.AlwaysFailed{})
	registers.Register("AlwaysSuccess", &decorators.AlwaysSuccess{})
	// action
	registers.Register("Wait", &actions.Wait{})
	registers.Register("WaitRandom", &actions.WaitRandom{})
	registers.Register("WaitFailure", &actions.WaitFailure{})
	// condition
	registers.Register("TickTimes", &condition.TickTimes{})
}

func LoadBehaviorTreeFromFile(path string) (*core.BehaviorManager, error) {
	bh3data, err := datas.LoadB3File(path)
	if err != nil {
		return nil, err
	}

	mgr, err := core.NewBehaviorManager(&bh3data.Data)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}
