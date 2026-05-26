package stoplist

import (
	"sync"
	"testing"
)

func TestStopList_AddAndContains(t *testing.T) {
	sl := NewStopList()

	sl.Add("SPAM")
	sl.Add("  Advert  ")

	if !sl.Contains("spam") {
		t.Error("Expected case-insensitive match")
	}
	if !sl.Contains("advert") {
		t.Error("Expected trimmed match")
	}
	if sl.Contains("good") {
		t.Error("Should not contain")
	}
}

func TestStopList_Remove(t *testing.T) {
	sl := NewStopList()
	sl.Add("test")

	if !sl.Remove("test") {
		t.Error("Remove should return true")
	}
	if sl.Contains("test") {
		t.Error("Word should be removed")
	}
	if sl.Remove("nonexistent") {
		t.Error("Remove nonexistent should return false")
	}
}

func TestStopList_BatchAdd(t *testing.T) {
	sl := NewStopList()
	added := sl.BatchAdd([]string{"a", "b", "c", ""})

	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}
	if sl.Size() != 3 {
		t.Errorf("Expected size 3, got %d", sl.Size())
	}
}

func TestStopList_Concurrent(t *testing.T) {
	sl := NewStopList()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sl.Add("test")
		}()
	}

	wg.Wait()
	if sl.Size() != 1 {
		t.Errorf("Expected size 1, got %d", sl.Size())
	}
}
