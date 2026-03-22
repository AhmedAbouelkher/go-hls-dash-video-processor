package main

import "sync"

type ConsumerFunc[T any, A any] func(e T) (A, error)
type Result[T any, A any] struct {
	job T
	res A
	err error
}

type Pool[T any, A any] struct {
	Workers  int
	Consumer ConsumerFunc[T, A]
	Data     []T
}

func runWorkersPool[T any, A any](p *Pool[T, A]) []Result[T, A] {
	batch := len(p.Data)
	workers := p.Workers

	var wg sync.WaitGroup
	wg.Add(batch)

	// In order to use our pool of workers we need to send
	// them work and collect their results. We make 2
	// channels for this.
	jobs := make(chan T, batch)
	results := make(chan Result[T, A], batch)

	// This starts up n workers, initially blocked
	// because there are no jobs yet.
	for w := 1; w <= workers; w++ {
		go func(jobs <-chan T, results chan<- Result[T, A], consumer ConsumerFunc[T, A]) {
			for job := range jobs {
				func() {
					defer wg.Done()
					res, err := consumer(job)
					results <- Result[T, A]{job, res, err}
				}()
			}
		}(jobs, results, p.Consumer)
	}

	// Here we send k `jobs` and then `close` that
	// channel to indicate that's all the work we have.
	for _, job := range p.Data {
		jobs <- job
	}
	close(jobs)

	wg.Wait()
	close(results)

	res := make([]Result[T, A], 0)

	// Finally we collect all the results of the work.
	for range batch {
		r := <-results
		res = append(res, r)
	}

	return res
}
