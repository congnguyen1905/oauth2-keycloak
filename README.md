# OAuth2 Demo: Keycloak + Auth Service (Golang) + Next.js + KrakenD Gateway

Hệ thống OAuth2 hiện đại với kiến trúc microservices:
- **Keycloak 22** làm Authorization Server
- **Auth Service (Golang)** xác thực JWT cục bộ, quản lý session (Redis/DB)
- **Next.js (TypeScript)** Frontend chỉ lưu sessionId
- **KrakenD** API Gateway
- **Spring Boot 2.7.18 / Java 8** Backend Resource Server

---

## 🏗️ Kiến trúc hệ thống

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   Next.js   │      │   KrakenD    │      │Auth Service │
│  Frontend   │─────▶│   Gateway    │─────▶│   (Golang)  │
│ (sessionId) │      │   (:8888)    │      │  (JWT decode)│
└─────────────┘      └──────────────┘      └─────────────┘
                            │                     │
                            │                     ▼
                            │              ┌─────────────┐
                            │              │    Redis    │
                            │              │  + Database │
                            │              └─────────────┘
                            ▼
                     ┌─────────────┐
                     │ Spring Boot │
                     │   Backend   │
                     │   (:8081)   │
                     └─────────────┘
                            │
                            ▼
                     ┌─────────────┐
                     │  Keycloak   │
                     │   (:8080)   │
                     └─────────────┘
```

### Luồng xác thực:
1. User login tại Next.js → gửi credentials đến Auth Service
2. Auth Service xác thực với Keycloak → tạo sessionId → lưu vào Redis/DB
3. Auth Service trả về sessionId cho Next.js (lưu trong localStorage)
4. Next.js gửi sessionId trong header `X-Session-ID` đến KrakenD
5. KrakenD route request đến Spring Boot hoặc Auth Service
6. Spring Boot gọi Auth Service để validate sessionId → lấy thông tin user

---

## 🚀 Khởi động nhanh

### Bước 1: Chạy toàn bộ hệ thống với Docker Compose

```bash
docker-compose up --build
```

Hệ thống sẽ khởi động:
- **Keycloak**: http://localhost:8080
- **KrakenD Gateway**: http://localhost:8888
- **Auth Service**: http://localhost:8082
- **Spring Boot Backend**: http://localhost:8081
- **Next.js Frontend**: http://localhost:3000
- **Redis**: localhost:6379

Chờ khoảng 60s để tất cả services sẵn sàng.

### Bước 2: Truy cập ứng dụng

Mở trình duyệt:
```
http://localhost:3000
```

Hoặc test API qua Gateway:
```
http://localhost:8888/api/app/public/hello
```

---

## 👥 Tài khoản test

| Username | Password | Roles |
|----------|----------|-------|
| `user`   | `user123`  | ROLE_USER |
| `admin`  | `admin123` | ROLE_USER, ROLE_ADMIN |

---

## 🔗 Các URL quan trọng

| Service | URL | Mô tả |
|---------|-----|-------|
| **Frontend** | `http://localhost:3000` | Next.js App |
| **Gateway** | `http://localhost:8888` | KrakenD API Gateway |
| **Keycloak** | `http://localhost:8080` | Keycloak Admin Console |
| **Auth Service** | `http://localhost:8082` | Golang Auth Service |
| **Spring Boot** | `http://localhost:8081` | Backend Resource Server |
| **OIDC Discovery** | `http://localhost:8080/realms/demo-realm/.well-known/openid-configuration` | OIDC Config |

---

## 🛡️ API Endpoints

### Qua Gateway (http://localhost:8888)

#### Auth Service Routes (`/api/auth/*`)
| Method | Path | Mô tả |
|--------|------|-------|
| POST | `/api/auth/login` | Đăng nhập, nhận sessionId |
| POST | `/api/auth/logout` | Đăng xuất, hủy session |
| GET | `/api/auth/validate` | Validate sessionId |
| GET | `/api/auth/introspect` | Lấy thông tin user từ token |

#### Spring Boot Routes (`/api/app/*`)
Gọi với header: `X-Session-ID: <sessionId>`

| Method | Path | Quyền |
|--------|------|-------|
| GET | `/api/app/public/hello` | Public |
| GET | `/api/app/user/profile` | Cần sessionId hợp lệ |
| GET | `/api/app/user/data` | `ROLE_USER` |
| GET | `/api/app/token/info` | Cần sessionId hợp lệ |
| GET | `/api/app/admin/stats` | `ROLE_ADMIN` |

---

## 📋 Test API với curl

### 1. Đăng nhập và lấy sessionId

```bash
curl -X POST http://localhost:8888/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user",
    "password": "user123"
  }'
```

Response:
```json
{
  "session_id": "abc123xyz...",
  "expires_at": "2024-01-01T12:00:00Z"
}
```

### 2. Gọi API với sessionId

```bash
SESSION_ID="abc123xyz..."

# Public API
curl http://localhost:8888/api/app/public/hello

# Protected API
curl -H "X-Session-ID: $SESSION_ID" \
  http://localhost:8888/api/app/user/profile

# Admin API (cần user admin)
curl -H "X-Session-ID: $SESSION_ID" \
  http://localhost:8888/api/app/admin/stats
```

### 3. Validate sessionId

```bash
curl -H "X-Session-ID: $SESSION_ID" \
  http://localhost:8888/api/auth/validate
```

### 4. Logout

```bash
curl -X POST -H "X-Session-ID: $SESSION_ID" \
  http://localhost:8888/api/auth/logout
```

---

## 📁 Cấu trúc project

```
oauth2-keycloak/
├── docker-compose.yml              ← Orchestration toàn bộ hệ thống
├── gateway/
│   └── krakend.json                ← KrakenD Gateway config
├── auth-service/
│   ├── main.go                     ← Entry point
│   ├── go.mod                      ← Go dependencies
│   ├── config/                     ← Configuration module
│   │   └── config.go
│   ├── pkg/
│   │   ├── models/                 ← Data models
│   │   │   └── models.go
│   │   ├── jwt/                    ← JWT decoding (không gọi Keycloak)
│   │   │   └── jwt.go
│   │   ├── storage/                ← Storage interface (Redis/DB)
│   │   │   ├── storage.go
│   │   │   ├── redis.go
│   │   │   └── database.go
│   │   ├── service/                ← Business logic
│   │   │   └── auth_service.go
│   │   └── handler/                ← HTTP handlers
│   │       └── handler.go
│   └── Dockerfile
├── frontend/
│   ├── package.json                ← Next.js dependencies
│   ├── pages/
│   │   ├── index.tsx               ← Login page
│   │   └── dashboard.tsx           ← Dashboard page
│   ├── styles/
│   └── Dockerfile
├── spring-app/
│   ├── pom.xml                     ← Spring Boot 2.7.18, Java 8
│   └── src/main/
│       ├── java/com/example/oauth2demo/
│       │   ├── OAuth2DemoApplication.java
│       │   ├── config/
│       │   │   ├── SecurityConfig.java
│       │   │   └── SessionAuthenticationFilter.java  ← Validate X-Session-ID
│       │   ├── client/
│       │   │   └── AuthServiceClient.java            ← Call Auth Service
│       │   └── controller/
│       │       ├── WebController.java
│       │       └── ApiController.java
│       └── resources/
│           ├── application.yml
│           └── templates/
└── keycloak-config/
    └── realm-export.json           ← Realm config
```

---

## ⚙️ Auth Service Architecture

### Thiết kế Module
Auth Service được chia thành các module độc lập, dễ mở rộng:

1. **config/**: Quản lý cấu hình (environment variables, YAML)
2. **pkg/models/**: Định nghĩa data structures
3. **pkg/jwt/**: Giải mã JWT cục bộ (sử dụng public key của Keycloak)
   - Không cần gọi lại Keycloak để verify
   - Hỗ trợ RS256 signing algorithm
4. **pkg/storage/**: Interface lưu trữ linh hoạt
   - `storage.go`: Interface chung
   - `redis.go`: Implement với Redis
   - `database.go`: Implement với Database (PostgreSQL/MySQL)
5. **pkg/service/**: Business logic layer
6. **pkg/handler/**: HTTP handlers (có thể thêm gRPC/HTTP3 sau)

### Mở rộng trong tương lai
- **gRPC**: Thêm `pkg/grpc/` với proto definitions
- **HTTP/3**: Thêm QUIC support trong server setup
- **Database**: Đã có interface, chỉ cần implement thêm driver

---

## 🔧 Tùy chỉnh

### Đổi cấu hình Auth Service
Sửa `auth-service/config/config.go` hoặc dùng environment variables:
```bash
AUTH_SERVICE_PORT=8082
KEYCLK_URL=http://keycloak:8080
REALM_NAME=demo-realm
CLIENT_ID=spring-boot-app
CLIENT_SECRET=spring-boot-secret-key-2024
REDIS_HOST=redis
REDIS_PORT=6379
```

### Đổi Database storage
Implement interface trong `auth-service/pkg/storage/database.go`:
```go
type Storage interface {
    Save(sessionID string, token string, expiresAt time.Time) error
    Get(sessionID string) (string, error)
    Delete(sessionID string) error
}
```

### Thêm user/role trong Keycloak
Đăng nhập Keycloak Admin → Realm `demo-realm` → Users/Roles.

---

## 🎯 Lợi ích kiến trúc mới

1. **Frontend nhẹ**: Chỉ lưu sessionId, không quản lý token phức tạp
2. **Backend stateless**: Spring Boot không cần biết về Keycloak
3. **Centralized Auth**: Auth Service đảm nhiệm mọi việc xác thực
4. **JWT local verification**: Auth Service giải mã JWT không cần gọi Keycloak
5. **Scalable**: Dễ dàng thêm gRPC, HTTP/3, đổi storage backend
6. **Gateway pattern**: KrakenD quản lý routing, rate limiting, authentication
