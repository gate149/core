# tester

[![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-008080?style=flat-square&logo=go&logoColor=white)](https://gofiber.io/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white)](https://www.docker.com/)
[![AWS S3](https://img.shields.io/badge/AWS_S3-F90?style=flat-square&logo=amazonaws&logoColor=white)](https://aws.amazon.com/s3/)
[![Valkey](https://img.shields.io/badge/Valkey-A00?style=flat-square&logo=redis&logoColor=white)](https://valkey.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![SeaweedFS](https://img.shields.io/badge/SeaweedFS-2A8000?style=flat-square)](https://github.com/seaweedfs/seaweedfs)
[![Redis](https://img.shields.io/badge/Redis-B00?style=flat-square&logo=redis&logoColor=white)](https://redis.io/)
[![Pandoc](https://img.shields.io/badge/Pandoc-4A4A4A?style=flat-square)](https://pandoc.org/)
[![OpenAPI v3](https://img.shields.io/badge/OpenAPI-v3-6BA81E?style=flat-square&logo=swagger&logoColor=white)](https://swagger.io/specification/)

`tester` is a backend service designed for managing programming competitions. It handles problems, contests,
participants, and their submissions, as well as user authentication and management. The service is developed in Go using
the Fiber framework. PostgreSQL serves as the relational database, Valkey (or Redis) is used for caching and session
management, and SeaweedFS provides distributed file storage with an S3-compatible interface. Pandoc is used to convert
problem statements from LaTeX to HTML.

For understanding the architecture, see the [documentation](https://github.com/gate149/docs).

## Features

- Manage programming contests, problems, and participant submissions.
- User authentication and management via JWT.
- File storage using SeaweedFS with an S3-compatible API.
- LaTeX to HTML conversion for problem statements using Pandoc.
- RESTful API defined with OpenAPI.
- Websocket support for real-time updates.

## Prerequisites

Before you begin, ensure you have the following dependencies installed:

- **Docker** and **Docker Compose**: To run PostgreSQL, Pandoc, Valkey, and SeaweedFS.
- **Goose**: For applying database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`).
- **oapi-codegen**: For generating OpenAPI code (`go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest`).

## 1. Running Dependencies

The service depends on PostgreSQL, Pandoc, Valkey, and SeaweedFS, which can be run using Docker Compose. Below is an
example `docker-compose.yml` configuration:

```yaml
version: '3.8'
services:
  pandoc:
    image: pandoc/latex
    ports:
      - "4000:3030"
    command: "server"
  postgres:
    image: postgres:14.1-alpine
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: supersecretpassword
      POSTGRES_DB: tester
    ports:
      - '5432:5432'
    volumes:
      - ./postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: pg_isready -U postgres -d tester
      interval: 10s
      timeout: 3s
      retries: 5
  valkey:
    image: valkey/valkey:latest
    volumes:
      - ./conf/valkey.conf:/usr/local/etc/valkey/valkey.conf
      - ./valkey-data:/data
    command: [ "valkey-server", "/usr/local/etc/valkey/valkey.conf" ]
    healthcheck:
      test: [ "CMD-SHELL", "valkey-cli ping | grep PONG" ]
      interval: 10s
      timeout: 3s
      retries: 5
    ports:
      - "6379:6379"
  master:
    image: chrislusf/seaweedfs:3.77_full
    ports:
      - "9333:9333"
      - "19333:19333"
      - "9324:9324"
    command: "-v 9 master -ip=master -ip.bind=0.0.0.0 -metricsPort=9324"
  volume:
    image: chrislusf/seaweedfs:3.77_full
    ports:
      - "8080:8080"
      - "18080:18080"
      - "9325:9325"
    command: '-v 9 volume -mserver="master:9333" -ip.bind=0.0.0.0 -port=2 -metricsPort=9325'
    depends_on:
      - master
  filer:
    image: chrislusf/seaweedfs:3.77_full
    ports:
      - "8888:8888"
      - "18888:18888"
      - "9326:9326"
    command: '-v 9 filer -master="master:9333" -ip.bind=0.0.0.0 -metricsPort=9326'
    depends_on:
      - master
      - volume
  s3:
    image: chrislusf/seaweedfs:3.77_full
    ports:
      - "8333:8333"
      - "9327:9327"
    command: '-v 9 s3 -filer="filer:8888" -ip.bind=0.0.0.0 -metricsPort=9327'
    depends_on:
      - master
      - volume
      - filer
  nats:
    image: nats:2.10
    ports:
      - "4222:4222"
      - "8222:8222"
    volumes:
      - nats-data:/data
      - ./nats.conf:/etc/nats/nats.conf
    command: ["-c", "/etc/nats/nats.conf"]

volumes:
  postgres-data:
  valkey-data:
  nats-data:
    name: nats-data
```

Start the services in detached mode:

```bash
docker-compose up -d
```

#### SeaweedFS Configuration

SeaweedFS is used for distributed file storage with an S3-compatible API. The s3.json file is required to configure S3
credentials and permissions. Place it in the project root or a designated configuration directory. An example s3.json is
shown below:

```json
{
  "identities": [
    {
      "name": "some_admin_user",
      "credentials": [
        {
          "accessKey": "some_access_key1",
          "secretKey": "some_access_key1"
        }
      ],
      "actions": [
        "Admin",
        "Read",
        "List",
        "Tagging",
        "Write"
      ]
    }
  ],
  "accounts": [
    {
      "id": "testid",
      "displayName": "M. Tester",
      "emailAddress": "tester@ceph.com"
    }
  ]
}
```

#### NAS Configuration

```
port: 4222
http: 8222
logfile: "/data/nats.log"
```

Ensure the S3_ACCESS_KEY and S3_SECRET_KEY in your .env file match the credentials defined in s3.json.

## 2. Configuration

The application uses environment variables for configuration. Create a .env file in the project root with the following
variables:

```dotenv
# Environment type (development or production)
ENV=dev

# Address and port where the tester service will listen
ADDRESS=0.0.0.0:13000

# Address of the running Pandoc service
PANDOC=http://localhost:4000

# PostgreSQL connection string (Data Source Name)
POSTGRES_DSN=host=localhost port=5432 user=postgres password=supersecretpassword dbname=tester sslmode=disable

# Valkey/Redis connection string
REDIS_DSN=valkey://localhost:6379/0

# Secret key for signing and verifying JWT tokens
JWT_SECRET=secret

# Default admin credentials
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin

# SeaweedFS S3 configuration
S3_ENDPOINT=http://localhost:8333
S3_ACCESS_KEY=some_access_key1
S3_SECRET_KEY=some_access_key1

# Cache configuration
# is needed to download archives from S3 and store tests in the cache
CACHE_DIR=C:\Users\You\gate7\tester\cache

NATS_URL=nats://localhost:4222
```

Important: Replace supersecretpassword, secret, admin, some_access_key1, and other sensitive values with secure, unique
values for production.

## 3. Database Migrations

The project uses goose to manage the database schema.
Ensure goose is installed:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Apply migrations to the PostgreSQL database:

```bash
goose -dir ./migrations postgres "host=localhost port=5432 user=postgres password=supersecretpassword dbname=tester sslmode=disable" up
```

## 4. OpenAPI Code Generation

The API is defined using OpenAPI, and Go code for handlers and models is generated with oapi-codegen.
Run the generation command:

```bash
make gen
```

## 5. Running the Application

Start the tester service:

```bash
go run ./main.go
```

The service will be available at the address specified in the ADDRESS variable (e.g., http://localhost:13000).

## 6. Authentication and User Management

The service handles user authentication using JWT tokens, with credentials stored in PostgreSQL and sessions managed via
Valkey. Default admin credentials are set via ADMIN_USERNAME and ADMIN_PASSWORD in the .env file. Users can be managed
through API endpoints defined in the OpenAPI specification.