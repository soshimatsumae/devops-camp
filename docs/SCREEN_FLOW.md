## 画面遷移図

各画面の詳細は[docs/SCREEN_LIST.md](./SCREEN_LIST.md)を参照。見た目の詳細は[ワイヤーフレーム](./design/README.md#ワイヤーフレーム)を参照(各フロー図の画面名に画面IDを併記している)。

ワイヤーフレームの各画面を並べ、ボタンから次の画面へユースケースごとに色分けした矢印でつないだ画面フローマップ。元データは[docs/design/flowmap.html](./design/flowmap.html)で、ローカルのブラウザで開くと凡例でフローを選んでそのフローだけを強調表示できる。

![画面フローマップ](./design/images/flowmap.png)

### 認証要否・未認証時の扱い

- `S-01`・`S-02`は認証不要。ログイン済みの状態でこれらにアクセスした場合は`S-03`へリダイレクトする(想定挙動として明記。バックエンド側では特に制御していないため、フロントエンド側のルーティングガードで実装する)
- `S-03`・`M-01`・`M-02`・`M-03`はすべて認証必要。JWTトークンを持たない状態(未ログイン、またはトークン期限切れで`401`が返ってきた場合)でアクセスすると`S-01`へリダイレクトする(詳細は下記「5. セッション切れ・未ログインアクセス」参照)
- ログアウトは専用APIを持たず、クライアント側でトークンを破棄して`S-01`へ遷移するだけで完結する([docs/SCREEN_LIST.md](./SCREEN_LIST.md)参照)。ログアウトボタンはS-03の共通ヘッダーにあり、モーダル(M-01〜M-03)表示中は背後のヘッダーを押せないため、ログアウトはS-03からのみ行う
- 一方、`401`によるS-01へのリダイレクトはS-03に限らず、M-01〜M-03でAPIを呼んだとき(作成・保存・詳細の読み込みなど)にも起こりうる
- S-03のタスク一覧、M-03の子タスク一覧における「クリックで子タスク/孫タスクを展開」操作は、同じ画面・モーダルの中でのインライン展開でありページ遷移・モーダル遷移を伴わないため、フロー図には含めていない([docs/SCREEN_LIST.md](./SCREEN_LIST.md)参照)

### ユースケース別フロー図

上の静的マップは「どの画面同士がつながっているか」は分かっても、開始・終了地点や正常系/准正常系/異常系の分岐が追いにくかった。そのため主要な6つのユースケースを、始点・終点・分岐を明示したフロー図として個別に書き起こす。各ノードには画面IDとワイヤーフレームのタブ名を併記し、エラーメッセージは実際にモックへ入れた文言をそのまま使っている。

#### 1. ログイン(正常系・異常系)

```mermaid
flowchart TD
    start(["開始: S-01(ログイン)を開く"]) --> input["メールアドレス・パスワードを入力"]
    input --> submit[["ログインボタン押下<br/>AUTH-02 POST /auth/login"]]
    submit --> ok{"200 OK?"}
    ok -- Yes --> token["access_tokenを保存"]
    token --> toDash["S-03(ダッシュボード)へ遷移"]
    toDash --> fin(["終了"])
    ok -- "No(401 INVALID_CREDENTIALS)" --> err["エラー表示:<br/>メールアドレスまたはパスワードが正しくありません"]
    err --> input
```

#### 2. 新規登録(正常系・異常系)

```mermaid
flowchart TD
    start(["開始: S-02(新規登録)を開く"]) --> input["お名前・メールアドレス・パスワードを入力"]
    input --> submit[["登録するボタン押下<br/>AUTH-01 POST /auth/register"]]
    submit --> ok{"201 Created?"}
    ok -- Yes --> toLogin["S-01へ遷移(登録後は改めてログインする方式)"]
    toLogin --> fin(["終了"])
    ok -- "No(422 EMAIL_ALREADY_REGISTERED)" --> err1["エラー表示:<br/>このメールアドレスは既に使われています"]
    ok -- "No(422 VALIDATION_ERROR)" --> err2["エラー表示:<br/>パスワードは8文字以上で入力してください"]
    err1 --> input
    err2 --> input
```

#### 3. タスク作成(正常系・異常系、親/子タスクの分岐)

```mermaid
flowchart TD
    start(["開始: S-03の「+ 新規タスク作成」<br/>または M-03の「子タスクを追加」"]) --> open["M-01(タスク作成モーダル)を開く"]
    open --> fromChild{"M-03から開いたか?"}
    fromChild -- Yes --> hint["「子タスクとして作成します」の案内を表示<br/>parent_idを引き継ぐ"]
    fromChild -- No --> normal["通常のタスク作成フォーム"]
    hint --> input["タイトル・詳細・想定の工数・締切日を入力"]
    normal --> input
    input --> submit[["作成ボタン押下<br/>TASK-02 POST /tasks"]]
    submit --> ok{"201 Created?"}
    ok -- Yes --> close["M-01を閉じ、呼び出し元(S-03またはM-03)に反映"]
    close --> fin(["終了"])
    ok -- "No(422 ESTIMATED_HOURS_REQUIRED)" --> err1["エラー表示:<br/>想定の工数は必須です"]
    ok -- "No(422 VALIDATION_ERROR)" --> err2["エラー表示:<br/>工数は0.5〜9999の範囲で入力してください"]
    err1 --> input
    err2 --> input
```

#### 4. タスク完了(正常系・異常系)

```mermaid
flowchart TD
    start(["開始: M-03の「編集」ボタン"]) --> open["M-02(タスク編集モーダル)を開く"]
    open --> change["statusをdoneに変更"]
    change --> input["実績の工数を入力"]
    input --> submit[["保存ボタン押下<br/>TASK-04 PATCH /tasks/{id}"]]
    submit --> ok{"200 OK?"}
    ok -- Yes --> reflect["completed_atが自動セットされM-03に反映"]
    reflect --> fin(["終了"])
    ok -- "No(422 ACTUAL_HOURS_REQUIRED)" --> err["エラー表示:<br/>実績の工数は必須です"]
    err --> input
```

#### 5. タスク削除(確認ダイアログ・子孫タスクの連動削除)

```mermaid
flowchart TD
    start(["開始: M-03の「削除」ボタン"]) --> confirm{"削除確認ダイアログ"}
    confirm -- キャンセル --> cancelEnd(["終了: 何も変わらない"])
    confirm -- 削除する --> submit[["TASK-05 DELETE /tasks/{id}"]]
    submit --> ok{"204 No Content?"}
    ok -- Yes --> hasChild{"子・孫タスクはあるか?"}
    hasChild -- Yes --> cascade["子・孫タスクも連動削除される<br/>(schema.sqlのON DELETE CASCADE、個別の確認は挟まない)"]
    hasChild -- No --> single["対象タスクのみ削除"]
    cascade --> close["M-03を閉じS-03の一覧から消える"]
    single --> close
    close --> fin(["終了"])
    ok -- "No(404 TASK_NOT_FOUND)" --> err["エラー表示: 既に削除済み、または他ユーザーのタスク"]
    err --> close
```

#### 6. セッション切れ・未ログインアクセス(准正常系)

```mermaid
flowchart TD
    start(["開始: S-03・M-01〜M-03のいずれかで<br/>画面表示・作成・保存などAPIを呼ぶ操作"]) --> hasToken{"トークンを保持しているか?"}
    hasToken -- No --> redirect
    hasToken -- Yes --> call[["APIを呼び出す"]]
    call --> status{"401 UNAUTHORIZED?"}
    status -- No --> show["画面を表示"]
    show --> fin1(["終了"])
    status -- Yes --> redirect["S-01へ強制リダイレクト"]
    redirect --> note["⚠ 未ログインとトークン期限切れは同じ401を返すため、<br/>フロントエンドはレスポンスだけでは区別できない<br/>(docs/assignment/step4.md「苦労したこと3.」参照)"]
    note --> fin2(["終了"])
```
