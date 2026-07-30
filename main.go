package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"string"

	"k8s.io/client-go/informers"

	"github.com/keepchasingheaven/encore-workspaces-proxy/internal/logz"
	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/auth"
	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/config"
	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/encore"
	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/k8s"
	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/logging"
	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/server"
	"github.com/keepchasingheaven/encore-workspaces-proxy/pkg/upstream"
	
	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	workspaceAgentIDLabel = "agent.encore.com/id"
)

func main() { // nolint:cyclop
	cfg := config.Config{}

	var kubeconfig string
	var watchNamespaces string

	flag.StringVar(&cfg.Auth.ClientID, "auth-client-id", getEnvString("AUTH_CLIENT_ID", ""), "Auth Client ID (env: AUTH_CLIENT_ID)")
	flag.StringVar(&cfg.Auth.ClientSecret, "auth-client-secret", getEnvString("AUTH_CLIENT_SECRET", ""), "Auth Client Secret (env: AUTH_CLIENT_SECRET)")
	flag.StringVar(&cfg.Auth.RedirectURI, "auth-redirect-uri", getEnvString("AUTH_REDIRECT_URI", ""), "Auth Redirect URI (env: AUTH_REDIRECT_URI)")
	flag.StringVar(&cfg.Auth.Host, "auth-host", getEnvString("AUTH_HOST", ""), "Auth Host (env: AUTH_HOST)")
	flag.StringVar(&cfg.Auth.SigningKey, "auth-signing-key", getEnvString("AUTH_SIGNING_KEY", ""), "Auth Signing Key (env: AUTH_SIGNING_KEY)")
	flag.StringVar(&cfg.Auth.Protocol, "auth-protocol", getEnvString("AUTH_PROTOCOL", "https"), "Auth Protocol. Acceptable values are http, https (env: AUTH_PROTOCOL)")
	flag.StringVar(&cfg.MetricsPath, "metrics-path", getEnvString("METRICS_PATH", "/metrics"), "Metrics Path (env: METRICS_PATH)")
	flag.StringVar(&cfg.LogLevel, "log-level", getEnvString("LOG_LEVEL", "info"), "Log level. Acceptable values are info, debug, warn, error (env: LOG_LEVEL)")
	flag.BoolVar(&cfg.HTTP.Enabled, "http-enabled", getEnvBool("HTTP_ENABLED", true), "HTTP traffic enabled (env: HTTP_ENABLED)")
	flag.IntVar(&cfg.HTTP.Port, "http-port", getEnvInt("HTTP_PORT", 9876), "HTTP traffic port (env: HTTP_PORT)")
	flag.BoolVar(&cfg.SSH.Enabled, "ssh-enabled", getEnvBool("SSH_ENABLED", true), "SSH traffic enabled (env: SSH_ENABLED)")
	flag.IntVar(&cfg.SSH.Port, "ssh-port", getEnvInt("SSH_PORT", 22), "SSH traffic port (env: SSH_PORT)")
	flag.StringVar(&cfg.SSH.HostKey, "ssh-host-key", getEnvString("SSH_HOST_KEY", ""), "SSH host key (env: SSH_HOST_KEY)")
	flag.IntVar(&cfg.SSH.BackendPort, "ssh-backend-port", getEnvInt("SSH_BACKEND_PORT", 60022), "SSH backend port (env: SSH_BACKEND_PORT)")
	flag.StringVar(&cfg.SSH.BackendUsername, "ssh-backend-username", getEnvString("SSH_BACKEND_USERNAME", "encore-workspaces"), "SSH backend username (env: SSH_BACKEND_USERNAME)")
	flag.StringVar(&watchNamespaces, "watch-namespaces", getEnvString("WATCH_NAMESPACES", ""), "Comma-separated list of namespaces to watch. Set to empty to watch all namespaces (env: WATCH_NAMESPACES)")
	flag.StringVar(&kubeconfig, "kubeconfig", getEnvString("KUBECONFIG", ""), "Kubernetes config file (env: KUBECONFIG)")

	flag.Parse()

	cfg.WatchNamespaces = strings.Split(watchNamespaces, ",")
	for i := range cfg.WatchNamespaces {
		cfg.WatchNamespaces[i] = strings.TrimSpace(cfg.WatchNamespaces[i])
	}

	err := cfg.Validate()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "falha ao validar a configuração: %s\n", err)
		
		os.Exit(-1)
	}

	ctx := context.Background()

	var zapLevel zap.AtomicLevel
	
	err = zapLevel.UnmarshalText([]byte(cfg.LogLevel))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "falha ao ler o nível do log: %s\n", err)

		os.Exit(-1)
	}
	
	logConfig := zap.NewProductionConfig()
	logConfig.Level = zapLevel

	logger, err := logConfig.Build()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "falha ao criar o logger: %s\n", err)
		
		os.Exit(-1)
	}

	defer func() {
		err = logger.Sync()
		
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "falha ao sincronizar o logger: %s\n", err)
			
			return
		}
	}()

	k8sClient, err := k8s.New(logger, kubeconfig)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "falha ao criar cliente kubernetes: %s\n", err)
		
		os.Exit(-1)
	}

	apiFactory := func(accessToken string) encore.API {
		return encore.NewClient(logger, accessToken, cfg.Auth.Host, encore.BearerTokenType)
	}

	upstreamTracker := upstream.NewTracker(logger)
	loggingMiddleware := logging.NewMiddleware(logger)
	authMiddleware := auth.NewMiddleware(logger, &cfg.Auth, upstreamTracker, apiFactory)

	opts := &server.Options{
		HTTPConfig:        cfg.HTTP,
		SSHConfig:         cfg.SSH,
		LoggingMiddleware: loggingMiddleware,
		AuthMiddleware:    authMiddleware,
		Logger:            logger,
		Tracker:           upstreamTracker,
		MetricsPath:       cfg.MetricsPath,
		APIFactory:        apiFactory
	}

	s := server.New(opts)
	svcTweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = workspaceAgentIDLabel
	}
	eventHandler := k8s.GetSvcInformerActionHandler(logger, upstreamTracker)

	informerOptions := []informers.SharedInformerOption{
		informers.WithTweakListOptions(svcTweakListOptions)
	}

	// se tivermos namespaces, iniciar um informante para cada um deles
	for _, namespace := range cfg.WatchNamespaces {
		logger.Info("iniciando informante de namespace", logz.WatchedNamespace(namespace))

		err = k8sClient.SubscribeToInformerEvents(ctx, eventHandler, append(informerOptions, informers.WithNamespace(namespace))...)
		if err != nil {
			logger.Error("falha ao iniciar informante", logz.Error(err), logz.WatchedNamespace(namespace))
			
			return
		}
	}

	// caso contrário, iniciar um informante para observar todos os namespaces
	if len(cfg.WatchNamespaces) == 0 {
		logger.Info("iniciando informante para todos os namespaces")

		err = k8sClient.SubscribeToInformerEvents(ctx, eventHandler, informerOptions...)
		if err != nil {
			logger.Error("falha ao iniciar informante", logz.Error(err))

			return
		}
	}

	err = s.Start(ctx)
	if err != nil {
		logger.Error("falha ao iniciar servidor", logz.Error(err))
	}
}

// getenvstring retorna o valor da variável de ambiente caso definida, caso
// contrário, retorna o valor padrão
func getEnvString(envVar, defaultValue string) string {
	if value, present := os.LookupEnv(envVar); present {
		return value
	}

	return defaultValue
}

// getenvint retorna o valor da variável de ambiente como int caso definida, caso
// contrário, retorna o valor padrão
func getEnvInt(envVar string, defaultValue int) int {
	if value, present := os.LookupEnv(envVar); present {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}

	return defaultValue
}

// getenvbool retorna o valor da variável de ambiente como bool caso definida, caso
// contrário, retorna o valor padrão
func getEnvBool(envVar string, defaultValue bool) bool {
	if value, present := os.LookupEnv(envVar); present {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}

	return defaultValue
}
