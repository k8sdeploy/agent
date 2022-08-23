package service

import (
	"fmt"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugMiddleware "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/k8sdeploy/agent/internal/agent"
	"github.com/k8sdeploy/agent/internal/config"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"net/http"
)

type Service struct {
	Config *config.Config
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go startHealth(s.Config, errChan)
	go startAgent(s.Config, errChan)

	return <-errChan
}

func startHealth(cfg *config.Config, errChan chan error) {
	p := fmt.Sprintf(":%d", cfg.Local.HTTPPort)
	bugLog.Local().Infof("Starting agent on %s", p)

	r := chi.NewRouter()
	r.Use(bugMiddleware.BugFixes)
	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)
	if err := http.ListenAndServe(p, r); err != nil {
		errChan <- err
	}
}

func startAgent(cfg *config.Config, errChan chan error) {
	if err := agent.NewAgent(cfg).Start(); err != nil {
		errChan <- err
	}
}
