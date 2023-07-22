package store

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"math"
	"strings"
	"time"
)

type PostgresConfig struct {
	Host     string `config:"host"`
	Port     string `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	DBName   string `config:"db_name"`
}

type postgresImpl struct {
	db     *gorm.DB
	tp     trace.TracerProvider
	config PostgresConfig
}

func NewPostgres(config PostgresConfig, traceProvider trace.TracerProvider) (Message, error) {
	p := &postgresImpl{tp: traceProvider, config: config}
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

	if err := p.db.AutoMigrate(&postgresMessage{}, &postgresSequence{}); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *postgresImpl) initDB() error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s",
		p.config.Host, p.config.User, p.config.Password, p.config.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	result := db.Exec(fmt.Sprintf("CREATE DATABASE %s ;", p.config.DBName))
	if result.Error != nil && !strings.Contains(result.Error.Error(), "exists") {
		return result.Error
	}

	return nil
}

func (p *postgresImpl) createNextId(ctx context.Context, subject string) (int32, error) {
	subjectSequence := postgresSequence{
		Subject: subject,
	}
	if err := p.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).Create(&subjectSequence).Error; err != nil && err != gorm.ErrDuplicatedKey {
		return 0, err
	}

	subjectSequence = postgresSequence{
		Subject: subject,
	}
	if err := p.db.WithContext(ctx).Model(&subjectSequence).
		Clauses(clause.Returning{}).Update("val", gorm.Expr("val + 1")).Error; err != nil {
		return 0, err
	}
	return subjectSequence.Val, nil
}

func (p *postgresImpl) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	newId, err := p.createNextId(ctx, subject)
	if err != nil {
		return err
	}

	msg := postgresMessage{
		Subject:           subject,
		Id:                newId,
		Body:              message.Body,
		ExpirationSeconds: int(math.Round(message.Expiration.Seconds())),
	}

	result := p.db.WithContext(ctx).Create(&msg)
	if result.Error != nil {
		return result.Error
	}

	message.Id = int(msg.Id)

	return nil
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

	message := broker.Message{
		Id:         id,
		Body:       msg.Body,
		Expiration: time.Second * time.Duration(msg.ExpirationSeconds),
	}

	return &message, nil
}

type postgresSequence struct {
	Subject string `gorm:"primaryKey"`
	Val     int32  `gorm:"not null"`
}

func (p *postgresSequence) TableName() string {
	return "sequences"
}

type postgresMessage struct {
	Subject           string `gorm:"primaryKey"`
	Id                int32  `gorm:"primaryKey;autoIncrement:false"`
	Body              string
	ExpirationSeconds int
}

func (p *postgresMessage) TableName() string {
	return "messages"
}
