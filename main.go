package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/Medium/brigade/backend"
	"github.com/Medium/brigade/frontend"
)

var (
	bucket = kingpin.Arg("bucket", "S3 bucket name").Required().String()
	dev    = kingpin.Flag("dev", "Use development logging").Short('d').Bool()
	listen = kingpin.Flag("listen", "[address:]port to listen on").Default("8080").Short('l').String()
	region = kingpin.Flag("region", "AWS region the bucket is in").OverrideDefaultFromEnvar("AWS_DEFAULT_REGION").Short('r').String()
)

func main() {
	kingpin.CommandLine.Name = "brigade"
	kingpin.CommandLine.DefaultEnvars()
	kingpin.Parse()

	var logger *zap.Logger
	var err error
	if *dev {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	if !strings.Contains(*listen, ":") {
		*listen = fmt.Sprintf(":%s", *listen)
	}

	var sess *session.Session

	if *region != "" && *region != "default" {
		sess = session.Must(session.NewSession(&aws.Config{
			Region: region,
		}))
	} else {
		sess = session.Must(session.NewSession())
	}

	storage := backend.NewStorage(s3.New(sess), *bucket)
	server := frontend.Server{Backend: storage, Logger: logger}

	logger.Info("Listening for requests", zap.String("address", *listen))
	http.ListenAndServe(*listen, &server)
}
