package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

type Order struct {
    OrderID    string    `json:"order_id"`
    CustomerID int       `json:"customer_id"`
    Status     string    `json:"status"`
    CreatedAt  time.Time `json:"created_at"`
}

func HandleSNSEvent(ctx context.Context, snsEvent events.SNSEvent) error {
    for _, record := range snsEvent.Records {
        // Parse the order from SNS message
        var order Order
        if err := json.Unmarshal([]byte(record.SNS.Message), &order); err != nil {
            return fmt.Errorf("error parsing order: %v", err)
        }
        
        fmt.Printf("Processing order %s for customer %d\n", order.OrderID, order.CustomerID)
        
        // Simulate payment processing (3 seconds)
        time.Sleep(3 * time.Second)
        
        fmt.Printf("Successfully processed order %s\n", order.OrderID)
    }
    
    return nil
}

func main() {
    lambda.Start(HandleSNSEvent)
}