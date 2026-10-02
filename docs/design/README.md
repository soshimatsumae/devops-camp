# 画面デザイン資料

画面のワイヤーフレームと画面フローマップ。元データはこのディレクトリのHTMLで、GitHub上ではHTMLが描画されないため、書き出したPNGを下に載せている。

| ファイル | 内容 |
| :---- | :---- |
| [wireframes.html](./wireframes.html) | 6画面(S-01〜S-03、M-01〜M-03)のワイヤーフレーム。タブで画面を切り替える |
| [flowmap.html](./flowmap.html) | ワイヤーフレームの画面を並べ、ユースケースごとの遷移を色分けした矢印でつないだフローマップ。凡例でフローを選ぶと、そのフローだけを強調表示できる |
| [images/](./images/) | 上の2つから書き出したPNG(GitHub表示用) |

HTMLをブラウザで操作したい場合は、リポジトリをcloneしてローカルで開く(例: `open docs/design/flowmap.html`)。

## 画像の更新手順

HTMLを編集したら、リポジトリのルートで次のコマンドを実行してPNGを書き出し直し、HTMLと一緒にコミットする(Google Chromeが必要)。

```sh
# docs/design/images/ 配下のPNGをすべて書き出し直す
./scripts/export-design-images.sh
```

## 画面フローマップ

丸い番号が操作の起点と手順の順番、実線の矢印が正常系、破線の矢印がログアウト・セッション切れ(准正常系)、画面の下の破線の枠が異常系(エラーを表示して同じ画面に留まる)を表す。分岐の論理は[docs/SCREEN_FLOW.md](../SCREEN_FLOW.md)の「ユースケース別フロー図」を参照。

![画面フローマップ](./images/flowmap.png)

## ワイヤーフレーム

各画面の目的・UI要素・対応APIは[docs/SCREEN_LIST.md](../SCREEN_LIST.md)を参照。

### S-01 ログイン画面

![S-01 ログイン画面](./images/wireframe-s01.png)

### S-02 新規登録画面

![S-02 新規登録画面](./images/wireframe-s02.png)

### S-03 ダッシュボード画面

![S-03 ダッシュボード画面](./images/wireframe-s03.png)

### M-01 タスク作成モーダル

![M-01 タスク作成モーダル](./images/wireframe-m01.png)

### M-02 タスク編集モーダル

![M-02 タスク編集モーダル](./images/wireframe-m02.png)

### M-03 タスク詳細モーダル

![M-03 タスク詳細モーダル](./images/wireframe-m03.png)
