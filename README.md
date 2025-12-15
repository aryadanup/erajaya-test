# Erajaya Test Product API

## 🛒 Introduction
The Product API is a robust, scalable service for managing product data. It is built with a focus on high performance, maintainability, and clean architecture principles. This service interacts with both PostgreSQL for persistent storage and Redis for high-speed caching.

## 🏗 Architecture
This project adheres to **Clean Architecture** principles to ensure separation of concerns and independent testability of business logic. The application is divided into four main layers:

1.  **Entity (Models)**: Defines the core data structures and business rules. Located in `internal/models/entity`.
2.  **Repository (Data Access)**: Handles direct interactions with databases (Postgres, Redis). It abstracts the data source from the business logic. Located in `internal/repository`.
3.  **Usecase (Business Logic)**: Contains the application's business rules and orchestrates data flow between the repository and the delivery layer. Located in `internal/usecase`.
4.  **Delivery (Transport)**: Handles HTTP requests, parses input, and formats responses. Located in `internal/delivery/http`.

🏗 **Why Clean Architecture?**
<br> We use Clean Architecture to keep the project organized, stable, and easy to grow.

-   **Reliable Testing**: We can test the main business logic quickly and safely. This allows us to achieve 100% Code Coverage without needing a complicated setup.

-   **Easy to Update**: The core application is safe from outside changes. We can switch databases or upgrade tools in the future without breaking the main features.

-   **Tidy & Organized Code**: Every part of the code has a specific place. This helps new developers understand the project faster and prevents messy code ("spaghetti code").

🧩 **Key Design Patterns**
-  **Repository Pattern**: Acts as a "Data Manager." The application simply asks for data, and this pattern handles whether to fetch it from the fast cache (Redis) or the main database (PostgreSQL).

-   **Factory Pattern**: "Centralizes the database setup in shared/datastore. It handles the complexity of connecting to Postgres or Redis, keeping the main application code clean and focused."

-   **Dependency Injection (DI)**: "Promotes a modular design where system components are loosely coupled. This makes it effortless to test individual parts or swap technologies in the future."

```
.
├── app             # Application bootstrapping (Database, Logger, Router)
├── conf            # Configuration files
├── internal        # Private application code
│   ├── delivery    # HTTP Handlers
│   ├── interfaces  # Interface definitions for dependency inversion
│   ├── models      # Domain entities and DTOs
│   ├── repository  # Database implementations
│   └── usecase     # Business logic implementations
├── migrations      # SQL migration files
├── shared          # Shared utilities, constants, and middleware
│   ├── datastore   # Database factory and connection logic
│   ├── middlewares # Context, Rate Limiting, etc.
│   ├── response    # Standardized API response structure
│   └── utils       # Helper functions (Validator, etc.)
├── swagger         # API Documentation
├── tests           # Integration tests
├── main.go         # Entry point
└── Makefile        # Build and run commands
```


## 🛠 Tech Stack
-   **Language**: Go 1.24+
-   **Web Framework**: Echo v4
-   **Database**: PostgreSQL 15+ (Optimized with pg_trgm)
-   **Cache**: Redis 7+
-   **ORM**: GORM v2
-   **Configuration**: Viper
-   **Logging**: Zap (Structured JSON Logging)
-   **Testing**: Testify (Suite, Assert, Mock)
-   **Docs**: Swaggo (Swagger UI)

## 🗄️ Database & Indexing Strategy
The database schema is heavily optimized for fuzzy search (ILIKE) and pagination, addressing common performance bottlenecks in e-commerce systems.

#### Table
<details>
<summary><b>View Table Structure</b></summary>

| Column        | Type                     | Description                     |
| :---          | :---                     | :---                            |
| `id`          | `SERIAL PRIMARY KEY`     | Unique identifier               |
| `name`        | `VARCHAR(255)`           | Product name (Indexed)          |
| `price`       | `BIGINT`                  | Product price                   |
| `description` | `TEXT`                   | Detailed description            |
| `quantity`    | `INT`                    | Available stock                 |
| `created_at`  | `TIMESTAMP`              | Creation timestamp              |
| `created_by`  | `VARCHAR(255)`           | Creator identifier              |
| `updated_at`  | `TIMESTAMP`              | Last update timestamp           |
| `updated_by`  | `VARCHAR(255)`           | Last updater identifier         |
| `deleted_at`  | `TIMESTAMP`              | Soft delete timestamp           |
| `deleted_by`  | `VARCHAR(255)`           | Deleter identifier              |

</details>

#### Indexes
We utilize Partial Indexes (WHERE deleted_at IS NULL) to reduce index size and ensure only active records are indexed. Crucially, we utilize the PostgreSQL Trigram Extension (pg_trgm) with GIN Indexes to handle wildcard searches (%keyword%) efficiently

<details>
<summary><b>View Indexes</b></summary>

| Index Name                    | Description                              |
| :---                          | :---                                     |
| idx_products_search_gin       | GIN index for ILIKE search                |
| idx_products_created_at_sort  | Sort by created_at                        |
| idx_products_price_sort       | Sort by price                             |
| idx_products_name_sort        | Sort by name                              |

</details>
Migrations are handled using `golang-migrate` to ensure schema version control.

## 🚀 Performance Benchmarking
To validate system limits under high concurrency, we conducted load testing using Apache Benchmark (ab).

Test Scenario: 
- **Endpoint:** GET /api/v1/products?search=LG (Simulating search with Cache Hit)
- **Total Requests:** 10,000
- **Concurrency:** 100 concurrent users
- **Rate Limiter:** Adjusted to 10,000 req/s for stress testing.

📈 Benchmark Results
The system demonstrated exceptional throughput, handling over 14,500 requests per second on a single instance with zero failures.

<details>
<summary><b>View Benchmark </b></summary>

| Metric | Value | Description |
| :--- | :--- | :--- |
| **Requests per Second** | `14,584.11` | Throughput capacity |
| **Time per Request (Mean)** | `6.857 ms` | Average latency |
| **Failed Requests** | `0` | Stability indicator |
| **Transfer Rate** | `43.6 MB/sec` | Network efficiency |
| **95th Percentile** | `11 ms` | 95% of users waited less than 11ms |

</details>

Caching Strategy
-   **TTL**: 5 minutes default expiration.
-   **Invalidation**: Creating a new product invalidates related cache entries (`products*`).

Key Naming Convention
| Key Pattern                    | Description                              |
| :---                           | :---                                     |
| `products:detail:{id}`         | Cache for single product details         |
| `products:list:{query_string}` | Cache for product list with search/filter|



## 🧪 Testing & Code Coverage

We maintain high coding standards with a strict **100% Code Coverage** policy across all layers (Delivery, Usecase, Repository). This ensures robust business logic, secure data handling, and reliable API responses.

![Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen?style=for-the-badge&logo=go) ![Tests](https://img.shields.io/badge/tests-passing-brightgreen?style=for-the-badge)

### 📊 Coverage Summary

| Layer | Component | Coverage | Status |
| :--- | :--- | :--- | :--- |
| **Delivery** | HTTP Handler | **100.0%** | ✅ |
| **Business Logic** | Product Usecase | **100.0%** | ✅ |
| **Repository** | Postgres Implementation | **100.0%** | ✅ |
| **Repository** | Redis Implementation | **100.0%** | ✅ |
| **Total** | **All Statements** | **100.0%** | ✅ |

<details>
<summary><b>Click to view detailed terminal output</b></summary>

```bash
PASS
coverage: [no statements]
ok      erajaya-test/tests/integration  1.805s  coverage: [no statements]
erajaya-test/internal/delivery/http/product_handler.go:18:      NewHandler              100.0%
erajaya-test/internal/delivery/http/product_handler.go:36:      CreateProduct           100.0%
erajaya-test/internal/delivery/http/product_handler.go:68:      ListProducts            100.0%
erajaya-test/internal/delivery/http/product_handler.go:111:     GetProductByID          100.0%
erajaya-test/internal/repository/postgres_repository.go:18:     NewProductRepository    100.0%
erajaya-test/internal/repository/postgres_repository.go:24:     Create                  100.0%
erajaya-test/internal/repository/postgres_repository.go:28:     GetByID                 100.0%
erajaya-test/internal/repository/postgres_repository.go:40:     Fetch                   100.0%
erajaya-test/internal/repository/redis_repository.go:21:        NewRedisRepository      100.0%
erajaya-test/internal/repository/redis_repository.go:27:        Set                     100.0%
erajaya-test/internal/repository/redis_repository.go:31:        Get                     100.0%
erajaya-test/internal/repository/redis_repository.go:35:        Delete                  100.0%
erajaya-test/internal/repository/redis_repository.go:43:        deleteByPatternInternal 100.0%
erajaya-test/internal/usecase/product_usecase.go:24:            NewProductUsecase       100.0%
erajaya-test/internal/usecase/product_usecase.go:31:            CreateProduct           100.0%
erajaya-test/internal/usecase/product_usecase.go:53:            GetProductByID          100.0%
erajaya-test/internal/usecase/product_usecase.go:76:            ListProducts            100.0%
total:                                                          (statements)            100.0%
```
</details>


## 🛡 Security & Integration Test
The API is hardened against common vulnerabilities suitable for Enterprise deployment:
-   security & integration test located in `tests/integration/`

### Security Tests
**The suite includes dedicated tests for common vulnerabilities:**
-   **SQL Injection**: Verifies that inputs like `' OR '1'='1` are handled safely.
-   **XSS**: Checks that script tags in input are escaped or rejected.
-   **Input Validation**: Tests edge cases like negative prices or huge payloads.
-   **Rate Limiting**: Tests that requests exceed the limit.
-   **Information Disclosure**: Tests that sensitive information is not exposed.
-   **Error Handling**: Tests that errors are handled gracefully.

### Integration Tests
**These tests verify the end-to-end flow using:**
-   Actual **PostgreSQL** database.
-   Actual **Redis** instance.
-   Real HTTP requests via `httptest`.

## 🚦 Getting Started

### Prerequisites
-   Go 1.24+
-   Docker & Docker Compose
-   Make

<details>
<summary><b>Run Service</b></summary>

<br> **Step To Run Service**
1.  **Start Dependencies (DB/Redis)**:
    ```bash
    make docker-up
    ```
2.  **Intall Golang Migrate**:
    ```bash
    make migrate-setup
    ```
3.  **Run Migrations**:
    ```bash
    make migrate-up
    ```
4.  **Add config.json in conf/**:
    configuration is managed via JSON files in the conf/ directory.
    ```json
    {
        "server": {
            "port": "8080",
            "timeout": 30,
            "host": "localhost",
            "env": "development",
            "version": "1.0.0",
            "app_name": "Product API",
            "rate_limit": 10000
        },
        "postgres": {
            "host": "localhost",
            "port": 5432,
            "user": "user",
            "password": "password",
            "dbname": "erajaya_db",
            "debug": true,
            "max_idle_conns": 50,
            "max_open_conns": 100,
            "conn_max_lifetime": "1h",
            "conn_max_idle_time": "10m"
        },
        "redis": {
            "host": "localhost",
            "port": 6379,
            "password": "",
            "dbname": "0"
        }
    }
    ```
5.  **Run Application**:
    ```bash
    make run
    ```

</details>

<details>
<summary><b>Run Unit Test</b></summary>

<br> Step To Run Test
-   **Unit Tests**:
    ```bash
    make test
    ```
-   **With Race Detector**:
    ```bash
    make test-race
    ```
</details>

<details>
<summary><b>Generate Mock with Mockery</b></summary>

<br>To install and run mockery:
1.  **Install Mockery**:
    ```bash
    make mockery-setup
    ```
2.  **Generate Mocks**:
    ```bash
    make mocks
    ```
</details>


## 📝 API Documentation
**Swagger UI is available at**:

```bash
    http://localhost:8080/swagger/index.html
```

**Generate Swagger**:

```bash
    make swagger
```
    

### **Endpoints**

-   **POST /api/v1/products**: Create a new product.

```bash
curl --location 'http://localhost:8080/api/v1/products' \
--header 'Content-Type: application/json' \
--data '{
    "name": "Samsung Galaxy S24 Ultra",
    "price": 19000000,
    "quantity": 100,
    "description": "AI Phone with Snapdragon 8 Gen 3",
    "created_by": "arya"
}'
```

-   **GET /api/v1/products**: List products (Supports: page, limit, search, sort). *Supports pagination, searching, and sorting.* <br>

    query parameter:
    -   **Newest**: `sort=newest` (Default)
    -   **Cheapest**: `sort=cheapest`
    -   **Most Expensive**: `sort=expensive`
    -   **Name (A-Z)**: `sort=name asc`
    -   **Name (Z-A)**: `sort=name desc` <br>


```bash
curl --location 'http://localhost:8080/api/v1/products?page=1&limit=10&sort=newest'
```
-   **GET /api/v1/products/:id**: Get product details.
```bash
curl --location 'http://localhost:8080/api/v1/products/1'
```

### Dictionary
| Code          | HTTP Status | Description                            |
| :---          | :---        | :---                                   |
| `PRD-ERA-200` | 200 OK      | Success                                |
| `PRD-ERA-201` | 201 Created | Resource successfully created          |
| `PRD-ERA-400` | 400 Bad Request| invalid input / Validation Error    |
| `PRD-ERA-410` | 400 Bad Request| Error Bind (JSON parsing failed)      |
| `PRD-ERA-404` | 404 Not Found| Resource not found                    |
| `PRD-ERA-405` | 405 Method Not Allowed| Method not supported            |
| `PRD-ERA-408` | 408 Request Timeout| Request Timeout    |
| `PRD-ERA-429` | 429 Too Many Requests| Rate limit exceeded           |
| `PRD-ERA-500` | 500 Internal Server Error| Unexpected server error    |

The API implements rate limiting (e.g., 20 requests/sec) to prevent abuse. Exceeding this limit triggers a `PRD-ERA-429` response.

We use **Zap Logger** for high-performance, structured logging.
**Standard Fields**:
-   `X-Request-Id`: Unique tracing ID.
-   `Method`: HTTP Method.
-   `Url`: Full request URL.
-   `Duration`: Processing time in milliseconds.
-   `Body`: Request payload (sanitized).
-   `ServerTime`: Timestamp of the log.