package transport

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

type HTTPServer struct {
	Mux         *http.ServeMux
	Server      *http.Server
	Description string
}

func (h *HTTPServer) Listen(ctx context.Context, wg *sync.WaitGroup) {
	wg.Go(func() {
		log.Printf("started http server for %s on %s", h.Description, h.Server.Addr)
		if err := h.Server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Printf("error occured during listening %s", err.Error())
			}
			log.Printf("server %s closed", h.Server.Addr)
		}
	})
	go func() {
		<-ctx.Done()
		var (
			lctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		)
		defer cancel()

		if err := h.Server.Shutdown(lctx); err != nil {
			log.Printf("error occured during server shutdown %s", err.Error())
		}

	}()
}
