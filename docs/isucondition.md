# ISUCONDITION アプリケーションマニュアル

![ISUCONDITIONロゴ](https://s3-ap-northeast-1.amazonaws.com/hackmd-jp1/uploads/upload_f7b1f89768906f3a6d8280f4814af567.png)


## ISUCONDITION とは

**"ISUとつくる新しい明日"**

あなたの大事なパートナーであるISUが教えてくれるコンディションから、コンディションレベルやその変化をスコアとして知ることで、大事なISUと長く付き合っていくためのサービスです。

## サービス紹介

[ISUCONDITION コマーシャル](TODO:リンク)

## 用語

- **ISU**: この世界で愛されるあなたの大事なパートナー。いろいろな性格を持っていて、その時その時のコンディションを通知してくれる。
- **ユーザ**: ISUCONDITIONに登録をしている人。
- **閲覧者**: ISUCONDITIONに登録はしていないが、Topページでトレンドを見ている人。
- **JIA**: Japan ISU Associationの略。この世界において日本のイスを取りまとめる団体。
- **コンディション**: ISUから送られてくる情報。ISUが誰かに座られているか、どんな状態かなどを教えてくれる。
- **コンディションレベル**: 
    - **Info**: 良い状態。何も問題が起きていない状態。
    - **Warning**: 1〜2個問題が起きている状態。
    - **Critical**: 3個以上の問題が起きている状態。
- **スコア**: コンディションから計算された1時間毎の点数（0~100）
- **グラフ**: 1日の中でスコアが遷移した記録
- **トレンド**: ISUCONDITIONに登録されているISUたちの、性格ごとの最新の状態を表示したもの。

## ISUCONDITIONの機能とユーザ、ISU、閲覧者について

### 1. ユーザ登録/ログイン

![ログインの動き](https://user-images.githubusercontent.com/210692/128741078-dfa3c27c-52d0-4a84-bfba-1ae7ad90019f.png)

ユーザは、トップページ (`/`) からユーザ登録/ログインを行います。
"JIAのアカウントでログイン" のボタンを押下するとJIAのページへ遷移します。
JIAのページでJIAのアカウントを利用してログインを行い、そこで得たトークン（JWT:JSON Web Token）を元にISUCONDITIONへアクセスを行います。
ISUCONDITIONは送られてきたJWTが正しいかを検証し、正しいと判断した場合にSession IDをユーザに返します。

### 2. ISUの登録とISUのコンディション送信処理

![ISUのアクティベートイメージ](https://user-images.githubusercontent.com/210692/128740543-d4cbc6b8-d1b3-46ee-9e6b-9d660648edd5.png)

ユーザは、自分の大事なパートナーであるISUを、JIAが発行するISU固有のID(以下、ISU UUID)を使い、ISUCONDITIONに登録 (`/register`) します。ユーザは、登録を行ったISUの詳細 (`/isu/:jia_isu_uuid`) や、ISUのコンディション (`/isu/:jia_isu_uuid/condition`) 、グラフとスコア (`/isu/:jia_isu_uuid/graph`) を見ることができます。

JIA は ISUCONDITION から ISU のアクティベートリクエストを受け取ることで、 対象のISU が特定の URL へのコンディション送信 (`POST /api/condition/:jia_isu_uuid`) を開始するよう指示します。
コンディション送信先 URL はアクティベート時に ISUCONDITION が JSON で送信する `target_base_url` と `isu_uuid` により以下のように決定されます。

```
$target_base_url/api/condition/$isu_uuid
```

注意点として、以下の2点があります。

- `target_base_url` を変更することで ISU がコンディションを送る先を変更することが可能ですが、既に登録済みの ISU には反映されません。
- `target_base_url` に含まれる FQDN に `isucondition-[1-3].t.isucon.dev` 以外を指定した場合 ISU のアクティベートに失敗します。

なお上記の `target_base_url` は環境変数 `POST_CONDITION_TARGET_BASE_URL` で指定されています。

ISUは、JIAから送信開始の指示を受け取った時点から、自身のコンディションをISUCONDITIONに対して送信 (`POST /api/condition/:jia_isu_uuid`) を続けます。ISUのコンディションは悪くなる事があり、ユーザが改善を行わない限りコンディションが良くなる事はありません。

ISUから定期的に送信されるデータには複数のコンディションが含まれます。
コンディションにはコンディションが記録された時間が含まれますが、この時間は過去に戻ることはありません。また、1つのISUが同一時間のコンディションを送信することはありません。

ISUCONDITIONは、ISUから送信されるコンディション (`POST /api/condition/:jia_isu_uuid`) データを保持しますが、アプリケーションが高負荷な場合データを保存せず `503 Service Unavailable` を返すことを許容しています。

### 3. 登録済みのISUの確認

ユーザは、一定の頻度でログイン後のトップページ (`/`) に表示される、自身が登録したISUの一覧を確認しています。ユーザはISUの一覧を受け取ったとき、各ISUの詳細 (`/isu/:jia_isu_uuid`) を確認します。
他のユーザのISUについて見ることできません。

### 4. ISUのコンディション確認

ユーザは、ISUの詳細を確認後に「グラフの確認」、「コンディションの確認」、「コンディションレベルが悪いISUの改善」を行います。
これらの3つ処理が正しく1回行われると、閲覧者が1人増えます。（MEMO: 現在の仕様であり変更が入るかも）

#### 4.1. グラフの確認

ユーザは、当日のグラフを確認 (`/isu/:jia_isu_uuid/graph`) 後、まだ確認していない日付のグラフがある場合、過去日付に遡ってグラフを確認します。
未読のグラフ確認後、最後に見ていたグラフのデータが存在する時間をランダムに1時間選択し、コンディションを確認します。

#### 4.2. コンディションの確認

ユーザは、ISUのコンディションを確認 (`/isu/:jia_isu_uuid/condition`) 後、まだ確認していないコンディションがある場合、過去に遡ってコンディションを確認します。

#### 4.3. コンディションレベルが悪いISUの改善

ユーザは、ISUのコンディションレベルが悪く(`warning`, `critical`)なっていないかを確認し、コンディションレベルが悪いISUがあった場合に掃除や修理などでISUのコンディションを改善します。

ユーザによりコンディションの改善が行われたISUは、次のコンディション送信からは改善されたコンディションを送信します。

### 5. トレンド

閲覧者は、**"未ログイン状態"** でISUCONDITIONのトップページ (`/`) に表示されるトレンドを確認しています。
トレンドではISUCONDITIONに登録されているすべてのISUたちの最新のコンディションが取得できます。
トレンドで取得されるコンディションはISUが持つ性格ごとにまとまった形となっています。

サービスに興味を持った閲覧者はサービストップページに表示されるトレンドを一定間隔で閲覧し、トレンドの変化に注目しています。
閲覧者がトレンドの変化を1回確認するたびに、ユーザが1人増加します。（MEMO: 現在の仕様であり変更が入るかも）

トレンドはISUCONDITIONのサービスを知ってもらうために、ユーザ以外にも公開されているためログインは不要です。

## Japan ISU Association (JIA) の API 

現在JIAが公開しているAPIは以下の２つです

### /api/activate [POST]

JIAが管理するISUに対して指定のURLに対し、センサーデータを送るように指示するためのエンドポイント。
アクティベートに成功すると、ISUは `target_base_url` で指定されたURLに対しセンサーデータの送信を継続する。
レスポンスにはアクティベートされたISUの性格が含まれる。

+ Request (application/json)
    + Attributes (object)
        + target_base_url: `https://isucondition-1.t.isucon.dev:3000` (string, required) - ISUCONDITIONのサービスURL
        + isu_uuid: `0694e4d7-dfce-4aec-b7ca-887ac42cfb8f` (string, required) - JIAが発行するISUのID

    + Schema

            {
                "target_base_url": "string",
                "isu_uuid": "string"
            }


+ Response 202 (application/json)
    + Attributes (object)
        + character: `いじっぱり` (string) - アクティベートされたISUの性格

    + Schema

            {
                "character": "string"
            }
- Response 400 (text/plain)
- Response 403 (text/plain)
- Response 500 (text/plain)

### /api/auth [POST]

JIAへログインするためのエンドポイント。
ログインに成功をするとJWTを生成して返す。

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
