package store

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gocql/gocql/otelgocql"
	"go.opentelemetry.io/otel/trace"
	"log"
	"math"
	"time"
)

type CassandraConfig struct {
	Host     string `config:"host"`
	Keyspace string `config:"keyspace"`
}

type cassandra struct {
	session   *gocql.Session
	config    CassandraConfig
	sequences Sequence
}

func NewCassandra(config CassandraConfig, sequence Sequence, tracerProvider trace.TracerProvider) (Message, error) {
	ctx := context.Background()
	{
		cluster := gocql.NewCluster(config.Host)
		session, err := otelgocql.NewSessionWithTracing(ctx, cluster, otelgocql.WithTracerProvider(tracerProvider))
		defer session.Close()
		if err != nil {
			return nil, err
		}
		createKeyspaceStatement :=
			fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};", config.Keyspace)
		if err := session.Query(createKeyspaceStatement).Exec(); err != nil {
			return nil, err
		}

		log.Printf("keyspace %q created\n", config.Keyspace)
	}

	cluster := gocql.NewCluster(config.Host)
	cluster.Consistency = gocql.All
	cluster.Keyspace = config.Keyspace

	session, err := otelgocql.NewSessionWithTracing(ctx, cluster, otelgocql.WithTracerProvider(tracerProvider))
	if err != nil {
		return nil, err
	}

	c := &cassandra{
		session:   session,
		config:    config,
		sequences: sequence,
	}

	err = c.init()
	if err != nil {
		return nil, err
	}

	if err := c.loadSequences(ctx); err != nil {
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

	return nil
}

func (c *cassandra) loadSequences(ctx context.Context) error {
	iter := c.session.Query(
		"SELECT subject, MAX(id) FROM messages_by_subject_and_id GROUP BY subject ;",
	).WithContext(ctx).Iter()

	var subject string
	var lastId int
	for iter.Scan(&subject, &lastId) {
		if err := c.sequences.Load(ctx, subject, int32(lastId)); err != nil {
			return err
		}
	}

	return iter.Close()
}

func (c *cassandra) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	newId, err := c.sequences.CreateNewId(ctx, subject)
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
		if err == gocql.ErrNotFound {
			return nil, ErrExpired
		}
		return nil, err
	}

	message.Expiration = time.Duration(expiration.Nanoseconds)
	return &message, nil
}
