## データベース仕様書 {#データベース仕様書}

### データベース概要 {#データベース概要}

タスクごとに「想定の大変さ」と「実際の大変さ」を記録し、見積もり精度を振り返れるタスク管理ツール用のデータベース。タスクは親子(孫)の階層構造を持ち、完了時の重さを日別に集計してコントリビューションカレンダー表示に用いる。

ER図は[ER_DIAGRAM.md](./ER_DIAGRAM.md)を参照。

### テーブル一覧 {#テーブル一覧}

| テーブル名 | 概要 |
| :---- | :---- |
| users | ユーザーアカウント情報 |
| tasks | タスク本体(親子・孫の階層関係、見積もり/実績の重さを保持) |

### users {#users}

ユーザーのアカウント情報を保持するテーブル。認証(ログイン)とタスクの所有者判定に使用する。

| カラム名 | 型 | 必須 | 制約 | 説明 |
| :---- | :---- | :---- | :---- | :---- |
| id | BIGINT | ○ | PK, AUTO\_INCREMENT | ユーザーID |
| email | VARCHAR(255) | ○ | UNIQUE | ログインに使用するメールアドレス |
| password\_hash | VARCHAR(255) | ○ | \- | ハッシュ化されたパスワード |
| name | VARCHAR(100) | ○ | \- | 表示名 |
| created\_at | TIMESTAMP | ○ | DEFAULT CURRENT\_TIMESTAMP | 作成日時 |
| updated\_at | TIMESTAMP | ○ | DEFAULT CURRENT\_TIMESTAMP ON UPDATE CURRENT\_TIMESTAMP | 更新日時 |

インデックス

| インデックス名 | 対象カラム | 用途 |
| :---- | :---- | :---- |
| idx\_users\_email | email | ログイン時の検索・重複防止(UNIQUE) |

### tasks {#tasks}

ユーザーが作成するタスク本体。親子(孫)の階層構造、見積もり/実績の重さ、完了日時を保持し、コントリビューションカレンダー表示の集計元データとなる。

| カラム名 | 型 | 必須 | 制約 | 説明 |
| :---- | :---- | :---- | :---- | :---- |
| id | BIGINT | ○ | PK, AUTO\_INCREMENT | タスクID |
| user\_id | BIGINT | ○ | FK → users.id | タスクの所有者 |
| parent\_id | BIGINT | \- | FK → tasks.id(自己参照) | 親タスクのID(子・孫タスクを表現) |
| title | VARCHAR(255) | ○ | \- | タスクのタイトル |
| description | TEXT | \- | \- | タスクの詳細説明 |
| status | VARCHAR(20) | ○ | DEFAULT 'todo' | ステータス('todo'/'in\_progress'/'done') |
| estimated\_weight | SMALLINT | ○ | \- | 作成時に入力する想定の大変さ(1〜5の5段階評価) |
| actual\_weight | SMALLINT | \- | \- | 完了時に入力する実際の大変さ(1〜5の5段階評価) |
| due\_date | DATE | \- | \- | 締切日 |
| completed\_at | TIMESTAMP | \- | \- | 完了日時(カレンダー集計のキー) |
| created\_at | TIMESTAMP | ○ | DEFAULT CURRENT\_TIMESTAMP | 作成日時 |
| updated\_at | TIMESTAMP | ○ | DEFAULT CURRENT\_TIMESTAMP ON UPDATE CURRENT\_TIMESTAMP | 更新日時 |

インデックス

| インデックス名 | 対象カラム | 用途 |
| :---- | :---- | :---- |
| idx\_tasks\_user\_id | user\_id | ユーザーごとのタスク一覧取得 |
| idx\_tasks\_parent\_id | parent\_id | 子・孫タスクの一覧取得 |
| idx\_tasks\_status | status | ステータス別フィルタ |
| idx\_tasks\_completed\_at | completed\_at | 日別集計(コントリビューションカレンダー表示) |
