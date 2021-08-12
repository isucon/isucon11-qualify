# ISUCONDITION アプリケーションマニュアル

![ISUCONDITIONロゴ](https://s3-ap-northeast-1.amazonaws.com/hackmd-jp1/uploads/upload_f7b1f89768906f3a6d8280f4814af567.png)


## ISUCONDITION とは

**"ISUとつくる新しい明日"**

あなたの大事なパートナーであるISUが教えてくれるコンディションから、コンディションレベルやその変化をスコアとして知ることで、大事なISUと長く付き合っていくためのサービスです。

## サービス紹介

[ISUCONDITION コマーシャル](TODO:リンク)

## 用語

- **ISU**: この世界で愛されるあなたの大事なパートナー。いろいろな性格を持っていて、その時その時のコンディションを通知してくれる。
- **ユーザ**: ISUCONDITION に登録をしている人。
- **閲覧者**: ISUCONDITION に登録はしていないが、Topページでトレンドを見ている人。
- **JIA**: Japan ISU Association の略。この世界において日本のイスを取りまとめる団体。
- **コンディション**: ISU から送られてくる情報。ISUが誰かに座られているか、どんな状態かなどを教えてくれる。
- **コンディションレベル**: 
    - **Info**: 良い状態。何も問題が起きていない状態。
    - **Warning**: 1〜2 個問題が起きている状態。
    - **Critical**: 3 個以上の問題が起きている状態。
- **スコア**: コンディションから計算された1時間毎の点数（0~100）
- **グラフ**: 1 日の中でスコアが遷移した記録
- **トレンド**: ISUCONDITION に登録されている ISU たちの、性格ごとの最新の状態を表示したもの。

## ISUCONDITIONの機能とユーザ、ISU、閲覧者について

### 1. ユーザ登録/ログイン

![ログインの動き](https://user-images.githubusercontent.com/210692/128741078-dfa3c27c-52d0-4a84-bfba-1ae7ad90019f.png)

ユーザは、ISUCONDITION のトップページからユーザ登録/ログインを行います。
"JIA のアカウントでログイン" のボタンを押下すると JIA のページへ遷移します。
JIA のページで JIA のアカウントを利用してログインを行い、そこで得たトークン（JWT: JSON Web Token）を元に ISUCONDITION でログイン (`POST /api/auth`) を行います。
ISUCONDITION は送られてきたJWTが正しいかを検証し、正しいと判断した場合にSession IDをユーザに返します。

### 2. ISUの登録とISUのコンディション送信処理

![ISUのアクティベートイメージ](https://user-images.githubusercontent.com/210692/128740543-d4cbc6b8-d1b3-46ee-9e6b-9d660648edd5.png)

ユーザは、自分の大事なパートナーである ISU を、JIA が発行する ISU 固有の ID (以下、ISU UUID)を使い、ISUCONDITION に登録リクエスト (`POST /api/isu`) を送ります。
ISUCONDITION はユーザから ISU の登録リクエストを受け取った場合 JIA に対して ISU のアクティベーションリクエスト (`POST /api/activate`) を送信します。
JIA は ISUCONDITION から ISU のアクティベートリクエストを受け取ることで、 対象の ISU が特定の URL へのコンディション送信 (`POST /api/condition/:jia_isu_uuid`) を開始するよう指示します。
コンディションの送信先 URL はアクティベート時に ISUCONDITION が JSON で送信する `target_base_url` と `isu_uuid` により以下のように決定されます。

```
$target_base_url/api/condition/$isu_uuid
```

注意点として、以下の2点があります。

- `target_base_url` を変更することで ISU がコンディションを送る先を変更することが可能ですが、既に登録済みの ISU には反映されません。
- `target_base_url` に設定する FQDN に下記 3 つ以外を指定した場合 JIA から `400 Bad Request` が返され ISU のアクティベートに失敗します。
  - `isucondition-1.t.isucon.dev`
  - `isucondition-2.t.isucon.dev`
  - `isucondition-3.t.isucon.dev`

なお上記の `target_base_url` は環境変数 `POST_CONDITION_TARGET_BASE_URL` で指定されています。

ISU は、JIA から送信開始の指示を受け取った時点から、自身のコンディションを ISUCONDITION に対して送信リクエスト (`POST /api/condition/:jia_isu_uuid`) を続けます。ISU のコンディションは悪くなる事があり、ユーザが改善を行わない限り ISU のコンディションが良くなる事はありません。

ISU から定期的に送信されるデータには複数のコンディションが含まれます。
コンディションにはコンディションが記録された時間が含まれますが、この時間は過去に戻ることはありません。また、1つの ISU が同一時間のコンディションを送信することはありません。

ISUCONDITION は、ISU から送信されるコンディションのデータを保持しますが、アプリケーションが高負荷な場合データを保存せず `503 Service Unavailable` を返すことを許容しています。

### 3. 登録済みの ISU の確認

ユーザは、一定の間隔で自身が登録した ISU の一覧 (`GET /api/isu`) を確認しています。ユーザは ISU の一覧を受け取ったとき、各ISUの詳細 (`GET /api/isu/:jia_isu_uuid`) を確認します。
他のユーザの ISU について見ることできません。

### 4. ISU のコンディション確認

ユーザは、ISU の詳細を確認後に以下の3つの確認や検索を行っています。

- グラフの確認 (`GET /api/isu/:jia_isu_uuid/graph`)
- コンディションの確認 (`GET /api/condition/:jia_isu_uuid`) 
- コンディションレベルが悪いISUの検索 (`GET /api/condition/:jia_isu_uuid?condition_level=critical,warning`) 

これらの処理が正しく 1 回行われると、閲覧者が 1 人増えます。（MEMO: 現在の仕様であり変更が入るかも）

#### 4.1. グラフの確認

ユーザは、最新のグラフを確認後、まだ確認していない日付のグラフがある場合に、過去日付に遡ってグラフを確認します。
まだ確認していなかったグラフを確認後、最後に確認したグラフの中からデータが存在する時間をランダムに 1 時間選択し、コンディションを確認します。

#### 4.2. コンディションの確認

ユーザは、最新の ISU のコンディションを確認後、まだ確認していないコンディションがある場合、過去に遡って ISU のコンディションを確認します。

#### 4.3. コンディションレベルが悪い ISU の検索

ユーザは、ISU のコンディションレベルが悪く (`warning`, `critical`) なっていないかを確認しています。コンディションレベルが悪い ISU があった場合、ユーザは掃除や修理などで ISU のコンディションを改善します。
この改善はユーザがコンディションレベルが悪い ISU を確認するとすぐに行われ、 ISU は次のコンディション送信からは改善されたコンディションを送信します。

### 5. トレンド

閲覧者は、**"未ログイン状態"** で　ISUCONDITION　のトップページに表示されるトレンド (`GET /api/trend`) を確認しています。
トレンドでは ISUCONDITION に登録されているすべての ISU たちの最新のコンディションが取得できます。
トレンドで取得されるコンディションは ISU が持つ性格ごとにまとまった形となっています。

サービスに興味を持った閲覧者はサービストップページに表示されるトレンドを一定間隔で閲覧し、トレンドの変化に注目しています。
閲覧者がトレンドの変化を 500 回確認するたびに、ユーザが 1 人増加します。（MEMO: 現在の仕様であり変更が入るかも）

トレンドは ISUCONDITION のサービスを知ってもらうために、ユーザ以外にも公開されているためログインは不要です。

## Japan ISU Association (JIA) の API 

現在 JIA が公開している API は以下の２つです。

### /api/activate [POST]

JIA が管理する ISU に対して指定の URL に向けて、センサーデータを送るように指示するためのエンドポイント。
アクティベートに成功すると、ISU は `target_base_url` で指定された URL に対しセンサーデータの送信を継続します。
レスポンスにはアクティベートされた ISU の性格が含まれます。

+ Request (application/json)
    + Attributes (object)
        + target_base_url: `https://isucondition-1.t.isucon.dev:3000` (string, required) - ISUCONDITION の サービスURL
        + isu_uuid: `0694e4d7-dfce-4aec-b7ca-887ac42cfb8f` (string, required) - JIA が発行する ISU の ID

    + Schema

            {
                "target_base_url": "string",
                "isu_uuid": "string"
            }


+ Response 202 (application/json)
    + Attributes (object)
        + character: `いじっぱり` (string) - アクティベートされた ISU の性格

    + Schema

            {
                "character": "string"
            }
- Response 400 (text/plain)
- Response 403 (text/plain)
- Response 500 (text/plain)

### /api/auth [POST]

JIA へログインするためのエンドポイント。
ログインに成功をすると JWT を生成して返します。

+ Request (application/json)
    + Attributes (object)
        + user: `isucon` (string, required) - ログインをするユーザ名
        + password: `isucon` (string, required) - ログインパスワード

    + Schema

            {
                "user": "string",
                "password": "string"
            }

+ Response 200 (text/plain)
    + Body

            eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjg1NjMxODksImlhdCI6MTYyODU2MTM4OSwiamlhX3VzZXJfaWQiOiJpc3Vjb24ifQ.MuIl1-kVe60DzwoGHj2yrck8QwYWDH_N20uCqNVR1IZiuo7ArYiBDbMdTbEzFbkN52x8SxGS3GvKoGuMmRfZXQ

+ Response 400 (text/plain)
+ Response 401 (text/plain)
