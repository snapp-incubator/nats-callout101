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
	account string
	keypair nkeys.KeyPair
	sub     *nats.Subscription
}

func New(logger *slog.Logger, cfg Config) (*Authenticator, error) {
	nc, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("fialed to create nats connection %w", err)
	}

	kp, err := nkeys.FromSeed([]byte(cfg.NkeySeed))
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair from seed %w", err)
	}

	return &Authenticator{
		logger:  logger,
		conn:    nc,
		account: cfg.Account,
		keypair: kp,
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
	if err := auth.sub.Drain(); err != nil {
		return fmt.Errorf("draining connection failed %w", err)
	}

	return nil
}

func (auth *Authenticator) handler(msg *nats.Msg) {
	begin := time.Now()

	auth.logger.Info("received authentication request", slog.String("subject", msg.Subject), slog.String("reply", msg.Reply))

	rc, err := jwt.DecodeAuthorizationRequestClaims(string(msg.Data))
	if err != nil {
		auth.logger.Error("decoding authentication request failed", slog.String("error", err.Error()))

		_ = msg.Respond([]byte("failed"))
		return
	}

	userId := rc.ConnectOptions.Username
	auth.logger.Info("new client wants to connect", slog.String("username", rc.ConnectOptions.Username), slog.String("password", rc.ConnectOptions.Password))

	if userId == "" {
		_ = msg.Respond([]byte("failed"))
		return
	}

	claims := jwt.NewUserClaims(rc.UserNkey)
	claims.Audience = auth.account
	claims.Name = userId
	claims.Permissions = jwt.Permissions{
		Pub: jwt.Permission{
			Allow: jwt.StringList{
				"$JS.API.INFO", // General JS Info

				// Chat permisions
				fmt.Sprintf("chat.*.%s", userId),            // Publishing chat messages for this user id
				"$JS.API.STREAM.INFO.chat_messages",         // Getting info on chat_messages stream
				"$JS.API.CONSUMER.CREATE.chat_messages.>",   // Creating consumers on chat_messages stream
				"$JS.API.CONSUMER.MSG.NEXT.chat_messages.>", // Creating consumers on chat_messages stream

				// Workspace permissions
				"$JS.API.DIRECT.GET.KV_chat_workspace.>",        // Gets from workspace KV
				"$JS.API.STREAM.INFO.KV_chat_workspace",         // Info about workspace KV
				"$JS.API.CONSUMER.CREATE.KV_chat_workspace.*.>", // Creating consumers/watchers on workspace KV
			},
		},
	}

	vr := jwt.CreateValidationResults()

	claims.Validate(vr)
	if len(vr.Errors()) > 0 {
		auth.logger.Error("failed to validate claims", slog.String("error", errors.Join(vr.Errors()...).Error()))

		_ = msg.Respond([]byte("failed"))
		return
	}

	token, err := claims.Encode(auth.keypair)
	if err != nil {
		auth.logger.Error("failed to encode claims", slog.String("error", err.Error()))

		_ = msg.Respond([]byte("failed"))
		return
	}

	response := jwt.NewAuthorizationResponseClaims(rc.UserNkey)
	response.Audience = rc.Server.ID
	response.Jwt = token

	encResponse, err := response.Encode(auth.keypair)
	if err != nil {
		auth.logger.Error("failed to encode response", slog.String("error", err.Error()))

		_ = msg.Respond([]byte("failed"))
		return
	}

	if err := msg.Respond([]byte(encResponse)); err != nil {
		auth.logger.Error("failed to send back an authenticated response", slog.String("error", err.Error()))
	}

	auth.logger.Info("authentication successed", slog.Float64("took (s)", time.Since(begin).Seconds()))
}
