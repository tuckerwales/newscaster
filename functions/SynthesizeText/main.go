package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/aws/aws-sdk-go/service/s3"
)

func downloadText(bucketName string, objectKey string) (string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return "An error occurred.", err
	}

	var s3Client = s3.New(sess)

	object, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return "An error occurred.", err
	}

	contents, err := io.ReadAll(object.Body)
	if err != nil {
		return "An error occurred.", err
	}

	return string(contents), nil
}

func handleRequest(ctx context.Context, event events.S3Event) (string, error) {
	mediaBucket, mediaBucketPresent := os.LookupEnv("MEDIA_BUCKET_NAME")
	if !mediaBucketPresent {
		return "Invalid configuration.", nil
	}

	sess, err := session.NewSession()
	if err != nil {
		return "An error occurred.", err
	}

	var pollyClient = polly.New(sess)

	for _, record := range event.Records {
		log.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, record.S3.Bucket.Name, record.S3.Object.Key)

		contents, err := downloadText(record.S3.Bucket.Name, record.S3.Object.Key)
		if err != nil {
			return "An error occurred.", err
		}

		output, err := pollyClient.StartSpeechSynthesisTask(&polly.StartSpeechSynthesisTaskInput{
			Engine:             aws.String("neural"),
			LanguageCode:       aws.String("en-GB"),
			OutputFormat:       aws.String("mp3"),
			OutputS3BucketName: aws.String(mediaBucket),
			Text:               aws.String(contents),
			VoiceId:            aws.String("Brian"),
		})
		if err != nil {
			return "An error occurred.", err
		}

		log.Printf("Polly output: %s", output.String())
	}

	return "Hello, world!", nil
}

func main() {
	runtime.Start(handleRequest)
}
