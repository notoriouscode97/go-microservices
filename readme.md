# Coffee Shop Microservices Project

This project is a microservices-based system for a coffee shop application, built using Golang with four primary microservices to manage products, orders, product images, and currency conversion. It leverages RabbitMQ for message brokering, PostgreSQL for data storage, and gRPC for inter-service communication.

## Table of Contents
1. [Microservices Overview](#microservices-overview)
    - [Product API](#product-api)
    - [Order Service](#order-service)
    - [Product Images](#product-images)
    - [Currency Service](#currency-service)
2. [System Requirements](#system-requirements)
3. [Installation & Setup](#installation--setup)
4. [Environment Variables](#environment-variables)
5. [Usage](#usage)
    - [Product API](#product-api-1)
    - [Order Service](#order-service-1)
    - [Product Images](#product-images-1)
    - [Currency Service](#currency-service-1)
6. [Graceful Shutdown](#graceful-shutdown)
7. [License](#license)

## Microservices Overview

### 1. Product API
This service manages CRUD operations for products and initiates order creation. The API is built using the Gorilla Mux router and utilizes PostgreSQL as its database.

**Endpoints:**
- `/products`: CRUD operations for products
- `/orders`: Creates an order and publishes a message to the RabbitMQ exchange

**Message Broker:** Uses RabbitMQ to publish messages with the `orders_topic` topic.  
**Database:** PostgreSQL

### 2. Order Service
The Order Service processes orders by consuming messages from the RabbitMQ queue and handling order confirmation tasks.

**Message Consumption:** Listens to the `orders_topic` queue to process incoming order messages.  
**Database:** Writes order information to PostgreSQL.  
**Email Confirmation:** Sends a confirmation email upon successful order processing.

### 3. Product Images
The Product Images service handles image uploads for products, providing a handler for multipart data uploads.

**Endpoint:**
- `/images/{id}/{filename}`: Uploads and serves images associated with a product ID.

**File Server:** Configured with `http.StripPrefix` to serve files from the specified directory.

### 4. Currency Service
This service fetches currency rates from the European National Bank and communicates with the Product API via gRPC.

**Features:**
- Supports unary and bidirectional gRPC streams.
- Enables currency conversion for product prices based on real-time exchange rates.

## System Requirements
- Golang: v1.17 or above
- PostgreSQL: v13 or above
- RabbitMQ: v3.8 or above
- gRPC: v1.38 or above

## Installation & Setup
1. Clone the repository:
    ```bash
    git clone https://github.com/notoriouscode97/go-microservices.git
    cd go-microservices
    ```
2. Set up environment variables for each microservice (see below).
3. Run `go mod tidy` in each microservice directory to install dependencies.
4. Start each service:
    ```bash
    go run currency/main.go
    go run product-api/main.go
    go run order-service/main.go
    go run product-images/main.go
    ```

Ensure RabbitMQ and PostgreSQL are up and running, with connections configured in `.env` files for each service.

## Environment Variables
Each microservice relies on environment variables for configuration. Define these in a `.env` file or export them directly.

- `POSTGRES_HOST`: Hostname for PostgreSQL
- `POSTGRES_PORT`: Port for PostgreSQL
- `POSTGRES_USER`: Database user for PostgreSQL
- `POSTGRES_PASSWORD`: Password for PostgreSQL
- `RABBITMQ_URL`: Connection string for RabbitMQ
- `BASE_PATH`: Base directory for image storage in Product Images service

**Example .env:**
```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
BASE_PATH=./images
