package sample1

import (
	"fmt"
	"sync"
	"time"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	time               time.Time
	prices             map[string]float64
}

func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
		time:               time.Now(),
		prices:             map[string]float64{},
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	getService := true
	price, ok := c.prices[itemCode]
	if ok {
		maxAge := c.maxAge
		maxtimecache := c.time.Add(maxAge)
		getService = maxtimecache.Before(time.Now())
	}
	if getService {
		price, err := c.actualPriceService.GetPriceFor(itemCode)
		if err != nil {
			return 0, fmt.Errorf("getting price from service : %v", err.Error())
		}
		c.prices[itemCode] = price
		return price, nil
	}
	return price, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {
	results := make([]float64, len(itemCodes))
	var wg sync.WaitGroup
	wg.Add(len(itemCodes))
	for i, itemCode := range itemCodes {
		go func(i int, itemCode string) {
			defer wg.Done()
			price, err := c.GetPriceFor(itemCode)
			if err != nil {
			}
			results[i] = price
		}(i, itemCode)
	}
	wg.Wait()
	return results, nil
}
