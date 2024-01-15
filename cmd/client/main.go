package main

import (
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

const controllerServiceName = "dsv-k8s-controller.%s:80"

var keyDir = util.EnvString("KEY_DIR", "/tmp/keys/") //nolint:gochecknoglobals // no other possibility

func main() {

	logLevel := os.Getenv("LOG_LEVEL")
	name := os.Getenv("POD_NAME")
	namespace := os.Getenv("POD_NAMESPACE")
	podIP := os.Getenv("POD_IP")
	brokerNamespace := os.Getenv("BROKER_NAMESPACE")

	level, lErr := log.ParseLevel(logLevel)
	if lErr != nil {
		level = log.ErrorLevel
	}
	log.SetLevel(level)

	if brokerNamespace == "" {
		brokerNamespace = namespace
	}

	log.WithFields(log.Fields{
		"podName":         name,
		"podNamespace":    namespace,
		"podIp":           podIP,
		"brokerNamespace": brokerNamespace,
	}).Info("Client started")

	var (
		token credentials.PerRPCCredentials
		err   error
	)

	serverCert := keyDir + util.EnvString("SERVER_CRT", "cert.pem")
	if _, err = os.Stat(serverCert); err != nil {
		log.Info("Connecting over TCP with token")
		token, err = auth.GetToken(namespace+"/"+name, podIP, brokerNamespace)
	} else {
		log.Info("Connecting with TLS with token")
		token, err = auth.GetTLsToken(namespace+"/"+name, podIP, brokerNamespace, serverCert)
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

	creds, err := credentials.NewClientTLSFromFile(keyDir+util.EnvString("SERVER_CRT", "cert.pem"), "")
	if err != nil {
		log.Warn("Failed to get certificate keys: starting the server insecure: ", err.Error())
		conn, err = grpc.Dial(url, grpc.WithInsecure(), grpc.WithPerRPCCredentials(token))
	} else {
		conn, err = grpc.Dial(url, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(token))
	}

	if err != nil {
		log.WithFields(log.Fields{"url": url, "error": err.Error()}).Fatal("did not connect")
	}

	return conn
}
