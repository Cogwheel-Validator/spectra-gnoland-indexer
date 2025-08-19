package accountcache

func NewAccountCache() *AccountCache {
	return &AccountCache{
		address: make(map[string]int32),
	}
}
