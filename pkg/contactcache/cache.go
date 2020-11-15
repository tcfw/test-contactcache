package contactcache

import "time"

//Cacher interfaces to a caching server such as redis/memcached etc.
type Cacher interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Delete(key string) error
}
