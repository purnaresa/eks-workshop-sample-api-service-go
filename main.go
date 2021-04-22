package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var svc *rekognition.Rekognition
var svcS3 *s3.S3

func init() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	})

	if err != nil {
		log.Fatalln("Error while creating session,", err)
		return
	}

	svc = rekognition.New(sess)
	svcS3 = s3.New(sess)
}

func main() {

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/", healthcheck)
	router.POST("/compare", ComparesFace)
	router.Run(":8080")
}

func healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to eKYC Demo 0.1"})
}

func ComparesFace(c *gin.Context) {

	fileSource, _ := c.FormFile("fileSource")
	fileTarget, _ := c.FormFile("fileTarget")

	fileSourceNameData, err := fileSource.Open()
	if err != nil {
		log.Println(err)
	}
	fileSourceName, errUpload := UploadFile(fileSource.Filename, fileSourceNameData)
	if err != nil {
		log.Println(errUpload)
	}

	fileTargetData, err := fileTarget.Open()
	if err != nil {
		log.Println(err)
	}
	fileTargetName, errUpload := UploadFile(fileTarget.Filename, fileTargetData)
	if err != nil {
		log.Println(errUpload)
	}

	inputCompare := &rekognition.CompareFacesInput{
		SimilarityThreshold: aws.Float64(90.000000),
		SourceImage: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String("kyc-test-11571204"),
				Name:   aws.String(fileSourceName),
			},
		},
		TargetImage: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String("kyc-test-11571204"),
				Name:   aws.String(fileTargetName),
			},
		},
	}

	result, err := svc.CompareFaces(inputCompare)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rekognition.ErrCodeInvalidParameterException:
				log.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
			case rekognition.ErrCodeInvalidS3ObjectException:
				log.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
			case rekognition.ErrCodeImageTooLargeException:
				log.Println(rekognition.ErrCodeImageTooLargeException, aerr.Error())
			case rekognition.ErrCodeAccessDeniedException:
				log.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
			case rekognition.ErrCodeInternalServerError:
				log.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
			case rekognition.ErrCodeThrottlingException:
				log.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
			case rekognition.ErrCodeProvisionedThroughputExceededException:
				log.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case rekognition.ErrCodeInvalidImageFormatException:
				log.Println(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
			c.JSON(http.StatusBadRequest, err.Error())
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func UploadFile(fileName string, file multipart.File) (newFileName string, err error) {
	guid := xid.New()

	newFileName = fmt.Sprintf("%s-%s", guid.String(), fileName)

	input := &s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String("kyc-test-11571204"),
		Key:    aws.String(newFileName),
	}

	_, err = svcS3.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return
	}

	return

}
