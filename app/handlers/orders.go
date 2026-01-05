package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "flash-sale-app/models"
    "flash-sale-app/services"
    "github.com/google/uuid"
)

type OrderHandler struct {
    paymentProcessor *services.PaymentProcessor
    awsServices      *services.AWSServices
}

func NewOrderHandler(pp *services.PaymentProcessor, aws *services.AWSServices) *OrderHandler {
    return &OrderHandler{
        paymentProcessor: pp,
        awsServices:      aws,
    }
}

// HandleSyncOrder processes orders synchronously
func (h *OrderHandler) HandleSyncOrder(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var order models.Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Generate order ID and timestamp
    order.OrderID = uuid.New().String()
    order.CreatedAt = time.Now()
    order.Status = "processing"

    // Process payment synchronously (3 second delay)
    startTime := time.Now()
    if err := h.paymentProcessor.ProcessPayment(&order); err != nil {
        order.Status = "failed"
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Payment processing failed",
        })
        return
    }

    // Update order status
    order.Status = "completed"
    processingTime := time.Since(startTime)

    // Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "order_id": order.OrderID,
        "status": order.Status,
        "processing_time_ms": processingTime.Milliseconds(),
    })
}

// HandleAsyncOrder queues orders for async processing
func (h *OrderHandler) HandleAsyncOrder(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Check if AWS services are available
    if h.awsServices == nil {
        http.Error(w, "Async processing not available - AWS services not configured", http.StatusServiceUnavailable)
        return
    }

    var order models.Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Generate order ID and timestamp
    order.OrderID = uuid.New().String()
    order.CreatedAt = time.Now()
    order.Status = "pending"

    // Publish to SNS for async processing
    startTime := time.Now()
    if err := h.awsServices.PublishOrder(order); err != nil {
        fmt.Printf("Failed to publish order: %v\n", err)
        http.Error(w, "Failed to queue order", http.StatusInternalServerError)
        return
    }
    
    publishTime := time.Since(startTime)
    fmt.Printf("Order %s queued for processing in %dms\n", order.OrderID, publishTime.Milliseconds())

    // Return immediate response (202 Accepted)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "order_id": order.OrderID,
        "status": "accepted",
        "message": "Order queued for processing",
        "queue_time_ms": publishTime.Milliseconds(),
    })
}

// HealthCheck endpoint for ECS health checks
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}