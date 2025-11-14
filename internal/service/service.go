package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"172.21.5.249/airtrans/at-flight-authorization/internal/config"
	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"172.21.5.249/airtrans/at-flight-authorization/internal/stream"
	types "172.21.5.249/airtrans/at-flight-authorization/internal/types/authorization"
	"github.com/nats-io/nats.go"
	"github.com/qiniu/qmgo"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

var rdb *redis.Client

type Service struct {
	DbClient       *qmgo.Client
	nats           *stream.EmbeddedNats
	cfg            config.ServiceConfig
	db             *gorm.DB
	statusInterval time.Duration
	notifier       *Notifier
}

var tracer = otel.Tracer("at-flight-authorization-service")

const JETSTREAM_SUBJECT = "at-flight-authorization.>"
const STREAM_NAME = "at-flight-authorization"

func New(s *stream.EmbeddedNats, cfg config.ServiceConfig, db *gorm.DB) *Service {

	if cfg.RedisConfig.Enable {
		rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisConfig.Address,
			Password: cfg.RedisConfig.Password,
			DB:       cfg.RedisConfig.DB,
		})
	}

	s.Stream.DeleteStream(STREAM_NAME)

	info, err := s.Stream.AddStream(&nats.StreamConfig{
		Name:      STREAM_NAME,
		Subjects:  []string{JETSTREAM_SUBJECT},
		Storage:   nats.MemoryStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    1 * time.Hour,
		MaxMsgs:   10000,
	})

	s.Stream.PurgeStream(STREAM_NAME)

	if err != nil {
		log.Error().Err(err).Msgf("Failed to add stream: %v", err)
	} else {
		log.Debug().Msgf("Added stream: %v", *info)
	}

	// db.Migrator().DropTable(
	// 	//Authorization Form
	// 	&models.Operator{},
	// 	&models.Drone{},
	// 	&models.FlightArea{},
	// 	&models.Pilot{},
	// 	&models.FlightAuthorizationProposal{},
	// 	&models.FlightAreaCoordinate{},
	// 	//Authorization Approval
	// 	&models.FlightParameter{},
	// 	&models.AuthorizedFlightAreaCoordinate{},
	// 	&models.AuthorizedFlightArea{},
	// 	&models.FlightAuthorizationApproval{},
	// 	//Flight Notification
	// 	&models.IntendedFlightArea{},
	// 	&models.IntendedFlightAreaCoordinate{},
	// 	&models.FlightNotification{},
	// )

	//Migrate DB
	db.AutoMigrate(
		//Authorization Form
		&models.Operator{},
		&models.Drone{},
		&models.FlightArea{},
		&models.Pilot{},
		&models.FlightAuthorizationProposal{},
		&models.FlightAreaCoordinate{},
		//Authorization Approval
		&models.FlightParameter{},
		&models.AuthorizedFlightAreaCoordinate{},
		&models.AuthorizedFlightArea{},
		&models.FlightAuthorizationApproval{},
		//Flight Notification
		&models.IntendedFlightArea{},
		&models.IntendedFlightAreaCoordinate{},
		&models.FlightNotification{},
	)

	svc := &Service{
		nats:           s,
		cfg:            cfg,
		db:             db,
		statusInterval: time.Minute,
		notifier:       NewNotifier(s.Client),
	}

	svc.startProposalStatusWatcher()

	return svc
}

func redisSet(key string, value interface{}) error {
	if rdb == nil {
		return errors.New("redis feature not enabled")
	}

	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	rdb.Set(context.Background(), key, val, 0)
	return nil
}

func (s *Service) Notifier() *Notifier {
	return s.notifier
}

func redisGet(key string, dest interface{}) error {
	if rdb == nil {
		return errors.New("redis feature not enabled")
	}

	val, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(val, dest)
}

func redisDelete(keys ...string) {
	if rdb != nil {
		rdb.Del(context.Background(), keys...)
	}
}

func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	temp := strings.Split(fmt.Sprintf("%s", runtime.FuncForPC(pc).Name()), ".")
	return temp[len(temp)-1]
}

func (s *Service) publishEvent(ctx context.Context, subj string, data []byte) {
	js := s.nats.Stream

	ack, err := js.Publish(subj, data)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to publish to jetstream for %v", subj)
	} else {
		log.Debug().Msgf("Published to jetstream for subj %s: %v", subj, *ack)
	}
}

func (s *Service) SubscribeJS(ctx context.Context) {
	js := s.nats.Stream

	sub, err := js.PullSubscribe(
		JETSTREAM_SUBJECT,
		s.cfg.NatsConfig.NodeName,
		nats.PullMaxWaiting(128),
		nats.AckExplicit(),
		nats.BindStream(STREAM_NAME),
	)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create pull consumer for %s", s.cfg.NatsConfig.NodeName)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	go func() {
		for {
			msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || err == nats.ErrTimeout {
					continue
				}
				log.Error().Err(err).Msg("JetStream fetch error")
				continue
			}

			for _, msg := range msgs {
				log.Debug().Msgf("Node %s received msg on subj: %s", s.cfg.NatsConfig.NodeName, msg.Subject)

				if err := s.ProcessNatsMsg(msg); err != nil {
					log.Error().Err(err).Msgf("Error processing nats msg: %v", err)
				}

				if err := msg.Ack(); err != nil {
					log.Error().Err(err).Msgf("Error acknowledging nats msg: %v", err)
				}
			}
		}
	}()
}

func (s *Service) CleanupOldRecords(ctx context.Context, retention time.Duration) error {
	cutoff := time.Now().Add(-retention).Unix()

	cctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result := s.db.WithContext(cctx).
		Where("timestamp < ?", cutoff).
		Delete(&types.AuthorizationLog{})

	if result.Error != nil {
		return result.Error
	}

	log.Info().Msgf("Cleanup completed: %d rows removed (older than %d)", result.RowsAffected, cutoff)

	return nil
}
