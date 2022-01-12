package main

import (
	"context"
	"log"
)

type Pool struct {
	jobChan     chan PoolJob
	results     chan PoolJobResult
	errs        chan error
	closeSignal chan struct{}
	workerFunc  workerFunc
}

type PoolJob struct {
	ProviderCFG ContentConfig
	CountItems  int
	IP          string
	JobNum      int
}

type PoolJobResult struct {
	Data                []*ContentItem
	JobNum              int
	Err                 error
	CountItemsRequested int
}

type workerFunc func(providerCFG ContentConfig, ip string, count int) ([]*ContentItem, error)

// StartPool func
func StartPool(ctx context.Context, worker workerFunc, size int) *Pool {

	results := make(chan PoolJobResult)
	jobChan := make(chan PoolJob, size)
	closeChan := make(chan struct{})

	pool := &Pool{
		jobChan:     jobChan,
		results:     results,
		workerFunc:  worker,
		closeSignal: closeChan,
	}

	for i := 0; i < size; i++ {
		go pool.worker(ctx, i)
	}
	return pool
}

// StartJob func
func (p *Pool) StartJob(job PoolJob) {
	p.jobChan <- job
}

// Stop func
func (p *Pool) Stop() {
	close(p.closeSignal)
	close(p.results)
}

// Results func
func (p *Pool) Results() <-chan PoolJobResult {
	return p.results
}

func (p *Pool) worker(ctx context.Context, id int) {
	log.Println("[INFO] pool worker", id, "started")
	for {
		select {
		case job := <-p.jobChan:
			data, err := p.workerFunc(job.ProviderCFG, job.IP, job.CountItems)
			p.results <- PoolJobResult{
				Data:                data,
				JobNum:              job.JobNum,
				CountItemsRequested: job.CountItems,
				Err:                 err,
			}
		case <-ctx.Done():
			log.Println("[INFO] Stop worker with context", id)
			return
		case <-p.closeSignal:
			log.Println("[INFO] Stop worker", id)
			return
		}
	}
}
