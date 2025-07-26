package db

import (
	"log"
	"os"
	"time"

	"github.com/couchbase/gocb/v2"
)

func InitialiseBucket() *gocb.Bucket {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatalf("error loading .env file: %s", err)
	// }

	connectionString := os.Getenv("DB_CONNECTION_STRING")
	bucketName := os.Getenv("DB_BUCKET_NAME")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")

	options := gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: time.Second * 10,
		},
		SecurityConfig: gocb.SecurityConfig{
			TLSSkipVerify: false,
		},
	}

	// Sets a pre-configured profile called "wan-development" to help avoid latency issues
	// when accessing Capella from a different Wide Area Network
	// or Availability Zone (e.g. your laptop).
	if err := options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment); err != nil {
		log.Fatal(err)
	}

	// Initialize the Connection
	cluster, err := gocb.Connect("couchbases://"+connectionString, options)
	if err != nil {
		log.Fatal(err)
	}

	bucket := cluster.Bucket(bucketName)
	err = bucket.WaitUntilReady(10*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	return bucket
}
