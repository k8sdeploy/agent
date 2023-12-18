package service

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/agent/internal/agent"
	"github.com/k8sdeploy/agent/internal/config"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
)

type Service struct {
	Config *config.Config
}

func (s *Service) LocalStart() error {
	errChan := make(chan error)
	startAgent(s.Config, errChan)
	return <-errChan
}

func (s *Service) Start() error {
	errChan := make(chan error)
	if !s.Config.Config.Local.Development {
		go startHealth(s.Config, errChan)
	}
	go startAgent(s.Config, errChan)

	return <-errChan
}

func startHealth(cfg *config.Config, errChan chan error) {
	p := fmt.Sprintf(":%d", cfg.Local.HTTPPort)
	logs.Local().Infof("Starting agent healthchecks on %s", p)

	r := chi.NewRouter()
	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)

	srv := &http.Server{
		Addr:              p,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		errChan <- err
	}
}

func startAgent(cfg *config.Config, errChan chan error) {
	if err := agent.NewAgent(cfg).Start(); err != nil {
		errChan <- err
	}
}
