package store

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/internal/store/batch"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
	"time"
)

type PostgresConfig struct {
	Host           string `config:"host"`
	Port           string `config:"port"`
	User           string `config:"user"`
	Password       string `config:"password"`
	DBName         string `config:"db_name"`
	MaxConnections int    `config:"max_connections"`
}

type postgresImpl struct {
	db           *gorm.DB
	config       PostgresConfig
	sequences    Sequence
	batchHandler batch.Handler
	timeProvider TimeProvider
}

func NewPostgres(config PostgresConfig, sequence Sequence, batchHandlerProvider func(writer batch.Writer) batch.Handler, timeProvider TimeProvider, traceProvider trace.TracerProvider) (Message, error) {
	p := &postgresImpl{
		config:       config,
		sequences:    sequence,
		timeProvider: timeProvider,
	}
	p.batchHandler = batchHandlerProvider(p.saveBatch)

	if err := p.initDB(); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	p.db = db
	if err != nil {
		return nil, err
	}

	if err := p.db.AutoMigrate(&postgresMessage{}); err != nil {
		return nil, err
	}

	sqlDb, err := p.db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxIdleConns(config.MaxConnections)
	sqlDb.SetMaxOpenConns(config.MaxConnections)

	if err := p.db.Use(otelgorm.NewPlugin(otelgorm.WithTracerProvider(traceProvider),
		otelgorm.WithDBName(config.DBName))); err != nil {
		return nil, err
	}

	if err := p.loadSequences(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *postgresImpl) initDB() error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s",
		p.config.Host, p.config.User, p.config.Password, p.config.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	defer func() {
		if sqlDb, err := db.DB(); err != nil {
			sqlDb.Close()
		}
	}()
	if err != nil {
		return err
	}

	result := db.Exec(fmt.Sprintf("CREATE DATABASE %s ;", p.config.DBName))
	if result.Error != nil && !strings.Contains(result.Error.Error(), "exists") {
		return result.Error
	}

	return nil
}

func (p *postgresImpl) loadSequences() error {
	ctx := context.Background()
	rows, err := p.db.WithContext(ctx).Model(&postgresMessage{}).
		Select("subject, max(id) as last_id").Group("subject").Rows()
	defer rows.Close()
	if err != nil {
		return err
	}
	for rows.Next() {
		var subject string
		var lastId int32
		if err := rows.Scan(&subject, &lastId); err != nil {
			return err
		}
		if err := p.sequences.Load(ctx, subject, lastId); err != nil {
			return err
		}
	}

	return nil
}

func (p *postgresImpl) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	return p.batchHandler.AddAndWait(ctx, subject, message)
}

func (p *postgresImpl) GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error) {
	msg := postgresMessage{
		Subject: subject,
		Id:      int32(id),
	}
	p.db.WithContext(ctx).Take(&msg)

	err := p.db.Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrInvalidId
		}

		return nil, err
	}

	if p.timeProvider.GetCurrentTime().Sub(msg.CreatedAt).Seconds() > msg.ExpirationSeconds {
		return nil, ErrExpired
	}

	message := broker.Message{
		Id:         id,
		Body:       msg.Body,
		Expiration: time.Second * time.Duration(msg.ExpirationSeconds),
	}

	return &message, nil
}

func (p *postgresImpl) saveBatch(ctx context.Context, values []*batch.Item) error {
	messages := make([]postgresMessage, len(values))
	for i := range messages {
		newId, err := p.sequences.CreateNewId(ctx, values[i].Subject)
		if err != nil {
			return err
		}
		messages[i].Id = newId
		messages[i].Subject = values[i].Subject
		messages[i].Body = values[i].Message.Body
		messages[i].ExpirationSeconds = values[i].Message.Expiration.Seconds()
	}

	err := p.db.WithContext(ctx).CreateInBatches(messages, len(messages)).Error
	if err != nil {
		return err
	}

	for i, message := range messages {
		values[i].Message.Id = int(message.Id)
	}

	return nil
}

type postgresMessage struct {
	Subject           string `gorm:"primaryKey"`
	Id                int32  `gorm:"primaryKey;autoIncrement:false"`
	Body              string
	ExpirationSeconds float64
	CreatedAt         time.Time
}

func (p *postgresMessage) TableName() string {
	return "messages"
}
