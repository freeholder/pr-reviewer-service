package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/freeholder/pr-reviewer-service/internal/config"
	"github.com/freeholder/pr-reviewer-service/internal/logging"
	"github.com/freeholder/pr-reviewer-service/internal/migrate"
	"github.com/freeholder/pr-reviewer-service/internal/random"
	"github.com/freeholder/pr-reviewer-service/internal/repository/postgres"
	"github.com/freeholder/pr-reviewer-service/internal/service"
	httptransport "github.com/freeholder/pr-reviewer-service/internal/transport/http"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.NewLogger()

	ctx := context.Background()

	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		logger.Error("open db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(ctx); err != nil {
		logger.Error("ping db", "err", err)
		os.Exit(1)
	}

	if err := migrate.Up(db, "./migrations"); err != nil {
		logger.Error("migrate up", "err", err)
		os.Exit(1)
	}

	teamRepo := postgres.NewTeamRepo(db)
	userRepo := postgres.NewUserRepo(db)
	prRepo := postgres.NewPRRepo(db)
	statsRepo := postgres.NewStatsRepo(db)

	teamSvc := service.NewTeamService(logger, teamRepo, userRepo)
	userSvc := service.NewUserService(logger, userRepo, prRepo)
	prSvc := service.NewPRService(logger, userRepo, prRepo, random.DefaultRandomizer{})
	statsSvc := service.NewStatsService(statsRepo)

	handler := httptransport.NewHandler(logger, teamSvc, userSvc, prSvc, statsSvc)
	router := httptransport.NewRouter(handler)

	addr := ":" + cfg.HTTPPort
	logger.Info("starting http server", slog.String("addr", addr))

	if err := http.ListenAndServe(addr, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("http server error", "err", err)
		os.Exit(1)
	}
}
