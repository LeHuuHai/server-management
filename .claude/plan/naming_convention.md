# Naming Conventions for Microservices

Dựa trên cấu trúc đã được chuẩn hóa của `auth-service` và `server-service`, mọi service mới khi được migrate hoặc tạo mới phải tuân thủ nghiêm ngặt các quy tắc đặt tên file và thư mục sau đây để đảm bảo tính nhất quán trên toàn hệ thống.

---

## 1. Quy tắc chung (General Rules)
- **Thư mục (Directories)**: Luôn sử dụng `kebab-case` (ví dụ: `auth-service`, `server-service`). Ngoại trừ các thư mục cấu trúc bên trong thì viết thường, không dấu cách (ví dụ: `domain`, `infra`, `postgres`, `rpc`).
- **File Go (Go Files)**: Hầu hết sử dụng `camelCase.go` (ví dụ: `accountRepo.go`), ngoại trừ tầng RPC (`snake_case`).

---

## 2. Tầng Domain (`internal/domain/`)
Chứa các abstract interface cốt lõi của nghiệp vụ.
- **Thư mục**: Nhóm theo loại component (`repo`, `service`, `publisher`, `cache`).
- **Package name**: Trùng với tên thư mục component (ví dụ: `package repo`, `package service`).
- **Tên file**: Bắt buộc phải có hậu tố `Interface.go` và sử dụng `camelCase`.
- **Tên Interface (trong code)**: Phải có hậu tố `Interface` (Ví dụ: `type AccountRepositoryInterface interface`).
  - *Ví dụ:*
    - `internal/domain/repo/accountRepoInterface.go` (Package `repo`)
    - `internal/domain/service/authServiceInterface.go` (Package `serviceinterface`)
    - `internal/domain/publisher/eventPublisherInterface.go` (Package `publisher`)

---

## 3. Tầng Infrastructure (`internal/infra/`)
Chứa các implement thực tế của domain interface.
- **Thư mục**: Nhóm theo **công nghệ** được sử dụng (`postgres`, `redis`, `kafka`, v.v.). Các logic implement cho domain service thì nằm trong thư mục `service/`.
- **Package name**: Trùng với tên thư mục công nghệ (ví dụ: `package postgres`, `package kafka`, `package service`).
- **Tên file**: Sử dụng `camelCase.go`. Thường ghép từ tên model và tên loại file.
  - *Ví dụ:*
    - `internal/infra/postgres/accountRepo.go` (Package `postgres`)
    - `internal/infra/kafka/serverEventPublisher.go` (Package `kafka`)
    - `internal/infra/service/serverService.go` (Package `service`)

---

## 4. Tầng RPC / Handler (`internal/rpc/`)
Chứa các gRPC handler giao tiếp với Client/API Gateway.
- **Package name**: Luôn là `package rpc`.
- **Tên file**: Sử dụng `snake_case.go` và kết thúc bằng `_server.go`.
  - *Ví dụ:*
    - `internal/rpc/auth_server.go` (Package `rpc`)
    - `internal/rpc/server_server.go` (Package `rpc`)
- **Tên Struct (trong code)**: Sử dụng hậu tố `Handler` (Ví dụ: `type AuthHandler struct`).

---

## 5. Tầng Model (`internal/model/`)
Chứa các Struct định nghĩa dữ liệu (DB Models, DTOs, Event Payloads).
- **Package name**: Luôn là `package model`.
- **Tên file**: Sử dụng `camelCase.go`.
  - *Ví dụ:*
    - `internal/model/account.go` (Package `model`)
    - `internal/model/serverEvent.go` (Package `model`)
    - `internal/model/listServerFilter.go` (Package `model`)

---

## 6. Config và Runtime (`internal/config/`, `internal/infra/runtime/`)
Các file cốt lõi dùng để bootstrap service.
- **Config**: Luôn là `internal/config/config.go` (Package `config`). Khởi tạo các nested config từ thư viện chung `pkg/config`.
- **Runtime**: Luôn là `internal/infra/runtime/rt.go` (Package `rt`). Chứa struct `App` để wire dependencies.

---

## 7. Thư viện dùng chung (`pkg/`)
Chứa các helper, util, và wrapper infrastructure dùng chung cho mọi services.
- Nhóm theo chức năng cụ thể: `pkg/config/`, `pkg/db/`, `pkg/mq/`, `pkg/pb/`.
- **Package name**: Trùng với tên thư mục (ví dụ: `package config`, `package mq`).
- Các file bên trong dùng `snake_case.go` hoặc chữ thường hoàn toàn, tránh dùng camelCase trừ khi thật cần thiết.
  - *Ví dụ:*
    - `pkg/config/config.go` (Package `config`)
    - `pkg/db/postgres.go` (Package `db`)
    - `pkg/mq/writer.go` (Package `mq`)
