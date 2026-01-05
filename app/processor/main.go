package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"
    "sync"
    "time"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
)

type SQSMessage struct {
    Message string `json:"Message"`
}

type Order struct {
    OrderID    string    `json:"order_id"`
    CustomerID int       `json:"customer_id"`
    Status     string    `json:"status"`
    CreatedAt  time.Time `json:"created_at"`
}

func main() {
    // Get configuration from environment
    queueURL := os.Getenv("SQS_QUEUE_URL")
    if queueURL == "" {
        log.Fatal("SQS_QUEUE_URL environment variable is required")
    }
    
    workerCount := 1
    if wc := os.Getenv("WORKER_COUNT"); wc != "" {
        workerCount, _ = strconv.Atoi(wc)
    }
    
    fmt.Printf("Starting order processor with %d workers\n", workerCount)
    fmt.Printf("Queue URL: %s\n", queueURL)
    
    // Create AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"),
    })
    if err != nil {
        log.Fatalf("Failed to create AWS session: %v", err)
    }
    
    sqsClient := sqs.New(sess)
    
    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            processMessages(workerID, sqsClient, queueURL)
        }(i)
    }
    
    wg.Wait()
}

func processMessages(workerID int, sqsClient *sqs.SQS, queueURL string) {
    fmt.Printf("Worker %d started\n", workerID)
    
    for {
        // Receive messages from SQS
        result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
            QueueUrl:            aws.String(queueURL),
            MaxNumberOfMessages: aws.Int64(10),
            WaitTimeSeconds:     aws.Int64(20),
        })
        
        if err != nil {
            log.Printf("Worker %d: Error receiving messages: %v", workerID, err)
            time.Sleep(5 * time.Second)
            continue
        }
        
        if len(result.Messages) == 0 {
            continue
        }
        
        fmt.Printf("Worker %d: Processing %d messages\n", workerID, len(result.Messages))
        
        for _, msg := range result.Messages {
            // Parse SNS message wrapper
            var snsMsg SQSMessage
            if err := json.Unmarshal([]byte(*msg.Body), &snsMsg); err != nil {
                log.Printf("Worker %d: Error parsing SNS message: %v", workerID, err)
                continue
            }
            
            // Parse actual order
            var order Order
            if err := json.Unmarshal([]byte(snsMsg.Message), &order); err != nil {
                log.Printf("Worker %d: Error parsing order: %v", workerID, err)
                continue
            }
            
            // Simulate payment processing (3 seconds)
            startTime := time.Now()
            time.Sleep(3 * time.Second)
            processingTime := time.Since(startTime)
            
            fmt.Printf("Worker %d: Processed order %s in %dms\n", 
                workerID, order.OrderID, processingTime.Milliseconds())
            
            // Delete message from queue
            _, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
                QueueUrl:      aws.String(queueURL),
                ReceiptHandle: msg.ReceiptHandle,
            })
            if err != nil {
                log.Printf("Worker %d: Error deleting message: %v", workerID, err)
            }
        }
    }
}