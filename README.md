# Receipt Processor API

A REST API service that processes receipts and calculates reward points based on specific rules.

## Features

- Process receipts and generate unique IDs
- Calculate points based on receipt data
- Thread-safe receipt storage
- Input validation and error handling

### Prerequisites
- Go 1.21 or higher
- Git

### Installation
1. Clone the repository
```
git clone https://github.com/ChenBoYam/receipt-processor-challenge.git
cd receipt-processor
```

2. Initialize the Go module
```
go mod init receipt-processor
```

3. Install dependencies
```
go mod tidy
```

4. Run the application
```
go run main.go
```

The server will start at `http://localhost:8080`

## API Documentation

### 1. Process Receipt
**Endpoint:** `POST /receipts/process`

**Request:**
```
curl -X POST http://localhost:8080/receipts/process \
  -H "Content-Type: application/json" \
  -d '{
        "retailer": "Target",
        "purchaseDate": "2022-01-01",
        "purchaseTime": "13:01",
        "items": [
          {
            "shortDescription": "Mountain Dew 12PK",
            "price": "6.49"
          },{
            "shortDescription": "Emils Cheese Pizza",
            "price": "12.25"
          },{
            "shortDescription": "Knorr Creamy Chicken",
            "price": "1.26"
          },{
            "shortDescription": "Doritos Nacho Cheese",
            "price": "3.35"
          },{
            "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
            "price": "12.00"
          }
        ],
        "total": "35.35"
      }
```

**Success Response:**
```
{"id": "[uuid-id]" }
```

### 2. Get Points
**Endpoint:** `GET /receipts/{id}/points`

**Request:**
```
curl http://localhost:8080/receipts/[uuid-id]/points 
```

**Success Response:**
```
{"points": 28}
```

## Points Calculation Rules

1. One point for each alphanumeric character in the retailer name
2. 50 points if the total is a round dollar amount with no cents
3. 25 points if the total is a multiple of `0.25`
4. 5 points for every two items on the receipt
5. If the trimmed length of the item description is a multiple of `3`, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.
6. 6 points if the day in the purchase date is `odd`
7. 10 points if the time of purchase is between `2:00pm` and `4:00pm`

## Error Handling

The API returns appropriate HTTP status codes:
- 200: Successful operation
- 400: Invalid input
- 404: Receipt not found

Error responses include a message explaining the error, ex:
```
{"error": "invalid JSON"}
```

## Technical Details

- Uses Gin framework for routing and request handling
- Thread-safe with mutex for concurrent access
- UUID generation for receipt IDs

## License

This project is licensed under the MIT License - see the LICENSE file for details
