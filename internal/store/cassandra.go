package store

import (
	"context"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"github.com/gocql/gocql"
	"math"
	"sync"
	"time"
)

type CassandraConfig struct {
	Host     string `config:"host"`
	Keyspace string `config:"keyspace"`
}

type cassandra struct {
	session *gocql.Session
	config  CassandraConfig
	locker  sync.Mutex
}

func NewCassandra(config CassandraConfig) (Message, error) {
	cluster := gocql.NewCluster(config.Host)
	cluster.Consistency = gocql.All
	cluster.Keyspace = config.Keyspace

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	c := &cassandra{
		session: session,
		config:  config,
	}

	err = c.init()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *cassandra) init() error {
	ctx := context.Background()

	if err := c.session.Query(
		"CREATE TABLE IF NOT EXISTS messages_by_subject_and_id (subject text, id int, body text, expiration duration, PRIMARY KEY (subject, id));",
	).WithContext(ctx).Exec(); err != nil {
		return err
	}

	if err := c.session.Query(
		"CREATE TABLE IF NOT EXISTS sequences (subject text PRIMARY KEY, currentId counter);",
	).WithContext(ctx).Exec(); err != nil {
		return err
	}

	return nil
}

func (c *cassandra) createNewId(ctx context.Context, subject string) (int32, error) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if err := c.session.Query(
		"UPDATE sequences SET currentId = currentId + 1 WHERE subject = ?;",
		subject,
	).WithContext(ctx).Exec(); err != nil {
		return 0, err
	}

	var id int32
	if err := c.session.Query(
		"SELECT currentId FROM sequences WHERE subject = ?;",
		subject,
	).WithContext(ctx).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (c *cassandra) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	newId, err := c.createNewId(ctx, subject)
	if err != nil {
		return err
	}

	expirationSeconds := int(math.Round(message.Expiration.Seconds()))
	if expirationSeconds <= 0 {
		message.Id = int(newId)
		return nil
	}

	if err := c.session.Query(
		"INSERT INTO messages_by_subject_and_id (subject, id, body, expiration) VALUES (?, ?, ?, ?) USING TTL ?;",
		subject,
		newId,
		message.Body,
		message.Expiration,
		expirationSeconds,
	).WithContext(ctx).Exec(); err != nil {
		return err
	}

	message.Id = int(newId)

	return nil
}

func (c *cassandra) GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error) {
	var message broker.Message
	var expiration gocql.Duration

	if err := c.session.Query(
		"SELECT id, body, expiration FROM messages_by_subject_and_id WHERE subject=? AND id=?;",
		subject,
		id,
	).WithContext(ctx).Scan(&message.Id, &message.Body, &expiration); err != nil {
		return nil, err
	}

	message.Expiration = time.Duration(expiration.Nanoseconds)
	return &message, nil
}
