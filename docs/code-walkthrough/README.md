# コードウォークスルー

`main.go`および`internal/`配下のGoファイルそれぞれについて、何をしているか・なぜそう書いているかを解説する。1ファイル1解説の対応で、実際のディレクトリ構成をそのまま反映している(`main.go` → `main.go.md`)。

設計判断の背景(なぜこの技術・この構造を選んだか)は[docs/adr/](../adr/README.md)を参照。ここでは個々のファイルが「何をしているか」にフォーカスする。

## 索引

### エントリポイント

- [main.go](./main.go.md)

### internal/config

- [config.go](./internal/config/config.go.md)

### internal/db

- [db.go](./internal/db/db.go.md)

### internal/models

- [models.go](./internal/models/models.go.md)

### internal/httpx

- [response.go](./internal/httpx/response.go.md)

### internal/auth

- [jwt.go](./internal/auth/jwt.go.md) / [jwt_test.go](./internal/auth/jwt_test.go.md)
- [password.go](./internal/auth/password.go.md) / [password_test.go](./internal/auth/password_test.go.md)
- [middleware.go](./internal/auth/middleware.go.md) / [middleware_test.go](./internal/auth/middleware_test.go.md)

### internal/handlers

- [auth.go](./internal/handlers/auth.go.md) / [auth_test.go](./internal/handlers/auth_test.go.md)
- [tasks.go](./internal/handlers/tasks.go.md) / [tasks_test.go](./internal/handlers/tasks_test.go.md)
- [calendar_range_test.go](./internal/handlers/calendar_range_test.go.md)
- [testhelpers_test.go](./internal/handlers/testhelpers_test.go.md)
