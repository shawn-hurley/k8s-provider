package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/konveyor/k8s-provider/pkg/client"
	"github.com/sirupsen/logrus"
)

var (
	port       = flag.Int("port", 0, "Port must be set")
	initConfig = flag.String("initConfig", "settings.json", "Pass in the init config")
)

func main() {

	flag.Parse()
	logrusLog := logrus.New()
	logrusLog.SetOutput(os.Stdout)
	logrusLog.SetFormatter(&logrus.TextFormatter{})
	// need to do research on mapping in logrusr to level here TODO
	logrusLog.SetLevel(logrus.Level(5))

	log := logrusr.New(logrusLog)

	client := client.NewK8SProvider()
	if initConfig != nil {
		s := *initConfig
		runCLI(s, client, log)
		return
	}

	if port == nil || *port == 0 {
		panic(fmt.Errorf("must pass in the port for the external provider"))
	}

	s := provider.NewServer(client, *port, log)
	ctx := context.TODO()
	s.Start(ctx)
}

// TODO: Fix the panics
func runCLI(initConfigLocation string, client provider.BaseClient, log logr.Logger) {
	//get the init config
	bytes, err := os.ReadFile(initConfigLocation)
	if err != nil {
		panic(err)
	}
	initConfig := &provider.InitConfig{}
	json.Unmarshal(bytes, initConfig)

	serviceClient, err := client.Init(context.Background(), log, *initConfig)
	if err != nil {
		panic(err)
	}

	b := []byte("path: spec.template.spec.containers[] | select(.livenessProbe)\nresource: apps v1 Deployment")

	response, err := serviceClient.Evaluate(context.Background(), "k8s-resource-path", b)
	if err != nil {
		panic(err)
	}
	log.Info("Got Response", "response", response)
}
