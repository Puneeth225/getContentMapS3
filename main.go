package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type BucketBasics struct {
	S3Client *s3.Client
}

const (
	AWS_S3_REGION = "us-east-1" // Region
)

var nameToContent = make(map[string]string)
var name string
var st []string

// DownloadFile gets an object from a bucket and stores it in a local file.
func (basics BucketBasics) DownloadFile(bucketName string, objectKey string) (string, error) {
	result, err := basics.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return "", err
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	ans := string(body)
	return ans, nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	name = r.URL.Path[12:]
	var f bool
	_, ok := nameToContent[name]
	if ok {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "")
		fmt.Fprintf(w, "%s", nameToContent[name])
		f = true
	}

	if f == false {
		fmt.Fprintf(w, "The file is not present in bucket!!!!OOPS!!!!")
	}
	return
}

func showfiles(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Files in bucket are: ")
	for i := 0; i < len(st); i++ {
		fmt.Fprintf(w, "%s", st[i])
		fmt.Fprintf(w, "\n")
	}
}

func main() {

	// Load the SDK's configuration from environment and shared config, and
	// create the client with this.
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile("personal-puneeth"),
		config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	output, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String("test-puneeth"),
	})
	for _, obj := range output.Contents {
		st = append(st, *obj.Key)
	}
	bucketBasics := BucketBasics{S3Client: s3Client}

	for j := 0; j < len(st); j++ {
		cnt, _ := bucketBasics.DownloadFile("test-puneeth", st[j])
		nameToContent[st[j]] = cnt
	}
	handler := http.HandlerFunc(handleRequest)
	http.Handle("/averlon/s3/", handler)
	http.HandleFunc("/averlon/s3", showfiles)
	http.ListenAndServe(":8080", nil)
}
