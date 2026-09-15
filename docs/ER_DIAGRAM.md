## ER図 {#er図}

テーブル定義の詳細は[DB_SPEC.md](./DB_SPEC.md)を参照。

```mermaid
erDiagram
    USERS ||--o{ TASKS : "creates"
    TASKS o|--o{ TASKS : "parent of"

    USERS {
        bigint id PK
        varchar email UK "NOT NULL"
        varchar password_hash "NOT NULL"
        varchar name "NOT NULL"
        timestamp created_at
        timestamp updated_at
    }

    TASKS {
        bigint id PK
        bigint user_id FK "NOT NULL"
        bigint parent_id FK "NULL可・自己参照"
        varchar title "NOT NULL"
        text description
        varchar status "todo/in_progress/done"
        smallint estimated_weight "NOT NULL・1〜5"
        smallint actual_weight "NULL可・1〜5"
        date due_date
        timestamp completed_at "NULL可・カレンダー集計キー"
        timestamp created_at
        timestamp updated_at
    }
```

カーディナリティは以下の通り(IE表記):

| 関係 | カーディナリティ | 説明 |
|---|---|---|
| USERS → TASKS | 1 : 0..\* | 1人のユーザーは0件以上のタスクを持つ(`tasks.user_id` → `users.id`) |
| TASKS → TASKS(自己参照) | 0..1 : 0..\* | 1つのタスクの親は0または1、子タスクは0件以上(`tasks.parent_id` → `tasks.id`) |
