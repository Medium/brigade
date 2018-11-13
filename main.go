package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/Medium/brigade/backend"
	"github.com/Medium/brigade/frontend"
)

var (
	bucket = kingpin.Arg("bucket", "S3 bucket name").Required().String()
	listen = kingpin.Flag("listen", "[address:]port to listen on").Default("8080").Short('l').String()
	region = kingpin.Flag("region", "AWS region the bucket is in").OverrideDefaultFromEnvar("AWS_DEFAULT_REGION").Short('r').String()
)

func main() {
	kingpin.CommandLine.Name = "brigade"
	kingpin.CommandLine.DefaultEnvars()
	kingpin.Parse()

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
	server := frontend.Server{Backend: storage}

	http.ListenAndServe(*listen, &server)
}
