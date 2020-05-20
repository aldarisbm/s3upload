package main

import (
	"bytes"
	"errors"
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
		"Name of s3 bucket to upload to",
	)
	pathToFile := flag.String(
		"path",
		"",
		"Path to the file to XML file to upload, key in s3 will the basepath",
	)
	pathToFolder := flag.String(
		"folder",
		"",
		"Path to the folder containing XML files, key in s3 will be the basepath",
	)
	s3Region := flag.String(
		"region",
		"us-east-1",
		"AWS region for the bucket location",
	)
	flag.Parse()

	S3_REGION = *s3Region
	S3_BUCKET = *s3Bucket
	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.Fatal(err)
	}

	outputs, err := handleInput(*pathToFile, *pathToFolder, s)
	if err != nil {
		fmt.Errorf("Something went wrong")
		os.Exit(1)
	}

	for i, output := range outputs {
		fmt.Printf("File %d: %s", i, output)
	}
}

func handleInput(pathToFile, pathToFolder string, s *session.Session) ([]string, error) {
	var outputs []string

	switch {
	case pathToFile != "" && pathToFolder != "":
		return nil, errors.New("Provide a file path or a folder path, not both")
	case pathToFile != "":
		output, err := putFileInS3(s, pathToFile)
		if err != nil {
			return nil, err
		}
		return append(outputs, output), nil
	case pathToFolder != "":
		outputs, err := putFolderInS3(s, pathToFolder)
		if err != nil {
			return nil, err
		}
		return outputs, nil
	default:
		return nil, errors.New("Wrong input, refer to -h")
	}
}

// putFileInS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func putFileInS3(s *session.Session, fileDir string) (string, error) {
	if !isXML(fileDir) {
		return "", errors.New("Provide an XML file")
	}

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

// putFolderInS3 will upload multiple files to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func putFolderInS3(s *session.Session, fileDir string) ([]string, error) {
	var outputs []string
	fileNames, err := getFileNames(fileDir)
	if err != nil {
		return nil, errors.New("Error parsing the files inside the folder")
	}
	for _, fileName := range fileNames {
		output, err := putFolderInS3(s, fileName)
		if err != nil {
			return nil, fmt.Errorf("Error trying to retrieve: %s", fileName)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func getFileNames(fileDir string) ([]string, error) {

}

func isXML(fileDir string) bool {
	return filepath.Ext(fileDir) == ".xml" || filepath.Ext(fileDir) == ".XML"
}
