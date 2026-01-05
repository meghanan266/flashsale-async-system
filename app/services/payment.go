package services

import (
    "fmt"
    "time"
    "flash-sale-app/models"
)

// PaymentProcessor handles payment verification
type PaymentProcessor struct {
    // Using a buffered channel to simulate the bottleneck
    // This ensures we can't process more than X payments concurrently
    semaphore chan struct{}
}

// NewPaymentProcessor creates a new payment processor
func NewPaymentProcessor(maxConcurrent int) *PaymentProcessor {
    return &PaymentProcessor{
        semaphore: make(chan struct{}, maxConcurrent),
    }
}

// ProcessPayment simulates payment verification with 3-second delay
func (p *PaymentProcessor) ProcessPayment(order *models.Order) error {
    // Acquire semaphore
    p.semaphore <- struct{}{}
    defer func() { <-p.semaphore }()
    
    // Simulate payment processing delay (3 seconds)
    time.Sleep(3 * time.Second)
    
    // In real world, this would call payment gateway
    // For now, just return success
    fmt.Printf("Payment processed for order: %s\n", order.OrderID)
    return nil
}