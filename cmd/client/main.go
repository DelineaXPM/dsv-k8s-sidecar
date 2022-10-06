package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/auth"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/env"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/secrets"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/util"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const controllerServiceName = "dsv-broker.%s:80"

var (
	keyDir          = util.EnvString("KEY_DIR", "/tmp/keys/")
	serverCert      = keyDir + util.EnvString("SERVER_CRT", "ca.pem")
	serverTokenCert = keyDir + util.EnvString("SERVER_CRT", "catoken.pem")
)

func main() {
	var logLevel string
	flag.StringVar(&logLevel, "log-level", "error", "Log Levels: panic,fatal,error,warn,info,debug,trace")
	flag.Parse()

	level, lErr := log.ParseLevel(logLevel)
	if lErr != nil {
		level = log.ErrorLevel
	}
	log.SetLevel(level)

	name := os.Getenv("POD_NAME")
	namespace := os.Getenv("POD_NAMESPACE")
	ip := os.Getenv("POD_IP")
	brokerNamespace := os.Getenv("BROKER_NAMESPACE")

	if brokerNamespace == "" {
		brokerNamespace = namespace
	}

	log.WithFields(log.Fields{
		"podName":         name,
		"podNamespace":    namespace,
		"podIp":           ip,
		"brokerNamespace": brokerNamespace,
	}).Info("Client started")

	var (
		token credentials.PerRPCCredentials
		err   error
	)

	if _, err = os.Stat(serverTokenCert); err != nil {
		log.Info("Connecting over TCP: token")
		token, err = auth.GetToken(namespace+"/"+name, ip, brokerNamespace)
	} else {
		log.Info("Connecting with TLS: token")
		token, err = auth.GetTLsToken(namespace+"/"+name, ip, brokerNamespace, serverTokenCert)
	}
	if err != nil {
		log.Fatalf("Unable to get token: %s", err)
	}

	grpcConn := getGRPCConnection(token, brokerNamespace)
	defer grpcConn.Close()

	client := secrets.NewDsvClient(grpcConn)
	agent := env.CreateEnvironmentAgent(client)

	defer agent.Close()

	errs := agent.Run()

	log.Infof("terminated %s", <-errs)
}

func getGRPCConnection(token credentials.PerRPCCredentials, brokerNamespace string) *grpc.ClientConn {
	url := fmt.Sprintf(controllerServiceName, brokerNamespace)
	var (
		conn *grpc.ClientConn
		err  error
	)

	creds, err := credentials.NewClientTLSFromFile(serverCert, "")
	if err != nil {
		log.Warn("Failed to get certificate keys: starting the server unsecure: ....", err.Error())
		conn, err = grpc.Dial(url, grpc.WithInsecure(), grpc.WithPerRPCCredentials(token))
	} else {
		conn, err = grpc.Dial(url, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(token), grpc.WithWaitForHandshake())
	}

	if err != nil {
		log.WithFields(log.Fields{"url": url, "error": err.Error()}).Fatal("did not connect")
	}

	return conn
}
