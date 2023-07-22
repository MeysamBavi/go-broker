package store

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	db *gorm.DB
}

func NewPostgres(config PostgresConfig) (Message, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	return &postgresImpl{db: db}, err
}

func (p *postgresImpl) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	msg := postgresMessage{
		Subject:           subject,
		Body:              message.Body,
		ExpirationSeconds: int(message.Expiration.Seconds()),
	}

	result := p.db.WithContext(ctx).Create(&msg)
	if result.Error != nil {
		return result.Error
	}

	message.Id = int(msg.ID)

	return nil
}

func (p *postgresImpl) GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error) {
	msg := postgresMessage{
		Model: gorm.Model{
			ID: uint(id),
		},
		Subject: subject,
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

type postgresMessage struct {
	gorm.Model
	Subject           string `gorm:"primaryKey"`
	Body              string
	ExpirationSeconds int
}

func (p *postgresMessage) TableName() string {
	return "messages"
}
