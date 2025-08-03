package database

import (
	"sync"
	"time"
)

// MemoryLimiter implementa a interface Limiter usando um armazenamento em memória.
type MemoryLimiter struct {
	mu   sync.Mutex
	data map[string]*RateLimitData
}

// RateLimitData armazena todas as informações para uma chave específica
type RateLimitData struct {
	Count         int
	LastRequestAt time.Time
	BlockedUntil  time.Time
}

// NewMemoryLimiter cria uma nova instância de MemoryLimiter e inicia a rotina de limpeza.
func NewMemoryLimiter() *MemoryLimiter {
	limiter := &MemoryLimiter{
		data: make(map[string]*RateLimitData),
	}

	go limiter.cleanup(5 * time.Minute)

	return limiter
}

// IsAllowed é a função principal que determina se uma requisição é permitida.
func (l *MemoryLimiter) IsAllowed(key string, limit int, expTime time.Duration, blockTime time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	rateData, ok := l.data[key]
	if !ok {
		rateData = &RateLimitData{
			Count:         1,
			LastRequestAt: time.Now(),
		}
		l.data[key] = rateData
		return true
	}

	if !rateData.BlockedUntil.IsZero() {
		if time.Now().Before(rateData.BlockedUntil) {
			return false
		}
		rateData.BlockedUntil = time.Time{}
		rateData.Count = 0
	}

	if time.Now().After(rateData.LastRequestAt.Add(expTime)) {
		rateData.Count = 0
	}

	rateData.Count++
	rateData.LastRequestAt = time.Now()

	if rateData.Count > limit {
		rateData.BlockedUntil = time.Now().Add(blockTime)
		return false
	}

	return true
}

// cleanup remove chaves que não foram usadas por um período prolongado.
func (l *MemoryLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()

		var keysToRemove []string
		inactiveThreshold := interval * 3

		for key, data := range l.data {
			if time.Since(data.LastRequestAt) > inactiveThreshold {
				keysToRemove = append(keysToRemove, key)
			}
		}

		for _, key := range keysToRemove {
			delete(l.data, key)
		}

		l.mu.Unlock()
	}
}
