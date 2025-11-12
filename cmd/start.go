package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"172.21.5.249/airtrans/at-flight-authorization/internal/config"
	"172.21.5.249/airtrans/at-flight-authorization/internal/gapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/router"
	"172.21.5.249/airtrans/at-flight-authorization/internal/postgres"
	"172.21.5.249/airtrans/at-flight-authorization/internal/service"
	"172.21.5.249/airtrans/at-flight-authorization/internal/stream"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const probeFlag string = "probe"

var serverCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the server",
	Long:  "Starts server",
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	serverCmd.Flags().BoolP(probeFlag, "p", false, "Probe readiness before startup.")
	rootCmd.AddCommand(serverCmd)
}

func runServer() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("read config error")
		os.Exit(1)
	}

	log.Debug().Msgf("config %v", cfg)

	if cfg.OtherConfig.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05 02-01-2006"})
	}

	tp, err := initTracer(cfg.OtherConfig)
	if err != nil {
		log.Error().Err(err).Msg("Init tracer error")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error().Msgf("Error shutting down tracer provider: %v", err)
		}
	}()

	db, err := postgres.NewConnection(&postgres.Config{
		Host:     cfg.PostgresConfig.Host,
		Port:     cfg.PostgresConfig.Port,
		Password: cfg.PostgresConfig.Password,
		User:     cfg.PostgresConfig.User,
		DBName:   cfg.PostgresConfig.DBName,
		SSLMode:  cfg.PostgresConfig.SSLMode,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to postgresql")
		os.Exit(1)
	}

	nats, err := stream.StartEmbeddedServer(cfg.NatsConfig.NodeName, cfg.NatsConfig.BindAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize nats server")
		os.Exit(1)
	}

	defer nats.Client.Close()

	svc := service.New(nats, cfg, db)

	errs := make(chan error, 2)

	log.Info().Msg("start http server")
	httpServer := hapi.NewServer(svc, cfg)
	httpServer.InitI18n()
	router.Init(httpServer)
	go httpServer.Start(errs)

	log.Info().Msg("start grpc server")
	grpcServer := gapi.NewServer(cfg, svc)
	go grpcServer.Start(errs)

	svc.SubscribeJS(context.Background())

	ticker := time.NewTicker(cfg.PostgresConfig.CleanupInterval)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := svc.CleanupOldRecords(ctx, cfg.PostgresConfig.Retention); err != nil {
					log.Error().Msgf("Cleanup error: %v", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	log.Fatal().Err(err).Msg("Services terminate")
}
