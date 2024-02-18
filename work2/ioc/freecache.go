package ioc

import "github.com/coocood/freecache"

func InitFreeCache() *freecache.Cache {
	size := 200 * 1024 * 1024
	return freecache.NewCache(size)
}
