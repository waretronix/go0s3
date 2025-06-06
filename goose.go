package goose

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func Upload(staticOptions [5]string, bucketName string, filePath string, objectName string, domainrewrite bool, cleanup bool) string {

	endpoint := staticOptions[0]
	accessKeyID := staticOptions[1]
	secretAccessKey := staticOptions[2]
	location := staticOptions[3]
	domain := staticOptions[4]

	ctx := context.Background()
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Check if bucket exists
	//if so, we'll upload to this bucket, if not - create new
	if bucketName != "" {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
		if err != nil {
			// Check to see if we already own this bucket (which happens if you run this twice)
			exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				log.Printf("We already own %s\n", bucketName)
			} else {
				log.Fatalln(err)
			}
		} else {
			log.Printf("Successfully created %s\n", bucketName)
		}
	} else {
		log.Fatalln("Empty bucket name")
	}

	if objectName == "" {
		object := strings.Split(filePath, ".")
		id := uuid.New()
		objectName = id.String() + "." + object[len(object)-1]
	}
	contentType := "application/octet-stream"

	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	
	res := "https://" + endpoint + "/" + bucketName + "/" + objectName
	if domainrewrite {
		res = "https://" + domain + "/" + objectName
	}
	
	//cleanup files if needed
	if cleanup {
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Failed to delete %s", filePath)
		}
	}

	log.Printf("Uploaded to %s", res)
	return res
}
