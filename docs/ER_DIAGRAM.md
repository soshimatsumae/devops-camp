## ER図 {#er図}

テーブル定義の詳細は[DB_SPEC.md](./DB_SPEC.md)を参照。

```mermaid
erDiagram
    USERS ||--o{ TASKS : "creates"
    USERS ||--o{ LISTS : "creates"
    TASKS o|--o{ TASKS : "parent of"
    LISTS o|--o{ TASKS : "groups"

    USERS {
        bigint id PK
        varchar email UK "NOT NULL"
        varchar password_hash "NOT NULL"
        varchar name "NOT NULL"
        timestamp created_at
        timestamp updated_at
    }

    LISTS {
        bigint id PK
        bigint user_id FK "NOT NULL"
        varchar name "NOT NULL"
        timestamp created_at
        timestamp updated_at
    }

    TASKS {
        bigint id PK
        bigint user_id FK "NOT NULL"
        bigint parent_id FK "NULL可・自己参照"
        bigint list_id FK "NULL可・親タスクのみ設定可"
        varchar title "NOT NULL"
        text description
        varchar status "todo/in_progress/done"
        decimal estimated_hours "NOT NULL・0.5〜9999の時間"
        decimal actual_hours "NULL可・0.5〜9999の時間"
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
| USERS → LISTS | 1 : 0..\* | 1人のユーザーは0件以上のリストを持つ(`lists.user_id` → `users.id`) |
| TASKS → TASKS(自己参照) | 0..1 : 0..\* | 1つのタスクの親は0または1、子タスクは0件以上(`tasks.parent_id` → `tasks.id`) |
| LISTS → TASKS | 0..1 : 0..\* | 1つのタスク(親タスクのみ)が所属するリストは0または1(`tasks.list_id` → `lists.id`、ON DELETE SET NULL) |
