package youtube

import (
	"time"
)

const defaultCacheExpiration = time.Minute * time.Duration(5)

type playerCache struct {
	key       string
	expiredAt time.Time
	config    playerConfig
}

// Get : get cache  when it has same video id and not expired
func (s playerCache) Get(key string) playerConfig {
	return s.GetCacheBefore(key, time.Now())
}

// GetCacheBefore : can pass time for testing
func (s playerCache) GetCacheBefore(key string, time time.Time) playerConfig {
	if key == s.key && s.expiredAt.After(time) {
		return s.config
	}
	return nil
}

// Set : set cache with default expiration
func (s *playerCache) Set(key string, operations playerConfig) {
	s.setWithExpiredTime(key, operations, time.Now().Add(defaultCacheExpiration))
}

func (s *playerCache) setWithExpiredTime(key string, config playerConfig, time time.Time) {
	s.key = key
	s.config = config
	s.expiredAt = time
}
