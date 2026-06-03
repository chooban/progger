package server

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/chooban/progger/reader/config"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

//go:embed templates/*.html
var templateFS embed.FS

type Server struct {
	router   *gin.Engine
	cfg      *config.Config
	listener net.Listener
}

func NewServer(ctx context.Context, cfg *config.Config, handlers *Handlers) *Server {
	s := &Server{
		cfg: cfg,
	}

	gin.SetMode(gin.ReleaseMode)
	s.router = gin.Default()
	logger, _ := logr.FromContext(ctx)
	s.router.Use(gin.Recovery())
	s.router.Use(requestid.New())
	s.router.Use(func(c *gin.Context) {
		requestID := c.GetString(requestid.Get(c))

		reqLogger := logger.WithValues(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		if body, err := io.ReadAll(c.Request.Body); err == nil {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) > 0 {
				reqLogger.V(1).Info("request body", "body", string(body))
			}
		}

		reqLogger.V(1).Info("request headers", "headers", headerMap(c.Request.Header))

		loggerContext := logr.NewContext(c.Request.Context(), reqLogger)
		c.Request = c.Request.WithContext(loggerContext)

		c.Next()

		reqLogger.V(1).Info("response headers", "headers", headerMap(c.Writer.Header()))
	})
	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
	s.router.SetHTMLTemplate(tmpl)
	ConfigureRoutes(s.router, handlers)

	return s
}

func (s *Server) Run() error {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var operr error
			c.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
			return operr
		},
	}
	ln, err := lc.Listen(context.Background(), "tcp", s.cfg.Host)
	if err != nil {
		return err
	}
	s.listener = ln
	return s.router.RunListener(ln)
}

func (s *Server) Shutdown() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func loggerFromContext(ctx context.Context) logr.Logger {
	if logger, ok := ctx.Value("logger").(logr.Logger); !ok {
		l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		return zerologr.New(&l)
	} else {
		return logger
	}
}

func headerMap(h http.Header) string {
	var b strings.Builder
	first := true
	for k, v := range h {
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(fmt.Sprintf("%s: %s", k, strings.Join(v, "; ")))
	}
	return b.String()
}
