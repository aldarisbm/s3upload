package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	S3_BUCKET = ""
	S3_REGION = ""
)

func main() {
	s3Bucket := flag.String(
		"bucketname",
		"",
		"Name of s3 bucket to upload to")
	pathToFile := flag.String(
		"path",
		"",
		"Path to the file to upload, key in s3 will the basepath")
	s3Region := flag.String(
		"region",
		"us-east-1",
		"AWS region for the bucket location")
	flag.Parse()

	S3_REGION = *s3Region
	S3_BUCKET = *s3Bucket
	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.Fatal(err)
	}
	output, err := PutFileInS3(s, *pathToFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(output)
}

// PutFileInS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func PutFileInS3(s *session.Session, fileDir string) (string, error) {
	file, err := os.Open(fileDir)
	if err != nil {
		return "", err
	}
	defer file.Close()
	fileKey := filepath.Base(fileDir)
	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	output, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(S3_BUCKET),
		Key:                  aws.String(fileKey),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	return fmt.Sprintf("%s", *output), err
}
