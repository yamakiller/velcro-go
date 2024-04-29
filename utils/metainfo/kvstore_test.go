package metainfo

import (
	"fmt"
	"testing"
)

func TestKVStore(t *testing.T) {
	store := newKVStore()
	store["a"] = "a"
	store["a"] = "b"
	if store["a"] != "b" {
		t.Fatal()
	}
	store.recycle()
	if store["a"] == "b" {
		t.Fatal()
	}
	store = newKVStore()
	if store["a"] == "b" {
		t.Fatal()
	}
}

func BenchmarkMap(b *testing.B) {
	for keys := 1; keys <= 1000; keys *= 10 {
		b.Run(fmt.Sprintf("keys=%d", keys), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m := make(map[string]string)
				for idx := 0; idx < 1000; idx++ {
					m[fmt.Sprintf("key-%d", idx)] = string('a' + byte(idx%26))
				}
			}
		})
	}
}

func BenchmarkKVStore(b *testing.B) {
	for keys := 1; keys <= 1000; keys *= 10 {
		b.Run(fmt.Sprintf("keys=%d", keys), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m := newKVStore()
				for idx := 0; idx < 1000; idx++ {
					m[fmt.Sprintf("key-%d", idx)] = string('a' + byte(idx%26))
				}
				m.recycle()
			}
		})
	}
}
