package broker

import (
	"context"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	service broker.Broker
	mainCtx = context.Background()
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	service = NewModule()
	m.Run()
}

func TestPublishShouldFailOnClosed(t *testing.T) {
	service = NewModule()
	msg := createMessage()

	err := service.Close()
	assert.Nil(t, err)

	_, err = service.Publish(mainCtx, "ali", msg)
	assert.Equal(t, broker.ErrUnavailable, err)
}

func TestSubscribeShouldFailOnClosed(t *testing.T) {
	service = NewModule()
	err := service.Close()
	assert.Nil(t, err)

	_, err = service.Subscribe(mainCtx, "ali")
	assert.Equal(t, broker.ErrUnavailable, err)
}

func TestFetchShouldFailOnClosed(t *testing.T) {
	service = NewModule()
	err := service.Close()
	assert.Nil(t, err)

	_, err = service.Fetch(mainCtx, "ali", rand.Intn(100))
	assert.Equal(t, broker.ErrUnavailable, err)
}

func TestPublishShouldNotFail(t *testing.T) {
	service = NewModule()
	msg := createMessage()

	_, err := service.Publish(mainCtx, "ali", msg)

	assert.Equal(t, nil, err)
}

func TestSubscribeShouldNotFail(t *testing.T) {
	service = NewModule()
	sub, err := service.Subscribe(mainCtx, "ali")

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, sub)
}

func TestPublishShouldSendMessageToSubscribedChan(t *testing.T) {
	service = NewModule()
	msg := createMessage()

	sub, _ := service.Subscribe(mainCtx, "ali")
	_, _ = service.Publish(mainCtx, "ali", msg)
	in := <-sub

	assertMessagesEqual(t, msg, in)
}

func TestPublishShouldSendMessageToSubscribedChans(t *testing.T) {
	service = NewModule()
	msg := createMessage()

	sub1, _ := service.Subscribe(mainCtx, "ali")
	sub2, _ := service.Subscribe(mainCtx, "ali")
	sub3, _ := service.Subscribe(mainCtx, "ali")
	_, _ = service.Publish(mainCtx, "ali", msg)
	in1 := <-sub1
	in2 := <-sub2
	in3 := <-sub3

	assertMessagesEqual(t, msg, in1)
	assertMessagesEqual(t, msg, in2)
	assertMessagesEqual(t, msg, in3)
}

func TestPublishShouldPreserveOrder(t *testing.T) {
	service = NewModule()
	n := 50
	messages := make([]broker.Message, n)
	sub, _ := service.Subscribe(mainCtx, "ali")
	for i := 0; i < n; i++ {
		messages[i] = createMessage()
		_, _ = service.Publish(mainCtx, "ali", messages[i])
	}

	for i := 0; i < n; i++ {
		msg := <-sub
		assertMessagesEqual(t, messages[i], msg)
	}
}

func TestPublishShouldNotSendToOtherSubscriptions(t *testing.T) {
	service = NewModule()
	msg := createMessage()
	ali, _ := service.Subscribe(mainCtx, "ali")
	maryam, _ := service.Subscribe(mainCtx, "maryam")

	_, _ = service.Publish(mainCtx, "ali", msg)
	select {
	case m := <-ali:
		assertMessagesEqual(t, msg, m)
	case <-maryam:
		assert.Fail(t, "Wrong message received")
	}
}

func TestNonExpiredMessageShouldBeFetchable(t *testing.T) {
	service = NewModule()
	msg := createMessageWithExpire(time.Second * 10)
	id, _ := service.Publish(mainCtx, "ali", msg)
	fMsg, _ := service.Fetch(mainCtx, "ali", id)

	assertMessagesEqual(t, msg, fMsg)
}

func TestExpiredMessageShouldNotBeFetchable(t *testing.T) {
	service = NewModule()
	msg := createMessageWithExpire(time.Millisecond * 500)
	id, _ := service.Publish(mainCtx, "ali", msg)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	<-ticker.C
	fMsg, err := service.Fetch(mainCtx, "ali", id)
	assert.Equal(t, broker.ErrExpiredID, err)
	assert.Equal(t, broker.Message{}, fMsg)
}

func TestNewSubscriptionShouldNotGetPreviousMessages(t *testing.T) {
	service = NewModule()
	msg := createMessage()
	_, _ = service.Publish(mainCtx, "ali", msg)
	sub, _ := service.Subscribe(mainCtx, "ali")

	select {
	case <-sub:
		assert.Fail(t, "Got previous message")
	default:
	}
}

func TestConcurrentSubscribesOnOneSubjectShouldNotFail(t *testing.T) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := service.Subscribe(mainCtx, "ali")
				assert.Nil(t, err)
			}()
		}
	}
}

func TestConcurrentSubscribesShouldNotFail(t *testing.T) {
	service = NewModule()
	ticker := time.NewTicker(2000 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := service.Subscribe(mainCtx, randomString(4))
				assert.Nil(t, err)
			}()
		}
	}
}

func TestConcurrentPublishOnOneSubjectShouldNotFail(t *testing.T) {
	service = NewModule()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	msg := createMessage()

	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := service.Publish(mainCtx, "ali", msg)
				assert.Nil(t, err)
			}()
		}
	}
}

func TestConcurrentPublishShouldNotFail(t *testing.T) {
	service = NewModule()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var wg sync.WaitGroup

	msg := createMessage()

	for {
		select {
		case <-ticker.C:
			wg.Wait()
			return

		default:
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := service.Publish(mainCtx, randomString(4), msg)
				assert.Nil(t, err)
			}()
		}
	}
}

func TestDataRace(t *testing.T) {
	service = NewModule()
	duration := 500 * time.Millisecond
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	var wg sync.WaitGroup

	ids := make(chan int, 100000)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				return

			default:
				id, err := service.Publish(mainCtx, "ali", createMessageWithExpire(duration))
				ids <- id
				assert.Nil(t, err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				return

			default:
				_, err := service.Subscribe(mainCtx, "ali")
				assert.Nil(t, err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				return

			case id := <-ids:
				_, err := service.Fetch(mainCtx, "ali", id)
				assert.Nil(t, err)
			}
		}
	}()

	wg.Wait()
}

func BenchmarkPublish(b *testing.B) {
	service = NewModule()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.Publish(mainCtx, randomString(2), createMessage())
		assert.Nil(b, err)
	}
}

func BenchmarkSubscribe(b *testing.B) {
	service = NewModule()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.Subscribe(mainCtx, randomString(2))
		assert.Nil(b, err)
	}
}

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createMessage() broker.Message {
	body := randomString(16)

	return broker.Message{
		Body:       body,
		Expiration: 0,
	}
}

func createMessageWithExpire(duration time.Duration) broker.Message {
	body := randomString(16)

	return broker.Message{
		Body:       body,
		Expiration: duration,
	}
}

func assertMessagesEqual(t *testing.T, want broker.Message, got broker.Message) {
	assert.Equal(t, want.Body, got.Body)
	assert.Equal(t, want.Expiration, got.Expiration)
}
