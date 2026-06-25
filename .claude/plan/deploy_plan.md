# Kế hoạch triển khai Docker Swarm (`deploy_plan.md`)

Tài liệu này hướng dẫn cách cấu hình môi trường và triển khai hệ thống microservices lên hạ tầng **Docker Swarm**. Kế hoạch này áp dụng các quyết định thiết kế tối ưu nhất: **Không dùng `godotenv` ở Producton, quản lý cấu hình tập trung bằng YAML Anchors, và bảo mật thông tin nhạy cảm qua Docker Secrets.**

---

## 1. Nguyên tắc cấu hình môi trường ở Production

Để đảm bảo tính linh hoạt, bảo mật và chuẩn cloud-native (12-Factor App), cấu hình hệ thống tuân theo các nguyên tắc sau:
1. **Image bất biến (Immutable Images):** Docker Image của các microservices hoàn toàn không chứa file `.env` hay bất kỳ thông số cấu hình cứng nào. Một Image duy nhất có thể chạy ở Dev, Staging, và Production.
2. **Không sử dụng `godotenv` ở Production:** Toàn bộ cấu hình được nạp trực tiếp qua các biến môi trường (OS Environment Variables) do Docker Swarm inject vào container lúc khởi chạy.
3. **Phân tách dữ liệu nhạy cảm (Docker Secrets):** Mật khẩu, private keys, API keys được phân phối an toàn thông qua Docker Secrets (mount vào `/run/secrets/`).

---

## 2. Quản lý cấu hình tập trung bằng YAML Anchors & Aliases

Để tránh lặp lại cấu hình dùng chung (như Kafka broker, Redis address, cấu hình Log) cho 6 microservices khác nhau, chúng ta khai báo các "nhãn cấu hình" dùng chung ở đầu file compose và merge chúng vào các service.

### Ví dụ file stack triển khai: `docker-stack.prod.yml`

```yaml
version: '3.8'

# ==========================================
# 1. CẤU HÌNH DÙNG CHUNG (SHARED CONFIGS)
# ==========================================
x-kafka-config: &kafka-config
  KAFKA_BROKER: kafka:9092
  KAFKA_HEARTBEAT_TOPIC: heartbeats
  KAFKA_MAIL_TOPIC: mail

x-database-config: &database-config
  DB_HOST: pg-db
  DB_PORT: 5432
  DB_NAME: server_monitoring
  DB_USER: postgres

x-common-log-config: &log-config
  ENV: production
  LOG_LEVEL: info

# ==========================================
# 2. ĐỊNH NGHĨA CÁC SERVICES
# ==========================================
services:

  # Gateway (Chỉ cần Kafka & Logging)
  heartbeat-gateway:
    image: myregistry.com/heartbeat-gateway:latest
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
    environment:
      <<: *log-config
      <<: *kafka-config
      PORT: 8082
    ports:
      - "8082:8082"
    secrets:
      - gateway_api_key
    networks:
      - sm-network

  # Monitor Service (Cần DB, Kafka, Logging)
  monitor-service:
    image: myregistry.com/monitor-service:latest
    deploy:
      replicas: 2
    environment:
      <<: *log-config
      <<: *kafka-config
      <<: *database-config
    secrets:
      - db_password
    networks:
      - sm-network
    ports:
      - "50054:50054"

  # Mail Worker (Chỉ cần Kafka & Logging)
  mail-worker:
    image: myregistry.com/mail-worker:latest
    deploy:
      replicas: 2
    environment:
      <<: *log-config
      <<: *kafka-config
      GOMAIL_ADDR: smtp.gmail.com
      GOMAIL_PORT: 587
      REPORT_REPO_ADDR: monitor-service:50054
    secrets:
      - smtp_password
    networks:
      - sm-network

# ==========================================
# 3. MẠNG NỘI BỘ VÀ BẢO MẬT (SECRETS & NETWORKS)
# ==========================================
networks:
  sm-network:
    driver: overlay

secrets:
  db_password:
    external: true
  smtp_password:
    external: true
  gateway_api_key:
    external: true
```

---

## 3. Các bước triển khai chi tiết

### Bước 1: Thiết lập Docker Secrets trên Swarm Manager
Trước khi deploy stack, bạn cần khởi tạo các secret bảo mật trực tiếp trên node Swarm Manager. Các giá trị này sẽ được mã hóa và lưu trữ an toàn:

```bash
# Tạo mật khẩu Database
echo "SuperSecretDBPassword123" | docker secret create db_password -

# Tạo mật khẩu SMTP Mail
echo "smtp_app_password_here" | docker secret create smtp_password -

# Tạo API Key cho Gateway
echo "some-key" | docker secret create gateway_api_key -
```

### Bước 2: Deploy Stack lên Cluster
Chạy lệnh deploy stack lên Swarm cluster:
```bash
docker stack deploy -c docker-stack.prod.yml server-monitoring
```

### Bước 3: Cập nhật cấu hình không gián đoạn (Rolling Updates)
Khi cần thay đổi cấu hình (ví dụ: chuyển đổi `LOG_LEVEL` từ `info` sang `debug` hoặc đổi địa chỉ Kafka):
1. **Chỉ cần chỉnh sửa trực tiếp** file `docker-stack.prod.yml`.
2. **Chạy lại lệnh deploy:**
   ```bash
   docker stack deploy -c docker-stack.prod.yml server-monitoring
   ```
3. Docker Swarm sẽ so sánh cấu hình mới, phát hiện thay đổi và thực hiện khởi động lại tuần tự các container (Rolling Update) với các biến môi trường mới **mà không cần rebuild hay repush Docker Image**.

---

## 4. Triển khai Agent ở Host-Level (Máy chủ được giám sát)

Vì `agent` chạy phân tán trên các máy chủ mục tiêu (không nằm trong cụm Docker Swarm trung tâm):

1. **Phân phối:** Biên dịch `agent` thành binary chạy độc lập cho từng OS mục tiêu (ví dụ: Linux amd64).
2. **Cấu hình:** Sử dụng file cấu hình cục bộ riêng biệt cho từng Host tại `/etc/server-monitor/agent.env`:
   ```env
   APP_SERVER_ID=server-001  # Tên định danh duy nhất của server này
   APP_HEARTBEAT_URL=http://<SWARM_GW_IP>:8082/heartbeat
   APP_HEARTBEAT_KEY=some-key
   APP_CYCLE_HEARTBEAT=5000
   ```
3. **Khởi chạy:** Chạy Agent dưới dạng một `systemd` service để đảm bảo tự động phục hồi khi máy chủ restart.
