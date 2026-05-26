package stoplist

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

type StopList struct {
	mu    sync.RWMutex
	words map[string]struct{}
}

func NewStopList() *StopList {
	return &StopList{
		words: make(map[string]struct{}),
	}
}

func NewStopListFromFile(filePath string) (*StopList, error) {
	sl := NewStopList()

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return sl, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			sl.Add(word)
		}
	}
	return sl, nil
}

func (sl *StopList) Add(word string) {
	if word == "" {
		return
	}
	sl.mu.Lock()
	defer sl.mu.Unlock()
	word = strings.ToLower(strings.TrimSpace(word))
	sl.words[word] = struct{}{}
}

func (sl *StopList) Remove(word string) bool {
	if word == "" {
		return false
	}
	sl.mu.Lock()
	defer sl.mu.Unlock()

	word = strings.ToLower(strings.TrimSpace(word))
	if _, exists := sl.words[word]; exists {
		delete(sl.words, word)
		return true
	}
	return false
}

func (sl *StopList) Contains(word string) bool {
	if word == "" {
		return false
	}
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	word = strings.ToLower(strings.TrimSpace(word))
	_, exists := sl.words[word]
	return exists
}

func (s *StopList) GetAll() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	words := make([]string, 0, len(s.words))
	for word := range s.words {
		words = append(words, word)
	}
	return words
}

func (s *StopList) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.words)
}

func (s *StopList) SaveToFile(filePath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for word := range s.words {
		if _, err := writer.WriteString(word + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func (s *StopList) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.words = make(map[string]struct{})
}

func (s *StopList) BatchAdd(words []string) int {
	if len(words) == 0 {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	added := 0
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if word != "" {
			if _, exists := s.words[word]; !exists {
				s.words[word] = struct{}{}
				added++
			}
		}
	}
	return added
}
