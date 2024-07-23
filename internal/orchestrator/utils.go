package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("some_secret_key")

var Templates = make(map[string]*template.Template)

type Middleware func(http.Handler) http.Handler
type Middlewares []Middleware

type Controller struct {
	Logger        *log.Logger
	NextRequestID func() string
}

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

func GenerateAndSetToken(w http.ResponseWriter, login string) {
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &Claims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtKey))
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func GetLoginFromToken(r *http.Request) (string, error) {
	c, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	tknStr := c.Value
	claims := &Claims{}
	_, err = jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return "", err
	}
	return claims.Login, nil
}
func ValidateToken(r *http.Request) bool {
	c, err := r.Cookie("token")
	if err != nil {
		return false
	}
	tknStr := c.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return false
	}
	return tkn.Valid
}

func RefreshToken(w http.ResponseWriter, r *http.Request) error {
	c, err := r.Cookie("token")
	if err != nil {
		return err
	}
	tknStr := c.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return err
	}
	if !tkn.Valid {
		return errors.New("Token is invalid")
	}
	expirationTime := time.Now().Add(12 * time.Hour)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	return nil
}

func JSONResp(w http.ResponseWriter, message string, status int, token ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := make(map[string]any)
	resp["status"] = status
	resp["message"] = message
	if len(token) != 0 {
		resp["token"] = token[0]
	}
	json.NewEncoder(w).Encode(resp)
}

func ParseTemplateDir(dir string, subdir string) error {
	var paths []string
	err := filepath.Walk(dir+"/"+subdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, path := range entries {
		if !path.IsDir() {
			tmp, err := template.ParseFiles(append(paths, dir+"/"+path.Name())...)
			if err != nil {
				return err
			}
			Templates[path.Name()] = tmp
		}
	}
	return nil
}

func (c *Controller) Favicon(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/favicon.ico" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "../../web/images/calculator.png")
}

func (mws Middlewares) Apply(hdlr http.Handler) http.Handler {
	if len(mws) == 0 {
		return hdlr
	}
	return mws[1:].Apply(mws[0](hdlr))
}

func (c *Controller) Logging(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func(start time.Time) {
			requestID := w.Header().Get("X-Request-Id")
			if requestID == "" {
				requestID = "unknown"
			}
			c.Logger.Println(requestID, req.Method, req.URL.Path, req.RemoteAddr, req.UserAgent(), time.Since(start))
		}(time.Now())
		hdlr.ServeHTTP(w, req)
	})
}

func (c *Controller) Tracing(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = c.NextRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		hdlr.ServeHTTP(w, req)
	})
}

func (c *Controller) Shutdown(ctx context.Context, server *http.Server) context.Context {
	ctx, done := context.WithCancel(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer done()
		<-quit
		signal.Stop(quit)
		close(quit)
		server.ErrorLog.Printf("Server is shutting down...\n")
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			server.ErrorLog.Fatalf("Error while shutting down: %s\n", err)
		}
	}()
	return ctx
}
