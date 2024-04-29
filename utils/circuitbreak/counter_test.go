package circuitbreak

import (
	"sync"
	"testing"
)

func BenchmarkAtomicCounter_Add(b *testing.B) {
	c := atomicCounter{}
	b.SetParallelism(1000)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			c.Add(1)
		}
	})
}

func BenchmarkPerPCounter_Add(b *testing.B) {
	c := newPerPCounter()
	b.SetParallelism(1000)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			c.Add(1)
		}
	})
}

func TestPerPCounter(t *testing.T) {
	numPerG := 1000
	numG := 1000
	c := newPerPCounter()
	c1 := atomicCounter{}
	var wg sync.WaitGroup
	wg.Add(numG)
	for i := 0; i < numG; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < numPerG; i++ {
				c.Add(1)
				c1.Add(1)
			}
		}()
	}
	wg.Wait()
	total := c.Get()
	total1 := c1.Get()
	if total != c1.Get() {
		t.Errorf("expected %d, get %d", total1, total)
	}
	c.Zero()
	c1.Zero()
	if c.Get() != 0 || c1.Get() != 0 {
		t.Errorf("zero failed")
	}
}
