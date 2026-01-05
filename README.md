# Flash Sale Async Processing System

A cloud-native e-commerce order processing system demonstrating synchronous vs asynchronous architectures, built with Go, AWS ECS, SNS/SQS, and Lambda.

## 🎯 Project Overview

This project simulates a real-world flash sale scenario where an e-commerce platform needs to handle sudden traffic spikes. It compares synchronous and asynchronous order processing approaches, highlighting the benefits of event-driven architecture.

**Key Challenge**: Payment processing takes 3 seconds per order. During a flash sale with 60 orders/second, how do you prevent system failure?

**Solution**: Decouple order acceptance from payment processing using AWS SNS/SQS for reliable async processing.

## 🏗️ Architecture

### Synchronous Processing (Phase 1)
```
Customer → API → Payment (3s) → Response
```
- **Problem**: Customers wait 3+ seconds, system overwhelmed under load
- **Result**: Failed requests, poor user experience

### Asynchronous Processing (Phase 2-5)
```
Customer → API → SNS → SQS → Background Workers → Payment (3s)
         ↓
    202 Accepted (<100ms)
```
- **Benefit**: Instant response, orders queued for processing
- **Trade-off**: Queue management, worker scaling required

### Serverless Processing (Phase 6)
```
Customer → API → SNS → Lambda → Payment (3s)
```
- **Benefit**: Zero operational overhead, auto-scaling
- **Trade-off**: No message queuing, limited retry control

## 🛠️ Technology Stack

- **Language**: Go 1.21+
- **Cloud Platform**: AWS
  - ECS (Elastic Container Service) for containerized workloads
  - SNS (Simple Notification Service) for pub/sub messaging
  - SQS (Simple Queue Service) for reliable message queuing
  - Lambda for serverless compute
  - ALB (Application Load Balancer) for traffic distribution
- **Infrastructure as Code**: Terraform
- **Load Testing**: Locust (Python)
- **Containerization**: Docker

## 📁 Project Structure

```
flash-sale-async-system/
├── app/
│   ├── handlers/          # HTTP request handlers
│   ├── models/            # Order and Item data structures
│   ├── processor/         # SQS message processor (background worker)
│   └── services/          # Payment processing and AWS integrations
├── lambda/                # Serverless order processor
├── locust/                # Load testing scripts
├── terraform/             # Infrastructure as Code
├── Dockerfile             # Container for order API
├── Dockerfile.processor   # Container for background worker
├── go.mod                 # Go dependencies
└── main.go                # Application entry point
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- AWS Account with CLI configured
- Terraform 1.0+
- Docker
- Python 3.8+ (for load testing)

### Local Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/flash-sale-async-system.git
   cd flash-sale-async-system
   ```

2. **Install Go dependencies**
   ```bash
   go mod download
   ```

3. **Build the application**
   ```bash
   go build -o bin/api ./
   ```

### Infrastructure Deployment

1. **Initialize Terraform**
   ```bash
   cd terraform
   terraform init
   ```

2. **Deploy AWS infrastructure**
   ```bash
   terraform plan
   terraform apply
   ```

3. **Get the load balancer URL**
   ```bash
   terraform output alb_dns_name
   ```

### Running Load Tests

1. **Install Locust**
   ```bash
   pip install locust
   ```

2. **Run synchronous endpoint test**
   ```bash
   cd locust
   locust -f locustfile.py --host=http://YOUR-ALB-DNS
   ```

3. **Run asynchronous endpoint test**
   ```bash
   locust -f async_test.py --host=http://YOUR-ALB-DNS
   ```

## 📊 API Endpoints

### Synchronous Order Processing
```bash
POST /orders/sync
Content-Type: application/json

{
  "customer_id": 123,
  "items": [
    {
      "item_id": "item-1",
      "name": "Product Name",
      "price": 99.99,
      "quantity": 2
    }
  ]
}
```
**Response**: 200 OK (after 3s payment processing)

### Asynchronous Order Processing
```bash
POST /orders/async
Content-Type: application/json

{
  "customer_id": 123,
  "items": [...]
}
```
**Response**: 202 Accepted (immediate, <100ms)

### Health Check
```bash
GET /health
```
**Response**: 200 OK

## 🔧 Configuration

### Environment Variables

**Order API Service:**
- `PORT`: HTTP server port (default: 8080)
- `SNS_TOPIC_ARN`: SNS topic for publishing orders
- `SQS_QUEUE_URL`: SQS queue URL (optional for API)

**Order Processor Service:**
- `SQS_QUEUE_URL`: SQS queue URL (required)
- `WORKER_COUNT`: Number of concurrent workers (default: 1)
- `AWS_REGION`: AWS region (default: us-east-1)

**Lambda Function:**
- Configured via Terraform
- Triggered directly by SNS

## 📈 Performance Results

### Synchronous (Phase 1)
- **Normal Load (5 users)**: 100% success rate
- **Flash Sale (20 users)**: ~20% success rate, 3+ second response times

### Asynchronous with 1 Worker (Phase 3)
- **Flash Sale (60 req/s)**: 100% acceptance rate
- **Queue Depth**: Grows rapidly (queue buildup problem)

### Asynchronous with 20 Workers (Phase 5)
- **Flash Sale (60 req/s)**: 100% acceptance rate
- **Queue Depth**: Remains near zero
- **Processing Rate**: ~20 orders/second

### Serverless Lambda (Phase 6)
- **Flash Sale**: 100% acceptance rate
- **Cold Start Overhead**: ~70ms (2.4% of 3s processing)
- **Cost**: Free tier covers 267K orders/month

## 🔍 Monitoring

### CloudWatch Metrics
- `SQS ApproximateNumberOfMessagesVisible`: Queue depth
- `ECS CPUUtilization`: Container resource usage
- `ALB RequestCount`: Incoming traffic
- `Lambda Duration`: Function execution time
- `Lambda ConcurrentExecutions`: Auto-scaling behavior

## 🚨 Common Issues

**Problem**: Async endpoint returns 503  
**Solution**: Check that SNS_TOPIC_ARN and SQS_QUEUE_URL environment variables are set

**Problem**: Queue keeps growing  
**Solution**: Increase WORKER_COUNT in processor service

**Problem**: Lambda cold starts  
**Solution**: Expected behavior; occurs after ~5 minutes of inactivity



