package core

import (
	"github.com/yamakiller/velcro-go/behavior/datas"
)

func NewBehaviorManager(project *datas.Behavior3Project) (*BehaviorManager, error) {
	mgr := &BehaviorManager{
		selectedTree: project.SelectedTree,
		treelist: make(map[string]*BehaviorTree, len(project.Trees)),
		arrTree:  make([]BehaviorTree, len(project.Trees)),
	}

	for _, tree := range project.Trees {
		b3t := BehaviorTree{}
		b3t.Initialize()
		b3t.Load(tree)
		mgr.arrTree = append(mgr.arrTree, b3t)
		mgr.treelist[b3t.GetTitile()] = &mgr.arrTree[len(mgr.arrTree)-1]
	}

	return mgr, nil
}

type BehaviorManager struct {
	selectedTree string
	treelist map[string]*BehaviorTree
	arrTree  []BehaviorTree
}

func (mgr *BehaviorManager) SelectBehaviorTree(name string) *BehaviorTree {
	return mgr.treelist[name]
}

func (mgr *BehaviorManager) SelectedBehaviorDefaultTree()*BehaviorTree{
	return mgr.treelist[mgr.selectedTree]
}