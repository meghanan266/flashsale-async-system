package services

import (
    "encoding/json"
    "fmt"
    "os"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sns"
    "github.com/aws/aws-sdk-go/service/sqs"
)

type AWSServices struct {
    SNS         *sns.SNS
    SQS         *sqs.SQS
    TopicArn    string
    QueueURL    string
}

func NewAWSServices() (*AWSServices, error) {
    // Create AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"),
    })
    if err != nil {
        return nil, err
    }
    
    // Get SNS topic ARN and SQS queue URL from environment variables
    topicArn := os.Getenv("SNS_TOPIC_ARN")
    queueURL := os.Getenv("SQS_QUEUE_URL")
    
    if topicArn == "" || queueURL == "" {
        return nil, fmt.Errorf("SNS_TOPIC_ARN and SQS_QUEUE_URL environment variables must be set")
    }
    
    return &AWSServices{
        SNS:      sns.New(sess),
        SQS:      sqs.New(sess),
        TopicArn: topicArn,
        QueueURL: queueURL,
    }, nil
}

func (a *AWSServices) PublishOrder(order interface{}) error {
    // Marshal order to JSON
    orderJSON, err := json.Marshal(order)
    if err != nil {
        return fmt.Errorf("failed to marshal order: %v", err)
    }
    
    // Publish to SNS
    _, err = a.SNS.Publish(&sns.PublishInput{
        Message:  aws.String(string(orderJSON)),
        TopicArn: aws.String(a.TopicArn),
    })
    
    if err != nil {
        return fmt.Errorf("failed to publish to SNS: %v", err)
    }
    
    return nil
}

func (a *AWSServices) ReceiveMessages(maxMessages int64) ([]*sqs.Message, error) {
    result, err := a.SQS.ReceiveMessage(&sqs.ReceiveMessageInput{
        QueueUrl:            aws.String(a.QueueURL),
        MaxNumberOfMessages: aws.Int64(maxMessages),
        WaitTimeSeconds:     aws.Int64(20), // Long polling
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to receive messages: %v", err)
    }
    
    return result.Messages, nil
}

func (a *AWSServices) DeleteMessage(receiptHandle string) error {
    _, err := a.SQS.DeleteMessage(&sqs.DeleteMessageInput{
        QueueUrl:      aws.String(a.QueueURL),
        ReceiptHandle: aws.String(receiptHandle),
    })
    
    return err
}