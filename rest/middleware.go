package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/DoomLordor/logger"
)

const UserKey = "user"

type ErrorResponse struct {
	Error string `json:"error"`
}

type WsResponse struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type AuthFunc func(token string) (any, error)

type Middlewares struct {
	authFunc  AuthFunc
	logger    *logger.Logger
	requestId *atomic.Uint64
	upgrader  *websocket.Upgrader
}

func NewMiddlewares(authFunc AuthFunc, logger *logger.Logger) *Middlewares {
	return &Middlewares{
		authFunc:  authFunc,
		logger:    logger,
		requestId: &atomic.Uint64{},
		upgrader:  &websocket.Upgrader{},
	}
}

func (m *Middlewares) TokenMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			text := "No Authorization Header"
			m.logger.Warn().Msg(text)
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: text})
			return
		}
		rawToken := strings.Split(header, " ")

		if len(rawToken) != 2 || rawToken[0] != "Bearer" {
			text := "Invalid token"
			m.logger.Warn().Msg(text)
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: text})
			return
		}
		token := rawToken[1]

		if user, err := m.authFunc(token); err == nil {
			//m.logger.Info().Msgf("Authenticated user %s\n", user)
			ctx := context.WithValue(r.Context(), UserKey, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		} else {
			m.logger.Warn().Msg(err.Error())
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}
	}
	return http.HandlerFunc(f)
}

func (m *Middlewares) CommonMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		headers := "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header"
		w.Header().Set("Access-Control-Allow-Headers", headers)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func (m *Middlewares) TimeMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		requestId := m.requestId.Add(1)
		writer := NewLoggingResponseWriter(w)
		start := time.Now().UnixMilli()
		m.logger.Info().
			Str("method", r.Method).
			Str("url", r.RequestURI).
			Uint64("requestId", requestId).
			Msg("Start")

		next.ServeHTTP(writer, r)
		end := time.Now().UnixMilli() - start
		m.logger.Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("code", writer.Code()).
			Int64("response_time", end).
			Uint64("requestId", requestId).
			Msg("End")
	}
	return http.HandlerFunc(f)
}

func (m *Middlewares) LoggingMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		requestId := m.requestId.Add(1)
		m.logger.Info().
			Str("url", r.RequestURI).
			Uint64("requestId", requestId).
			Msg("connect")

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func (m *Middlewares) RecoveryMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		defer m.recover(w)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (m *Middlewares) recover(w http.ResponseWriter) {
	r := recover()
	if r != nil {
		var err error
		switch t := r.(type) {
		case string:
			err = errors.New(t)
		case error:
			err = t
		default:
			err = errors.New("unknown error")
		}

		m.logger.Err(err).Msg(string(debug.Stack()))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *Middlewares) HandleWrapper(hf HandlerFuncRest) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		res, code, err := hf(r)
		w.WriteHeader(code)
		if err != nil {
			res = ErrorResponse{Error: err.Error()}
			if code >= http.StatusInternalServerError {
				m.logger.Err(err).Str("method", r.Method).Str("url", r.RequestURI).Send()
			} else {
				m.logger.Warn().
					Str("method", r.Method).
					Str("url", r.RequestURI).
					Str("warning", err.Error()).
					Send()
			}
		}
		if res != nil {
			st, ok := res.(string)
			if ok {
				_, _ = io.WriteString(w, st)
				return
			}
			_ = json.NewEncoder(w).Encode(res)
		}
	}
	return http.HandlerFunc(f)
}

func (m *Middlewares) HandleWsWrapper(hf HandlerFuncWs) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		conn, err := m.upgrader.Upgrade(w, r, nil)
		if err != nil {
			m.logger.Err(err).Msgf("WS Url: %s", r.RequestURI)
			return
		}

		conn.SetPingHandler(nil)
		conn.SetPongHandler(nil)
		conn.SetCloseHandler(nil)

		code, err := hf(conn)

		if err != nil {
			switch code {
			case websocket.CloseNormalClosure:
				m.logger.Warn().Str("ws_url", r.RequestURI).Str("warning", err.Error()).Send()
			default:
				m.logger.Err(err).Str("ws_url", r.RequestURI).Send()
			}

			_ = conn.WriteJSON(
				WsResponse{
					Type: "errorSystem",
					Data: ErrorResponse{
						Error: err.Error(),
					},
				},
			)
		}

		closeFunc := conn.CloseHandler()
		err = closeFunc(websocket.CloseMessage, "")
		if err != nil {
			m.logger.Err(err).Msg("WS close error")
			return
		}

		err = conn.Close()

		if err != nil {
			m.logger.Err(err).Msg("WS close error")
		}
	}
	return http.HandlerFunc(f)
}
