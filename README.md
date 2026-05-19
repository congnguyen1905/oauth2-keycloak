# OAuth2 Demo: Keycloak + Spring Boot

Project mẫu OAuth2 đầy đủ với:
- **Keycloak 22** làm Authorization Server (chạy qua Docker)
- **Spring Boot 2.7.18 / Java 8** vừa là **OAuth2 Client** (browser login) vừa là **Resource Server** (JWT API)

---

## 🏗️ Kiến trúc

```
Browser
  │
  ▼ (1) Redirect /oauth2/authorization/keycloak
Keycloak :8080  ──── Authorization Code Flow ────▶ Spring Boot :8081
  │                                                        │
  │  (2) Issue access_token (JWT)                   ├─ OAuth2 Client
  │                                                  │   (browser session)
  │                                                  └─ Resource Server
  │                                                      (JWT validation)
  ▼
REST API caller ──── Bearer <JWT> ──────────────────▶ /api/**
```

---

## 🚀 Khởi động nhanh

### Bước 1: Chạy Keycloak

```bash
docker-compose up -d
```

Chờ Keycloak khởi động (~30-60s), kiểm tra:
```
http://localhost:8080/realms/demo-realm
```
→ Phải trả về JSON. Nếu chưa thấy thì đợi thêm.

**Keycloak Admin Console:** http://localhost:8080 (admin/admin)

### Bước 2: Chạy Spring Boot

```bash
cd spring-app
mvn spring-boot:run
```

Hoặc build jar:
```bash
mvn clean package -DskipTests
java -jar target/oauth2-demo-0.0.1-SNAPSHOT.jar
```

### Bước 3: Mở ứng dụng

```
http://localhost:8081
```

---

## 👥 Tài khoản test

| Username | Password | Roles |
|----------|----------|-------|
| `user`   | `user123`  | ROLE_USER |
| `admin`  | `admin123` | ROLE_USER, ROLE_ADMIN |

---

## 🔗 Các URL quan trọng

| URL | Mô tả |
|-----|-------|
| `http://localhost:8081/` | Trang chủ |
| `http://localhost:8081/dashboard` | Dashboard (cần login) |
| `http://localhost:8081/admin` | Admin page (cần ROLE_ADMIN) |
| `http://localhost:8080` | Keycloak Admin Console |
| `http://localhost:8080/realms/demo-realm/.well-known/openid-configuration` | OIDC Discovery |

---

## 🛡️ API Endpoints (Resource Server)

Gọi với header: `Authorization: Bearer <access_token>`

| Method | Path | Quyền |
|--------|------|-------|
| GET | `/api/public/hello` | Public, không cần auth |
| GET | `/api/user/profile` | Cần JWT hợp lệ |
| GET | `/api/user/data` | `ROLE_USER` |
| GET | `/api/token/info` | Cần JWT hợp lệ |
| GET | `/api/admin/stats` | `ROLE_ADMIN` |

### Lấy access token qua curl (Direct Grant)

```bash
curl -s -X POST http://localhost:8080/realms/demo-realm/protocol/openid-connect/token \
  -d "client_id=spring-boot-app" \
  -d "client_secret=spring-boot-secret-key-2024" \
  -d "username=user" \
  -d "password=user123" \
  -d "grant_type=password" | jq .access_token -r
```

### Gọi API với token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/demo-realm/protocol/openid-connect/token \
  -d "client_id=spring-boot-app" \
  -d "client_secret=spring-boot-secret-key-2024" \
  -d "username=admin" \
  -d "password=admin123" \
  -d "grant_type=password" | jq .access_token -r)

# Public API
curl http://localhost:8081/api/public/hello

# Protected API
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/user/profile

# Admin API (cần user admin)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/admin/stats
```

---

## 📁 Cấu trúc project

```
oauth2-keycloak/
├── docker-compose.yml              ← Keycloak container
├── keycloak-config/
│   └── realm-export.json           ← Realm config tự động import
└── spring-app/
    ├── pom.xml                     ← Spring Boot 2.7.18, Java 8
    └── src/main/
        ├── java/com/example/oauth2demo/
        │   ├── OAuth2DemoApplication.java
        │   ├── config/
        │   │   ├── SecurityConfig.java         ← Dual chain: Client + Resource Server
        │   │   └── KeycloakLogoutHandler.java  ← SSO logout
        │   └── controller/
        │       ├── WebController.java          ← Browser pages
        │       └── ApiController.java          ← REST API
        └── resources/
            ├── application.yml     ← OAuth2 config
            └── templates/          ← Thymeleaf pages
```

---

## ⚙️ Giải thích SecurityConfig (quan trọng)

Spring Boot app có **2 Security Filter Chain**:

### Chain 1 (`@Order(1)`) - Resource Server
```
/api/** → Xác thực bằng JWT Bearer token → STATELESS
```
- Không có session, không redirect login
- Đọc JWT từ header `Authorization: Bearer ...`
- Convert roles từ claim `roles` trong JWT

### Chain 2 (`@Order(2)`) - OAuth2 Client
```
/** → Redirect sang Keycloak login → SESSION-based
```
- Dùng Authorization Code Flow
- Sau login, lưu session trên server
- Thymeleaf pages đọc `OidcUser` từ security context

---

## 🔧 Tùy chỉnh

### Đổi client secret
Sửa trong `keycloak-config/realm-export.json` và `application.yml` cùng lúc.

### Thêm user/role
Đăng nhập Keycloak Admin → Realm `demo-realm` → Users/Roles.

### Thêm scope
Trong `application.yml`:
```yaml
scope:
  - openid
  - profile
  - email
  - phone   # thêm scope mới
```
