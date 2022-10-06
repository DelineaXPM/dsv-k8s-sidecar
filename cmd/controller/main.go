package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/auth"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/pods"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/secrets"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	clientCredentials = "client_credentials"
)

var (
	keyDir     = util.EnvString("KEY_DIR", "/tmp/keys/")
	serverCert = keyDir + util.EnvString("SERVER_CRT", "cert.pem")
	serverKey  = keyDir + util.EnvString("SERVER_KEY", "key.pem")

	serverTokenCert = keyDir + util.EnvString("SERVER_CRT", "certtoken.pem")
	serverTokenKey  = keyDir + util.EnvString("SERVER_KEY", "keytoken.pem")
)

func main() {
	var tenant, clientId, clientSecret, port, authport, authType, logLevel string

	flag.StringVar(&tenant, "tenant", "", "Required tenant name")
	flag.StringVar(&authType, "auth-type", "", "Required authtype (client_credentials, certificate)")
	flag.StringVar(&clientId, "client-id", "", "Client credential id Required, if authtype client_credentials")
	flag.StringVar(&clientSecret, "client-secret", "", "Client credential proxy, Required if authtype client_credentials")
	flag.StringVar(&port, "port", ":3000", "Port to run on")
	flag.StringVar(&authport, "auth-port", ":8080", "Auth Port to run on")
	flag.StringVar(&logLevel, "log-level", "error", "Log Levels: panic,fatal,error,warn,info,debug,trace")
	flag.Parse()

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.ErrorLevel
	}
	log.SetLevel(level)

	if tenant == "" {
		log.Error("Required flags tenant is missing")
		os.Exit(2)
	}

	if strings.ToLower(authType) == clientCredentials && (clientId == "" || clientSecret == "") {
		log.Error("Required flags (client-id, client-secret) are missing")
		os.Exit(2)
	}

	secretClient := secrets.CreateSecretClient(tenant, clientId, clientSecret, authType)
	secretServer := secrets.NewSecretServer(secretClient)

	registry := pods.NewPodRegistry(tenant, os.Getenv("SIDECAR_NAMESPACE"))
	authService := auth.NewAuthService(envString("AUTH_SECRET", "Secret"), registry)
	authHandler := auth.NewAuthHandler(authService)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Panic("Failed to listen: ", err.Error())
	}

	var grpcServer *grpc.Server
	if _, err = os.Stat(serverCert); err != nil {
		log.Warn("Failed to get certificate keys: starting the server over TCP: ...", err.Error())
		grpcServer = grpc.NewServer(authService.GetUnaryInterceptor())
	} else {
		log.Info("starting with tls ...")
		// Create the TLS credentials
		creds, err := credentials.NewServerTLSFromFile(serverCert, serverKey)
		if err != nil {
			log.Panic("Failed to listen: ", err.Error())
		}
		grpcServer = grpc.NewServer(grpc.Creds(creds), authService.GetUnaryInterceptor())
	}

	secrets.RegisterDsvServer(grpcServer, secretServer)
	errs := make(chan error, 1)
	go func() {
		log.Info("Listening on port " + port)
		errs <- grpcServer.Serve(lis)
	}()
	go func() {
		router := mux.NewRouter().StrictSlash(true)
		router.HandleFunc("/auth", authHandler.GetToken).Methods("POST")
		http.Handle("/", router)
		if _, err := os.Stat(serverTokenCert); err != nil {
			log.Info("Auth Listening on port over TCP" + authport)
			errs <- http.ListenAndServe(authport, nil)
			os.Exit(1)
		}
		log.Info("Auth Listening on port TLS" + authport)
		errs <- http.ListenAndServeTLS(":443", serverTokenCert, serverTokenKey, nil)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	log.Infof("terminated %s", <-errs)
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
