package https

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

type HTTPSServer struct {
	httpsImpl   *http.Server
	serveMux    *http.ServeMux
	tlsCertPath string
	tlsKeyPath  string
	started     bool
	startedAt   time.Time
	shutdownSem *sync.WaitGroup
}

func (s *HTTPSServer) Handle(pattern string, handler http.Handler) {
	s.serveMux.Handle(pattern, handler)
}

func (s *HTTPSServer) ListenAndServeAsync(errorChannel chan error) {
	if !s.started {
		s.shutdownSem.Add(1)
		go func() {
			shutdownError := s.httpsImpl.ListenAndServeTLS(s.tlsCertPath, s.tlsKeyPath)
			s.shutdownSem.Done()
			if !errors.Is(shutdownError, http.ErrServerClosed) {
				// TODO: Log "Server exited unexpectedly: %v", shutdownError
				errorChannel <- shutdownError
			}
		}()

		s.startedAt = time.Now().UTC()
		s.started = true
		// TODO: Log "Server started listening for connections on %s", s.httpsImpl.Addr
	}
}

func (s *HTTPSServer) Shutdown() error {
	if s.started {
		// TODO: Log "Server interrupted, shutting down..."
		shutdownErr := s.httpsImpl.Shutdown(context.TODO())
		s.shutdownSem.Wait()
		if shutdownErr != nil {
			// TODO: Return "Error encountered during server shutdown: %v", shutdownErr
		}
		// TODO: Log "Server stopped"
	}
	return nil
}
