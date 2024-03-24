package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"net/smtp"
	"os"
	"pigeon/constants"
	"time"
)

func sendEmail(to, subject, body string) error {
	smtpServer := os.Getenv(constants.SmtpHost)
	user := os.Getenv(constants.SmtpUser)
	pass := os.Getenv(constants.SmtpPass)
	smtpPort := "587" // Porta para conexão TLS

	from := user
	password := pass

	message := []byte("From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body)

	auth := smtp.PlainAuth("", from, password, smtpServer)

	client, err := smtp.Dial(smtpServer + ":" + smtpPort)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor SMTP:", err)
		return err
	}
	defer func(client *smtp.Client) {
		err := client.Close()
		if err != nil {
			fmt.Println("Erro ao fechar conexão:", err)
		}
	}(client)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err := client.Auth(auth); err != nil {
		return err
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	_, err = writer.Write(message)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	fmt.Println("Email enviado com sucesso!")
	return nil
}

func NewAwsClient() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv(constants.RegionAws))},
	)
	if err != nil {
		log.Fatal(err)
		return nil, err

	}
	return sess, nil

}

func GetUrl(bucketName, objectKey string) (string, error) {
	sess, err := NewAwsClient()
	if err != nil {
		fmt.Println("Erro ao criar sessão AWS:", err)
		return "", err
	}

	s3Svc := s3.New(sess)

	req, _ := s3Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})

	urlStr, err := req.Presign(30 * time.Minute)
	if err != nil {
		fmt.Println("Erro ao gerar URL pré-assinada:", err)
		return "", err
	}

	fmt.Println("URL Pré-assinada:", urlStr)
	return urlStr, nil
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) (string, error) {
	sess, _ := NewAwsClient()
	s3Client := s3.New(sess)

	//fmt.Printf("Evento recebido: %v\n", s3Event)
	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.URLDecodedKey

		headObjectInput := &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}
		headObjectOutput, err := s3Client.HeadObject(headObjectInput)
		if err != nil {
			fmt.Println("Erro ao obter metadados do objeto:", err)
			return "", err
		}

		emailAddress := headObjectOutput.Metadata[os.Getenv(constants.HeadMetadata)]

		//fmt.Printf("Metadados do objeto %s/%s:\n", bucket, key)
		//fmt.Printf("Tipo de conteúdo: %s\n", *headObjectOutput.ContentType)
		//fmt.Printf("Tamanho: %d bytes\n", headObjectOutput.ContentLength)
		//fmt.Printf("Última modificação: %v\n", headObjectOutput.LastModified)
		for key, value := range headObjectOutput.Metadata {
			fmt.Printf("Chave: %s, Valor: %v\n", key, *value)
		}

		if emailAddress != nil && *emailAddress != "" {

			downloadLink, err := GetUrl(bucket, key)
			if err != nil {
				fmt.Println("Erro ao obter URL pré-assinada:", err)
				continue
			}

			message := fmt.Sprintf("Um novo relatorio foi gerado: \n %s \n"+
				"Estes link expirará em 30 minutos.", downloadLink)
			subject := "Relatario de Batidas de Pontos"

			err = sendEmail(*emailAddress, subject, message)
			if err != nil {
				fmt.Println("Erro ao enviar email", err)
				continue
			}
		} else {
			fmt.Printf("E-mail não encontrado no metadado para a chave %s", key)
		}
	}

	return "Notificação enviada com sucesso!", nil
}

func main() {
	lambda.Start(HandleRequest)
}
