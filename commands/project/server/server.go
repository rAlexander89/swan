package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rAlexander89/swan/utils"
)

func WriteServer(projectPath string) error {
	projectName, err := utils.GetProjectName()
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	serverStructStr := `
type Server struct {
    srv            *http.Server
    mux            *http.ServeMux
    app            *app.App
    userService    service.UserService
    middleware     []Middleware
    wg             sync.WaitGroup
}`

	middlewareTypeStr := `
type Middleware func(http.HandlerFunc) http.HandlerFunc

type RouteGroup struct {
    prefix     string
    server     *Server
    middleware []Middleware
}

type RouteRegistrar interface {
    RegisterRoutes(group *RouteGroup)
}`

	newServerStr := `
func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
    application, err := app.NewApp(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize application: %w", err)
    }

    // initialize services
    userService := service.NewUserService(application.PostgresDB())

    server := &Server{
        mux:         http.NewServeMux(),
        app:         application,
        userService: userService,
        middleware:  make([]Middleware, 0),
    }

    // register routes
    userRoutes := users.NewRoutes(userService)
    server.Register(userRoutes)

    return server, nil
}`

	routeGroupStr := `
func (s *Server) Group(prefix string) *RouteGroup {
    return &RouteGroup{
        prefix:     prefix,
        server:     s,
        middleware: make([]Middleware, 0),
    }
}

func (s *Server) Register(registrar RouteRegistrar) {
    group := s.Group("")
    registrar.RegisterRoutes(group)
}

func (s *Server) Use(middleware ...Middleware) {
    s.middleware = append(s.middleware, middleware...)
}`

	routeHandlerStr := `
func (g *RouteGroup) Group(prefix string) *RouteGroup {
    return &RouteGroup{
        prefix:     g.prefix + prefix,
        server:     g.server,
        middleware: g.middleware,
    }
}

func (g *RouteGroup) Use(middleware ...Middleware) {
    g.middleware = append(g.middleware, middleware...)
}

func (g *RouteGroup) Handle(method, path string, handler http.HandlerFunc) {
    fullPath := g.prefix + path
    
    finalHandler := handler
    
    for i := len(g.middleware) - 1; i >= 0; i-- {
        finalHandler = g.middleware[i](finalHandler)
    }
    
    for i := len(g.server.middleware) - 1; i >= 0; i-- {
        finalHandler = g.server.middleware[i](finalHandler)
    }

    g.server.mux.HandleFunc(fullPath, finalHandler)
}`

	methodsStr := `
func (g *RouteGroup) GET(path string, handler http.HandlerFunc) {
    g.Handle(http.MethodGet, path, handler)
}

func (g *RouteGroup) POST(path string, handler http.HandlerFunc) {
    g.Handle(http.MethodPost, path, handler)
}

func (g *RouteGroup) PUT(path string, handler http.HandlerFunc) {
    g.Handle(http.MethodPut, path, handler)
}

func (g *RouteGroup) PATCH(path string, handler http.HandlerFunc) {
    g.Handle(http.MethodPatch, path, handler)
}

func (g *RouteGroup) DELETE(path string, handler http.HandlerFunc) {
    g.Handle(http.MethodDelete, path, handler)
}`

	runServerStr := `
func (s *Server) Run(port string) error {
    s.srv = &http.Server{
        Addr:    fmt.Sprintf(":%s", port),
        Handler: s.mux,
    }

    serverCtx, serverStopCtx := context.WithCancel(context.Background())

    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        <-sig

        shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
        defer cancel()

        go func() {
            <-shutdownCtx.Done()
            if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
                log.Print("graceful shutdown timed out.. forcing exit")
            }
        }()

        if err := s.shutdown(shutdownCtx); err != nil {
            log.Printf("error during shutdown: %v", err)
        }
        serverStopCtx()
    }()

    log.Printf("server starting on port %s", port)
    if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        return fmt.Errorf("error starting server: %w", err)
    }

    s.wg.Wait()
    return nil
}`

	shutdownStr := `
func (s *Server) shutdown(ctx context.Context) error {
    if err := s.srv.Shutdown(ctx); err != nil {
        return fmt.Errorf("error shutting down http server: %w", err)
    }

    if err := s.app.Shutdown(); err != nil {
        return fmt.Errorf("error shutting down application: %w", err)
    }

    return nil
}`

	serverContent := fmt.Sprintf(`package server

import (
    "context"
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "%s/internal/app"
    "%s/internal/infrastructure/config"
    "%s/internal/core/services/user_service/service"
    "%s/internal/app/routes/api/users"
)
%s
%s
%s
%s
%s
%s
%s
%s`,
		projectName, projectName, projectName, projectName,
		serverStructStr,
		middlewareTypeStr,
		newServerStr,
		routeGroupStr,
		routeHandlerStr,
		methodsStr,
		runServerStr,
		shutdownStr)

	serverDir := filepath.Join(projectPath, "internal", "infrastructure", "server")
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %v", err)
	}

	serverPath := filepath.Join(serverDir, "server.go")
	if err := os.WriteFile(serverPath, []byte(serverContent), 0644); err != nil {
		return fmt.Errorf("failed to write server.go: %v", err)
	}

	return nil
}
