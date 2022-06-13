package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andrei-cloud/gophermart/internal/repo"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

type worker struct {
	addr string
	db   repo.Repository
}

func NewWorker(addr string, db repo.Repository) *worker {
	return &worker{addr: addr, db: db}
}

func (w *worker) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	ordersChan := make(chan string)
	done := make(chan struct{})
	defer close(ordersChan)

	go w.Process(ctx, ordersChan)

	go func() {
		defer func() {
			ticker.Stop()
			close(done)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.GetJob(ordersChan)
			}
		}
	}()

	<-done
}

func (w *worker) GetJob(ch chan<- string) {
	orders, err := w.db.OrderToProcess()
	if err != nil {
		log.Error().AnErr("OrderToProcess", err).Msg("GetJob")
		return
	}
	//log.Debug().Msgf("GetJob: got %d order for processing", len(orders))
	for _, order := range orders {
		ch <- order
	}
}

func (w *worker) Process(ctx context.Context, ch <-chan string) {
	body := struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual"`
	}{}

	client := resty.New().SetTimeout(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case number := <-ch:
			log.Debug().Msgf("Process: got order %s for processing", number)
			res, err := client.R().Get(fmt.Sprintf("http://%s/api/orders/%s", w.addr, number))
			if err != nil {
				log.Debug().Msgf("Process: resty get %v", err.Error())
			}
			if res.StatusCode() == 200 {
				err := json.Unmarshal(res.Body(), &body)
				if err != nil {
					log.Debug().Msgf("Process: unmarshal %v", err.Error())
				}
				log.Debug().Msgf("Process: parsed %+v", body)
				err = w.db.OrderUpdate(body.Order, repo.OrderStatus(body.Status), body.Accrual)
				if err != nil {
					log.Debug().Msgf("Process: OrderUpdate %v", err.Error())
				}
			}
		}
	}
}
