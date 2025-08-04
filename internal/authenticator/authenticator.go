package authenticator

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

type Authenticator struct {
	logger  *slog.Logger
	conn    *nats.Conn
	users   map[string]User
	keypair nkeys.KeyPair
	sub     *nats.Subscription
}

func New(logger *slog.Logger, cfg Config) (*Authenticator, error) {
	nc, err := nats.Connect(cfg.URL, nats.RetryOnFailedConnect(true))
	if err != nil {
		return nil, fmt.Errorf("fialed to create nats connection %w", err)
	}

	logger.Info("connected successfully to nats")

	kp, err := nkeys.FromSeed([]byte(cfg.NkeySeed))
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair from seed %w", err)
	}

	return &Authenticator{
		logger:  logger,
		conn:    nc,
		users:   cfg.Users,
		keypair: kp,
		sub:     nil,
	}, nil
}

func (auth *Authenticator) Start() error {
	if auth.sub != nil {
		return nil
	}

	sub, err := auth.conn.Subscribe("$SYS.REQ.USER.AUTH", auth.handler)
	if err != nil {
		return fmt.Errorf("error subscribing to authentication subjec %w", err)
	}

	auth.sub = sub

	return nil
}

func (auth *Authenticator) Stop() error {
	err := auth.sub.Drain()
	if err != nil {
		return fmt.Errorf("draining connection failed %w", err)
	}

	return nil
}

// nolint: funlen
func (auth *Authenticator) handler(msg *nats.Msg) {
	begin := time.Now()

	auth.logger.Info(
		"received authentication request",
		slog.String("subject", msg.Subject),
		slog.String("reply", msg.Reply),
	)

	rc, err := jwt.DecodeAuthorizationRequestClaims(string(msg.Data))
	if err != nil {
		auth.logger.Error("decoding authentication request failed", slog.String("error", err.Error()))

		_ = msg.Respond([]byte("failed"))

		return
	}

	response := jwt.NewAuthorizationResponseClaims(rc.UserNkey)
	response.Audience = rc.Server.ID

	defer func() {
		encResponse, err := response.Encode(auth.keypair)
		if err != nil {
			auth.logger.Error("failed to encode response", slog.String("error", err.Error()))

			_ = msg.Respond([]byte("failed"))

			return
		}

		err = msg.Respond([]byte(encResponse))
		if err != nil {
			auth.logger.Error("failed to send back response", slog.String("error", err.Error()))
		}
	}()

	auth.logger.Info(
		"new client wants to connect",
		slog.String("username", rc.ConnectOptions.Username),
		slog.String("password", rc.ConnectOptions.Password),
		slog.String("host", rc.ClientInformation.Host),
		slog.String("name", rc.ClientInformation.Name),
		slog.String("kind", rc.ClientInformation.Kind),
	)

	user, ok := auth.users[rc.ConnectOptions.Username]
	if !ok || user.Password != rc.ConnectOptions.Password {
		auth.logger.Error(
			"invalid username and password",
			slog.String("username", rc.ConnectOptions.Username),
			slog.String("password", rc.ConnectOptions.Password),
		)

		response.Error = "invalid username and password"

		return
	}

	claims := jwt.NewUserClaims(rc.UserNkey)
	claims.Audience = user.Account
	claims.Name = rc.ConnectOptions.Username

	// Apply permissions based on privilege level
	if user.Privileged {
		// Privileged users get full access
		claims.Permissions = jwt.Permissions{
			Pub: jwt.Permission{
				Allow: []string{">"}, // Allow publishing to all subjects
				Deny:  []string{},
			},
			Sub: jwt.Permission{
				Allow: []string{">"}, // Allow subscribing to all subjects
				Deny:  []string{},
			},
			Resp: &jwt.ResponsePermission{
				MaxMsgs: 0,
				Expires: 0,
			},
		}
	} else {
		claims.Permissions = jwt.Permissions{
			Pub: jwt.Permission{
				Allow: []string{">"}, // Allow publishing to all subjects
				Deny: []string{
					"$JS.API.STREAM.CREATE.>", "$JS.API.STREAM.DELETE.>", "$JS.API.STREAM.PURGE.>",
					"$JS.API.STREAM.PEER.REMOVE.>", "$JS.API.STREAM.LEADER.STEPDOWN.>",
				}, // Deny JetStream API access
			},
			Sub: jwt.Permission{
				Allow: []string{">"}, // Allow subscribing to all subjects
				Deny: []string{
					"$JS.API.STREAM.CREATE.>", "$JS.API.STREAM.DELETE.>",
					"$JS.API.STREAM.PURGE.>", "$JS.API.STREAM.PEER.REMOVE.>", "$JS.API.STREAM.LEADER.STEPDOWN.>",
				},
			},
			Resp: &jwt.ResponsePermission{
				MaxMsgs: 0,
				Expires: 0,
			},
		}
	}

	vr := jwt.CreateValidationResults()

	claims.Validate(vr)

	if len(vr.Errors()) > 0 {
		auth.logger.Error("failed to validate claims", slog.String("error", errors.Join(vr.Errors()...).Error()))

		response.Error = "internal server error"

		return
	}

	token, err := claims.Encode(auth.keypair)
	if err != nil {
		auth.logger.Error("failed to encode claims", slog.String("error", err.Error()))

		response.Error = "internal server error"

		return
	}

	response.Jwt = token

	auth.logger.Info("authentication successed", slog.Float64("took (s)", time.Since(begin).Seconds()))
}
