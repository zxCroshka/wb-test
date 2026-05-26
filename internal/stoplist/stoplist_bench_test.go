package stoplist

import (
	"testing"
)

// BenchmarkStopList_Contains измеряет скорость проверки слова
func BenchmarkStopList_Contains(b *testing.B) {
	sl := NewStopList()
	sl.Add("testword")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.Contains("testword")
	}
}

// BenchmarkStopList_ContainsParallel измеряет конкурентную проверку
func BenchmarkStopList_ContainsParallel(b *testing.B) {
	sl := NewStopList()
	sl.Add("testword")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sl.Contains("testword")
		}
	})
}

// BenchmarkStopList_Add измеряет скорость добавления
func BenchmarkStopList_Add(b *testing.B) {
	sl := NewStopList()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.Add("newword")
	}
}

// BenchmarkStopList_Remove измеряет скорость удаления
func BenchmarkStopList_Remove(b *testing.B) {
	sl := NewStopList()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.Add("test")
		sl.Remove("test")
	}
}

// BenchmarkStopList_BatchAdd измеряет массовое добавление
func BenchmarkStopList_BatchAdd(b *testing.B) {
	sl := NewStopList()
	words := make([]string, 100)
	for i := 0; i < 100; i++ {
		words[i] = "word"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.BatchAdd(words)
	}
}
