package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// App represents the server's internal state.
// It holds configuration about providers and content
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

type pagination struct {
	count  int
	offset int
}

const MAX_POOL_WORKERS = 10

var (
	ErrInputParams = errors.New("bad count/order")
)

func (app App) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	count, offset, err := requestParams(req)
	if err != nil {
		//TODO: send 400 status
		return
	}

	// in some cases we might need a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// to control qty of goroutines per request lets start the pool of workers
	poolSize := MAX_POOL_WORKERS
	if poolSize > count {
		poolSize = count
	}
	pool := StartPool(ctx, app.worker, poolSize)

	// recalculate offset into index to start results from
	indexToStart := 0
	if offset > 0 {
		indexToStart = offset % len(app.Config)
	}

	// run jobs to fetch results frm providers
	go func(indexToStart int) {
		jobsStartedQTY := 0
		for {
			for i, cfg := range app.Config {
				if i >= indexToStart {
					pool.StartJob(
						PoolJob{
							ProviderCFG: cfg,
							CountItems:  1,
							IP:          getIP(req),
							JobNum:      jobsStartedQTY,
						},
					)
					jobsStartedQTY++
					if jobsStartedQTY >= count {
						return
					}
				}
			}
			indexToStart = 0
		}
	}(indexToStart)

	// fetch results from pool's results channel into map to restore order later
	//
	dataMap := map[int][]*ContentItem{}
	countResults := 0
	errJobNum := count
	for res := range pool.Results() {
		if res.Err == nil && res.JobNum < errJobNum { //populate results into map for jobs w/o err and jobNum lower then, skip "after-error" items
			dataMap[res.JobNum] = res.Data
		} else {
			if res.JobNum < errJobNum {
				errJobNum = res.JobNum //remember lowest jobnum with error
			}
		}
		countResults += res.CountItemsRequested
		if countResults >= count {
			pool.Stop()
		}
	}

	// restore order of items
	data := []*ContentItem{}
	for i := 0; i < errJobNum; i++ {
		data = append(data, dataMap[i]...)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

//
func (app App) worker(providerCFG ContentConfig, ip string, count int) ([]*ContentItem, error) {
	provType := providerCFG.Type
	client := app.ContentClients[provType]

	data, err := client.GetContent(ip, count)
	if err != nil {
		if providerCFG.Fallback != nil {
			provType = *providerCFG.Fallback
			client = app.ContentClients[provType]
			return client.GetContent(ip, count)
		}
	}

	return data, err
}

// helper func

func requestParams(req *http.Request) (count int, offset int, err error) {
	count, err = strconv.Atoi(req.URL.Query().Get("count"))
	if err != nil {
		return 0, 0, fmt.Errorf("%w:%v", ErrInputParams, err)
	}
	if count < 1 {
		return 0, 0, ErrInputParams
	}
	offset, err = strconv.Atoi(req.URL.Query().Get("offset"))
	if err != nil {
		return 0, 0, fmt.Errorf("%w:%v", ErrInputParams, err)
	}
	if offset < 0 {
		return 0, 0, ErrInputParams
	}

	return count, offset, nil
}

func getIP(req *http.Request) string {
	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		return req.RemoteAddr
	}
	return ip
}
