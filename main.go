package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"log"
	"net/smtp"
	"os"
	"pigeon/constants"
	"time"
)

type S3Event struct {
	Records []struct {
		S3 struct {
			Bucket struct {
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func sendEmail(to, subject, body string) error {
	smtpServer := os.Getenv(constants.SmtpHost)
	port := os.Getenv(constants.SmtpPort)
	user := os.Getenv(constants.SmtpUser) // Substitua pelo seu e-mail do Gmail
	pass := os.Getenv(constants.SmtpPass) // Substitua pela sua senha do Gmail ou senha de aplicativo

	from := user
	msg := []byte("From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body)

	auth := smtp.PlainAuth("", user, pass, smtpServer)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpServer,
	}

	conn, err := tls.Dial("tcp", smtpServer+":"+port, tlsConfig)
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor SMTP:", err)
		return err
	}

	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		fmt.Println("Erro ao criar cliente SMTP:", err)
		return err

	}

	if err := client.Auth(auth); err != nil {
		fmt.Println("Erro na autenticação:", err)
		return err

	}

	if err := client.Mail(from); err != nil {
		fmt.Println("Erro ao definir remetente:", err)
		return err

	}

	if err := client.Rcpt(to); err != nil {
		fmt.Println("Erro ao definir destinatário:", err)
		return err

	}

	wc, err := client.Data()
	if err != nil {
		fmt.Println("Erro ao obter writer:", err)
		return err

	}

	_, err = wc.Write(msg)
	if err != nil {
		fmt.Println("Erro ao escrever corpo do e-mail:", err)
		return err
	}

	err = wc.Close()
	if err != nil {
		fmt.Println("Erro ao fechar writer:", err)
		return err

	}

	if err := client.Quit(); err != nil {
		fmt.Println("Erro ao encerrar conexão:", err)
		return err

	}

	fmt.Println("E-mail enviado com sucesso!")
	//err = smtp.SendMail(smtpServer+":"+port, auth, from, []string{to}, msg)
	return nil
}

func sendEmailTwo(to, subject, body string) error {
	smtpServer := os.Getenv(constants.SmtpHost)
	//port := os.Getenv(constants.SmtpPort)
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
func decryptEmail(encryptedEmail string, key []byte) (string, error) {
	ciphertext, _ := base64.StdEncoding.DecodeString(encryptedEmail)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
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
	// Configuração da sessão AWS
	sess, err := NewAwsClient()
	if err != nil {
		fmt.Println("Erro ao criar sessão AWS:", err)
		return "", err
	}

	// Criação do cliente STS
	stsSvc := sts.New(sess)

	// Assumindo um papel IAM para obter credenciais temporárias
	creds := stscreds.NewCredentialsWithClient(stsSvc, os.Getenv("ROLE_ARN"))

	// Criação do cliente S3 com as credenciais temporárias
	s3Svc := s3.New(sess, &aws.Config{Credentials: creds})

	// Parâmetros para a URL pré-assinada
	req, _ := s3Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})

	// Gerando a URL pré-assinada
	urlStr, err := req.Presign(48 * time.Hour) // URL válida por 48 Hours
	if err != nil {
		fmt.Println("Erro ao gerar URL pré-assinada:", err)
		return "", err
	}

	fmt.Println("URL Pré-assinada:", urlStr)
	return "", nil
}

func HandleRequest(ctx context.Context, s3Event events.S3Event) (string, error) {
	sess, _ := NewAwsClient()
	s3Client := s3.New(sess)

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		headObjectInput := &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}
		headObjectOutput, err := s3Client.HeadObject(headObjectInput)
		if err != nil {
			return "", err
		}

		encryptedEmail := headObjectOutput.Metadata[os.Getenv(constants.HeadMetadata)]

		// Imprime os metadados do objeto
		fmt.Printf("Metadados do objeto %s/%s:\n", bucket, key)
		fmt.Printf("Tipo de conteúdo: %s\n", *headObjectOutput.ContentType)
		fmt.Printf("Tamanho: %d bytes\n", headObjectOutput.ContentLength)
		fmt.Printf("Última modificação: %v\n", headObjectOutput.LastModified)
		for key, value := range headObjectOutput.Metadata {
			fmt.Printf("Chave: %s, Valor: %v\n", key, *value)
		}
		fmt.Printf("E-mail criptografado: %s\n", *encryptedEmail)
		if encryptedEmail != nil && *encryptedEmail != "" {
			//decryptedEmail, err := decryptEmail(*encryptedEmail, []byte(os.Getenv(secretKey))) // Substitua "YOUR_ENCRYPTION_KEY" pela sua chave de criptografia
			//if err != nil {
			//	return "", err
			//}

			//downloadLink := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
			downloadLink, err := GetUrl(bucket, key)
			if err != nil {
				fmt.Println("Erro ao obter URL pré-assinada:", err)
				continue
			}

			message := fmt.Sprintf("Um novo arquivo foi carregado: %s", downloadLink)
			subject := "Novo arquivo carregado"

			err = sendEmailTwo(*encryptedEmail, subject, message)
			if err != nil {
				fmt.Println("Erro ao enviar email", err)
				continue
			}
		}
	}

	return "Notificação enviada com sucesso!", nil
}

func main() {
	lambda.Start(HandleRequest)
}
