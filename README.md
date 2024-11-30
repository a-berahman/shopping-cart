# Shopping Cart Service

Dear team,

I appreciate the opportunity to share this solution. I've focused on creating a demonestrate both technical excellence and practical considerations. While implementing this solution, I've made deliberate architectural choices that prioritize reliability, maintainability, and scalability.
I encourage you to pay particular attention to the asynchronous reservation system design, which addresses the unique challenge of handling slow external services without compromising user experience. You'll notice that I've chosen a pragmatic balance between architectural purity and practical implementation.
In the following documentation, I've detailed not only what has been implemented but also, importantly, what should be considered for a full production deployment. This transparency reflects my belief that acknowledging system limitations and future improvements is as crucial as highlighting current capabilities.

## What is this project?
This is a shopping cart service that lets users add items to their cart and handles item reservations automatically. The main challenge it solves is "How to handle slow reservation services without making users wait?"

## Key Features

- Fast API Responses and asynchronous processing for reliable item reservation
- Redis-based job queue for guaranteed message delivery
- PostgreSQL for persistent storage
- Configuration management
- Structured logging and error handling
- Graceful shutdown mechanisms

## Quick Start

### Installation
**Note**: You need to Make and Docker as prerequisites

```
git clone https://github.com/a-berahman/shopping-cart.git
cd shopping-cart

# Start services
make run

# Stop services
make stop

```

## API Documentation

### SequenceDiagram

![sequenceDiagram-2024-11-30-231902](https://github.com/user-attachments/assets/31012085-dbe8-4265-a06a-cfbc2dd017bd)

### Curl example:

#### Add Item to Cart
```
curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{
    "name": "laptop",
    "quantity": 1
  }'
```

#### List Cart Items

```
curl http://localhost:8080/api/v1/items
```

## Technical Architecture

I follow hexagonal principles while maintaining pragmatic choices for real world requirements, the core domain remains isolated from external concerns through well defined ports, while adapters handle infrastructure interactions.

![Hexagonal-2024-11-30-232514](https://github.com/user-attachments/assets/d3fd94d5-c11c-4df2-bbbd-0c70b6c10eb1)



### Architecture Benefits:

1. **Separation of Concerns**:
   - Clear boundaries between layers
   - Business logic isolated from external concerns
   - Easy to modify or replace components

2. **Testability**:
   - Easy to mock external dependencies
   - Clear interfaces for testing
   - Business logic can be tested in isolation

3. **Maintainability**:
   - Clear folder structure
   - Well defined responsibilities

4. **Flexibility**:
   - Easy to add new features
   - Simple to change external implementations

### Structure

```
├── cmd/
│   └── main.go                 # Contains the main application entry point
│
├── config/                     # Configuration management
├── internal/                   # Core application code
│   ├── adapters/               # External system integrations
│   │   ├── handler/            # HTTP request handlers
│   │   ├── repository/         # Database operations
│   │   ├── queue/              # Queue operations
│   │   └── reservation/        # External service communication
│   │
│   ├── core/                   # Business logic and interfaces
│   │
│   ├── service/               # Business logic implementation
│
│   └── worker/                 # Background job processing
│
├── migrations/                 # Database schema management
│
├── scripts/                    # Utility scripts
│   └── migration.sh

```


## Required Enhancements for Production


- [ ]  Rate Limiting
- [ ]  Barch processing for multiple reservation
- [ ]  Connection pooling
- [ ]  Metrics and monitoring implementation
- [ ]  Distributed tracing setup
- [ ]  Performance optimization
- [ ]  Security hardening
- [ ]  Documentation completeness (for example, Swagger)
- [ ]  Cover more edge cases in the unit tests

## Development Practices
#### Testing
```
# Run all tests
make test

# Generate coverage report
make coverage

```

##Conclusion
This implementation demonstrates a production minded approach to building a reliable shopping cart service, while core functionality is complete and robust, I've detailed the necessary enhancements for full production readiness. Also, the architecture allows for easy extension and maintenance.



## Exercise definition


## What?
Build an API (HTTP/JSON) for a shopping cart application that will reserve items once they are put into the cart. For the sake of this exercise, the API should offer the following endpoints:

GET /items which is listing all the items that were put into the cart
POST /items which will add a new item to the cart. Every item should have at least two properties (name and quantity)
The data should be stored in a database, so that it can survive application restarts.

While items can be added to the shopping cart only by providing their name, the API should have another useful feature: When an item is added, it should reach out to an external service which can reserve the item for the user. That service returns a reservation-ID, which should also be displayed to the shopping cart API user when they list all items.

The external reservation service provides an endpoint which could look like the following (but you are free to adapt the interface):

POST /reserve with body: {"item": "abc"} → returns a JSON like {"reservation_id": 1234}
Unfortunately this external reservation service is fairly slow - and calls to the above mentioned endpoint can easily take more than 30 seconds.

Please make sure to:

Build the shopping cart API such that it always responds to API requests immediately (so probably the reservation API call needs to run in the background)
Mock the (non-existing) reservation API in your tests
Preferred technologies: Python or Go.

## When?
Please send us a link to your finished work until the agreed upon time. Make sure to include all recipients of the original email.

Once we receive your submission, we will do a code review - and, if we like what we see, schedule the follow-up interviews.

## How?
We strongly recommend you to limit your time spent on this exercise to 3-4 hours.

In that timeframe, not everything can be done to perfection. We are aware of that. From our side, we would propose to focus on a few components/elements of the overall solution, making these parts as “production ready” as possible.

The code must not be runnable! So it's completely fine to have only stubs/mocks at various places in your code.

At the end, please add a README to the repository with at least the following content:

documentation on how to use the code
a self-assessment of your solution: What is missing and couldn’t be done anymore? Which elements should be improved to make it production-ready? Were there any topics you were struggling with?
some reasoning for major tooling/framework/library decisions you made during the development process (why did you pick solution A over B?)
Where?
Please place your code in a public repository (GitHub, Gitlab, ...). Be aware that we will ask you to remove the repository once we finish the review.

Things we care about
Easily readable, well-structured, maintainable code
Documentation
Test coverage
Build automation & easy deployability
