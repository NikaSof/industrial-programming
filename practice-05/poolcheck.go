package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Synthetic waiting workload; this is not a CRUD throughput benchmark.
func checkPool(db *sql.DB) error {
	for _, limit := range []int{1, 5, 10} {
		db.SetMaxIdleConns(0)
		db.SetMaxOpenConns(limit)
		db.SetMaxIdleConns(limit)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		before := db.Stats()
		start := time.Now()
		var wg sync.WaitGroup
		errs := make(chan error, 20)
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					if _, err := db.ExecContext(ctx, "SELECT pg_sleep(0.05)"); err != nil {
						errs <- err
						return
					}
				}
			}()
		}
		wg.Wait()
		after := db.Stats()
		cancel()
		close(errs)
		for err := range errs {
			return err
		}
		fmt.Printf("max=%d queries=100 elapsed=%s waits=%d wait_total=%s\n", limit, time.Since(start).Round(time.Millisecond), after.WaitCount-before.WaitCount, (after.WaitDuration - before.WaitDuration).Round(time.Millisecond))
	}
	return nil
}
