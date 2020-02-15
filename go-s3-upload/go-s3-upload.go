package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kelseyhightower/envconfig"
)

type Arguments struct {
	S3Bucket  string
	AWSRegion string
}

func AddFileToS3(s *session.Session, fileDir string) error {
	var a Arguments
	err := envconfig.Process("s3upload", &a)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Open the file for use
	file, err := os.Open(fileDir)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(a.S3Bucket),
		Key:                  aws.String(fileDir),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("STANDARD_IA"),
	})
	return err
}

func main() {
	var a Arguments
	err := envconfig.Process("s3upload", &a)
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(a.AWSRegion),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// walk file structure
	flag.Parse()
	root := flag.Arg(0)

	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		filename := root + "/" + file.Name()
		err := AddFileToS3(s, filename)
		if err != nil {
			log.Fatal(err)
		}
	}

}
