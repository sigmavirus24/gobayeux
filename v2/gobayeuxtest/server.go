package gobayeuxtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/sigmavirus24/gobayeux/v2"
)

const (
	VERSION = "1.0"
)

var (
	chars    = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmonpqrstuvwxyz0123456789")
	numChars = len(chars)
	advice   = &gobayeux.Advice{
		Reconnect: "handshake",
		Timeout:   int(30 * time.Second),
		Interval:  int(1 * time.Second),
	}
)

type Logger interface {
	Log(args ...any)
	Logf(format string, args ...any)
}

type Server struct {
	log Logger

	mu   sync.Mutex
	subs map[string][]gobayeux.Channel
}

func NewServer(logger Logger) *Server {
	return &Server{
		log:  logger,
		subs: make(map[string][]gobayeux.Channel),
	}
}

func (s *Server) RoundTrip(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	defer req.Body.Close()

	var msgs []*gobayeux.Message

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("issue reading body (%w)", err)
	}

	if err := json.Unmarshal(body, &msgs); err != nil {
		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Status:     http.StatusText(http.StatusUnprocessableEntity),
		}, nil
	}

	replies := []*gobayeux.Message{}
	statusCode := http.StatusOK

	for _, msg := range msgs {
		s.log.Logf("msg: %+v\n", msg)
		switch msg.Channel {
		case "/meta/handshake":
			replies = append(replies, &gobayeux.Message{
				Channel:                  "/meta/handshake",
				Version:                  msg.Version,
				SupportedConnectionTypes: msg.SupportedConnectionTypes,
				ClientID:                 generateID(10),
				Successful:               true,
				AuthSuccessful:           true,
				Advice:                   advice,
				ID:                       msg.ID,
			})
		case "/meta/connect":
			if channels, ok := s.subs[msg.ClientID]; ok {
				for _, ch := range channels {
					replies = append(replies, &gobayeux.Message{
						Channel:    ch,
						ID:         generateID(5),
						ClientID:   msg.ClientID,
						Data:       json.RawMessage(`{}`),
						Successful: true,
					})
				}
			}

			replies = append(replies, &gobayeux.Message{
				Channel:    "/meta/connect",
				Successful: true,
				ClientID:   msg.ClientID,
				Advice:     advice,
				ID:         msg.ID,
			})
		case "/meta/subscribe":
			if _, ok := s.subs[msg.ClientID]; !ok {
				s.subs[msg.ClientID] = make([]gobayeux.Channel, 0)
			}

			reply := &gobayeux.Message{
				Channel:      "/meta/subscribe",
				ID:           msg.ID,
				ClientID:     msg.ClientID,
				Successful:   true,
				Subscription: msg.Subscription,
			}

			for _, ch := range s.subs[msg.ClientID] {
				if ch == msg.Subscription {
					statusCode = http.StatusBadRequest
					reply.Successful = false
					reply.Error = "403:%s:already subscribed"
				}
			}

			s.subs[msg.ClientID] = append(s.subs[msg.ClientID], msg.Subscription)

			replies = append(replies, reply)
		default:
			s.log.Logf("unhandled: %+v", msg)
		}
	}

	reply, err := json.Marshal(replies)
	if err != nil {
		return nil, fmt.Errorf("issue marshaling body (%w)", err)
	}

	s.log.Logf("reply: %s\n", reply)

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(reply)),
	}, nil
}

func generateID(length int) string {
	ret := make([]rune, length)
	for i := range ret {
		ret[i] = chars[rand.Intn(numChars)]
	}

	return string(ret)
}
