package routes

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sing3demons/go-http-service/logger"
	"golang.org/x/net/http2"
)

type IMicroservice interface {
	Start()
	// HTTP Services
	Logger(next http.Handler) http.Handler
	GET(path string, h ServiceHandleFunc)
	POST(path string, h ServiceHandleFunc)
	PUT(path string, h ServiceHandleFunc)
	PATCH(path string, h ServiceHandleFunc)
	DELETE(path string, h ServiceHandleFunc)
}

type microservice struct {
	logger logger.ILogger
	mux    *http.ServeMux
}

const Key = "logger"
const XSession = "X-Request-Id"

func NewRouter() IMicroservice {
	mux := http.NewServeMux()
	lg := logger.NewLoggerWrapper("logrus", context.Background())
	return &microservice{logger: lg, mux: mux}
}

func (m *microservice) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqId := w.Header().Get(XSession)
		if reqId == "" {
			reqId = uuid.NewString()
			w.Header().Set(XSession, reqId)
		}

		// Set the logger in the context
		ctx := context.WithValue(r.Context(), ContextKey(XSession), reqId)
		r = r.WithContext(ctx)
		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the request
		m.logger.Info("Request", map[string]any{
			"method":     r.Method,
			"requestURI": r.RequestURI,
			"remoteAddr": r.RemoteAddr,
			"duration":   time.Since(start),
			"sessionId":  reqId,
		})
	})
}

func removeLeftBraces(str string) string {
	return strings.ReplaceAll(str, "{", "") // Remove left brace
}

func removeRightBrace(str string) string {
	return strings.ReplaceAll(str, "}", "") // Remove right brace
}

func (m *microservice) GET(path string, handler ServiceHandleFunc) {
	m.mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		r = setParam(path, r)
		handler(NewMyContext(w, r))
	})
}

type ContextKey string

func setParam(path string, r *http.Request) *http.Request {
	var paramKey ContextKey
	var paramValue string
	subPath := strings.Split(path, "/")
	for _, v := range subPath {
		if v != "" {
			if strings.HasPrefix(v, "{") {
				key := removeRightBrace(removeLeftBraces(v))
				if key != "" {
					paramKey = ContextKey(key)
				}
			} else {
				paramValue = splitPath(subPath, r.URL.Path, paramValue)
			}

		}
	}
	if paramKey != "" && paramValue != "" {
		ctx := context.WithValue(r.Context(), paramKey, paramValue)
		r = r.WithContext(ctx)
	}
	return r
}

func splitPath(subPath []string, urlPath string, paramValue string) string {
	for _, v := range subPath {
		if v != "" {
			if strings.Contains(urlPath, v) {
				paramValue = urlPath[strings.Index(urlPath, v):]
				paramValue = strings.ReplaceAll(paramValue, v, "")
				paramValue = strings.ReplaceAll(paramValue, "/", "")
			}
		}
	}
	return paramValue
}

func (m *microservice) POST(path string, handler ServiceHandleFunc) {
	m.mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		r = setParam(path, r)
		handler(NewMyContext(w, r))
	})
}

func (m *microservice) PUT(path string, handler ServiceHandleFunc) {
	m.mux.HandleFunc("PUT "+path, func(w http.ResponseWriter, r *http.Request) {
		r = setParam(path, r)
		handler(NewMyContext(w, r))
	})
}

func (m *microservice) PATCH(path string, handler ServiceHandleFunc) {
	m.mux.HandleFunc("PATCH "+path, func(w http.ResponseWriter, r *http.Request) {
		handler(NewMyContext(w, r))
	})
}

func (m *microservice) DELETE(path string, handler ServiceHandleFunc) {
	m.mux.HandleFunc("DELETE "+path, func(w http.ResponseWriter, r *http.Request) {
		r = setParam(path, r)
		handler(NewMyContext(w, r))
	})
}

func ReadCertAndKey() (cert, key string, err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	files, err := os.ReadDir(pwd + "/certificates")
	if err != nil {
		return "", "", err
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "cert.pem") {
			cert = pwd + "/certificates/cert.pem"
		} else if strings.Contains(file.Name(), "key.pem") {
			key = pwd + "/certificates/key.pem"
		}
	}

	return cert, key, nil

}

func (m *microservice) Start() {
	caFile, key, err := ReadCertAndKey()
	if err != nil {
		m.logger.Info("server started without TLS", map[string]any{})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	var wait time.Duration
	srv := &http.Server{
		Handler:      m.Logger(m.mux),
		Addr:         ":" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	go func() {
		hostName, err := os.Hostname()
		m.logger.Info("server started on port "+srv.Addr, map[string]any{
			"port":     srv.Addr,
			"hostName": hostName,
			"pid":      os.Getpid(),
			"ppid":     os.Getppid(),
			"uid":      os.Getuid(),
			"gid":      os.Getgid(),
			"error":    err,
		})

		if caFile != "" && key != "" {
			if err := m.StartTLS(caFile, port); err != nil {
				fmt.Printf("server listen err: %v\n", err)
				log.Fatal(err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("server listen err: %v\n", err)
				log.Fatal(err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown: ", err)
	}
	fmt.Println("server exited")
}

func (m *microservice) StartTLS(caFile string, port string) error {
	caCert, errReadFile := readFile(caFile, "CA Cert")
	if errReadFile != nil {
		return fmt.Errorf("failed to read ca certificate %s: %v", caFile, errReadFile)
	}
	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to append CA certificate %s", caFile)
	}
	serverCert, err := getServerCert()
	if err != nil {
		return fmt.Errorf("failed to get server certificate: %v", err)
	}

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS13,
		MaxVersion:       tls.VersionTLS13,
		CipherSuites:     Ciphers,
		NextProtos:       []string{HTTP2, HTTP11, ALPNProto},
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		Certificates:     []tls.Certificate{*serverCert},
		ClientAuth:       tls.VerifyClientCertIfGiven,
		Rand:             rand.Reader,
		RootCAs:          caCertPool,
		ClientCAs:        caCertPool,
	}

	tcpListener, _ := tls.Listen("tcp", ":"+port, tlsConfig)

	s2 := &http2.Server{
		MaxHandlers:                  0,
		MaxConcurrentStreams:         0,
		MaxDecoderHeaderTableSize:    0,
		MaxEncoderHeaderTableSize:    0,
		MaxReadFrameSize:             0,
		PermitProhibitedCipherSuites: true,
		IdleTimeout:                  10 * time.Second,
		MaxUploadBufferPerConnection: 65535,
		MaxUploadBufferPerStream:     1,
		NewWriteScheduler:            nil,
		CountError:                   nil,
	}

	server := &http.Server{
		Handler:           m.Logger(m.mux),
		Addr:              ":" + port,
		ReadHeaderTimeout: 120 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadTimeout:       120 * time.Second,
		TLSConfig:         tlsConfig,
		MaxHeaderBytes:    1048576,
	}

	if err := http2.ConfigureServer(server, s2); err != nil {
		return fmt.Errorf("failed to configure server for http2: %v", err)
	}

	if err := server.Serve(tcpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve server: %v", err)
	}
	return nil
}

var Ciphers = []uint16{
	// TLS 1.3
	tls.TLS_AES_256_GCM_SHA384,
	tls.TLS_AES_128_GCM_SHA256,
	tls.TLS_CHACHA20_POLY1305_SHA256,

	// ECDSA is about 3 times faster than RSA on the server side.
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,

	// RSA is slower on the server side but still widely used.
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
}
var (
	HTTP11    = "http/1.1"
	HTTP2     = "h2"
	ALPNProto = "acme-tls/1"
)

func getServerCert() (*tls.Certificate, error) {
	cert := "certificates/cert.pem"

	certBytes, err := readFile(cert, "Cert")
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate %s: %v", cert, err)
	}

	key := "certificates/key.pem"
	certKeyBytes, err := readFile(key, "Key")
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate key %s: %v", key, err)
	}

	serverCert, errCert := tls.X509KeyPair(certBytes, certKeyBytes)
	if errCert != nil {
		return nil, fmt.Errorf("failed to load certificate %s and key %s: %v", cert, key, errCert)
	}

	return &serverCert, nil
}

func readFile(file, errorPrefix string) ([]byte, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return nil, fmt.Errorf("%s file cannot be blank", errorPrefix)
	}

	osf, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("%s file not found %s: %v", errorPrefix, file, err)
	}

	if osf.IsDir() {
		return nil, fmt.Errorf("%s file needs to specify a file in its path", errorPrefix)
	}
	if osf.Size() < 1 {
		return nil, fmt.Errorf("%s file cannot be empty", errorPrefix)
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file %s: %v", errorPrefix, file, err)
	}

	if len(bytes) < 1 {
		return nil, fmt.Errorf("%s file %s is empty", errorPrefix, file)
	}

	return bytes, nil
}
