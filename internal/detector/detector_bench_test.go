package detector

import (
	"testing"
)

// измеряет скорость проверки
func BenchmarkAnomalyDetector_IsAnomaly(b *testing.B) {
	d := NewAnomalyDetector(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.IsAnomaly("iphone", "user1")
	}
}

// измеряет конкурентную проверку
func BenchmarkAnomalyDetector_IsAnomalyParallel(b *testing.B) {
	d := NewAnomalyDetector(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d.IsAnomaly("iphone", "user1")
		}
	})
}

// измеряет проверку с разными ключами
func BenchmarkAnomalyDetector_DifferentKeys(b *testing.B) {
	d := NewAnomalyDetector(100)
	queries := []string{"iphone", "samsung", "xiaomi", "poco", "google"}
	users := []string{"user1", "user2", "user3", "user4", "user5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		u := users[i%len(users)]
		d.IsAnomaly(q, u)
	}
}

// измеряет нагрузку с burst
func BenchmarkAnomalyDetector_HighLoad(b *testing.B) {
	d := NewAnomalyDetectorWithBurst(100, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.IsAnomaly("iphone", "user1")
	}
}
