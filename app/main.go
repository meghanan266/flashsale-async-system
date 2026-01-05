package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "flash-sale-app/handlers"
    "flash-sale-app/services"
)

func main() {
    // Get port from environment variable or default to 8080
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Initialize payment processor with concurrency limit
    paymentProcessor := services.NewPaymentProcessor(1)
    
    // Initialize AWS services (SNS/SQS)
    awsServices, err := services.NewAWSServices()
    if err != nil {
        log.Printf("Warning: AWS services not initialized: %v", err)
        log.Printf("Async endpoint will not work without SNS_TOPIC_ARN and SQS_QUEUE_URL environment variables")
        // Continue without AWS services for local testing
    }
    
    // Initialize handlers
    orderHandler := handlers.NewOrderHandler(paymentProcessor, awsServices)

    // Setup routes
    http.HandleFunc("/orders/sync", orderHandler.HandleSyncOrder)
    http.HandleFunc("/orders/async", orderHandler.HandleAsyncOrder)
    http.HandleFunc("/health", handlers.HealthCheck)

    // Start server
    fmt.Printf("Server starting on port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}