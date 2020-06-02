package lib

type Cache struct {
	Slab    map[string]bool
	List    []string
	Limit   int
	Counter int
}

func GetNewCache(limit int) Cache {
	return Cache{Slab: make(map[string]bool), Limit: limit}
}

func (c *Cache) InCache(key string) bool {
	_, ok := c.Slab[key]
	return ok
}
func (c *Cache) Reset() {
	c.Slab = make(map[string]bool)
	c.List = c.List[:0]
	c.Counter = 0
}
func (c *Cache) StoreCache(key string) {
	c.Slab[key] = true
	c.List = append(c.List, key)
	if c.Counter >= c.Limit {
		delete(c.Slab, c.List[0])
		c.List = c.List[1:]
		c.Counter--
	}
	c.Counter++
}
