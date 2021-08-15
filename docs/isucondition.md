# ISUCONDITION アプリケーションマニュアル

![ISUCONDITIONロゴ](https://s3-ap-northeast-1.amazonaws.com/hackmd-jp1/uploads/upload_f7b1f89768906f3a6d8280f4814af567.png)


## ISUCONDITION とは

**"ISU とつくる新しい明日"**

あなたの大事なパートナーであるISUが教えてくれるコンディションから、コンディションレベルやその変化をスコアとして知ることで、大事なISUと長く付き合っていくためのサービスです。

### ストーリー

20xx年、政府が働く人々にリモートワークを推奨したことにより、家での仕事を支える存在として ISU が大事にされるようになりました。
働く人々が ISU に愛着を持って大事にするようになった結果、大事な ISU のコンディションを知ることで ISU を理解し、ISU と長く付きあっていきたいと人々は願うようになりました。
ISUCONDITION はこうした人々のニーズに応えるサービスとしてリリース目前です。しかし、パフォーマンスに大きな問題を抱えていました。
あなたは ISUCONDITION の開発者として 18 時までにこの問題の改善し、人と ISU が作る新しい明日を支えなければなりません。

## 用語

- **ISU**: この世界で愛されるあなたの大事なパートナー。いろいろな性格を持っていて、その時その時のコンディションを通知してくれる。
- **ユーザ**: ISUCONDITION に登録をしている人。
- **閲覧者**: ISUCONDITION に登録はしていないが、トップページでトレンドを見ている人。
- **JIA**: Japan ISU Association の略。この世界において日本のイスを取りまとめる団体。
- **コンディション**: ISU から送られてくる情報。ISU が誰かに座られているか耐荷重を超えていないかと言った情報や、汚れていないか、壊れていないかなどを教えてくれる。
- **コンディションレベル**: ISU から送られた "is_dirty"、"is_overweight"、"is_broken" という3つの情報から決まる ISU の状態。以下の3つのレベルが存在します。
    - **Info**: "is_dirty"、"is_overweight"、"is_broken" 3つ全てで問題が起きていない状態。
    - **Warning**: "is_dirty"、"is_overweight"、"is_broken" 3 つの情報のうち 1〜2 つで問題が起きている状態。
    - **Critical**: "is_dirty"、"is_overweight"、"is_broken" 3 つ全てで問題が起きている状態。
- **スコア**: コンディションから計算された1時間毎の点数。1 以上 3 以下の整数値をとります。
- **グラフ**: 1 日の中でスコアが遷移した記録
- **トレンド**: ISUCONDITION に登録されている ISU たちの、性格ごとの最新の "Info", "Warning", "Critical" の割合を表示したもの。

## ISUCONDITONの機能とユーザ、ISU、閲覧者について

### 1. ユーザ登録/ログイン

ISUCONDITION はユーザの登録/ログインを JIA に委ねており、ユーザは JIA へログイン時に発行されるトークンを使って ISUCONDITION へのユーザ登録/ログインを行います。

ユーザ登録/ログインの処理は以下のような流れになります。

![ログインの動き](https://user-images.githubusercontent.com/210692/129367327-ff05fb22-46fe-4982-9b9b-b4b72613a6f2.png)

1. ユーザは、ISUCONDITION のトップページにアクセスします。
2. ISUCONDITION のトップページにある "JIA のアカウントでログイン" のボタンを押下すると JIA のページへ遷移します。
3. JIA のページで JIA のアカウントを利用してログインを行います
4. ログイン成功時にトークン（JWT: JSON Web Token）が発行され ISUCONDITION に転送されます。
5. ISUCONDITION はトークンが妥当なものかを検証します。
6. トークンの妥当性が確認された場合ログイン成功。

### 2. ISUの登録とISUのコンディション送信処理

ユーザが、ISUCONDITION に ISU を登録することで、ISU  から ISUCONDITION へのコンディション送信が開始されます。

ISU の登録は以下のような流れになります。

![ISUのアクティベートイメージ](https://user-images.githubusercontent.com/210692/129368206-8130c782-b7a5-44ed-8084-c370feab6a4b.png)

<<<<<<< HEAD
ユーザは、自分の大事なパートナーであるISUを、JIAが発行するISU固有のID(以下、ISU UUID)を使い、ISUCONDITIONに登録 (`/register`) します。ユーザは、登録を行ったISUの詳細 (`/isu/:jia_isu_uuid`) や、ISUのコンディション (`/isu/:jia_isu_uuid/condition`) 、グラフとスコア (`/isu/:jia_isu_uuid/graph`) を見ることができます。
=======
1. ISUCONDITION はユーザから ISU の登録リクエストを受け取った場合 JIA に対して ISU のアクティベーションリクエストを送信します。
2. JIA は ISUCONDITION から ISU のアクティベートリクエストを受け取ることで、 対象の ISU にコンディション送信を開始するよう指示します。
3. コンディションの送信先 URL はアクティベート時に ISUCONDITION が JSON で送信する `target_base_url` と `isu_uuid` により以下のように決定されます。
>>>>>>> 08.15.1

JIAは、ISUCONDITIONからISU UUIDを受け取ることで、当該のISU UUIDを持つISUに対してISUCONDITIONへコンディション送信 (`POST /api/condition/:jia_isu_uuid`) を開始する指示をします。

ISUは、JIAから送信開始の指示を受け取った時点から、自身のコンディションをISUCONDITIONに対して送信 (`POST /api/condition/:jia_isu_uuid`) を続けます。ISUのコンディションは悪くなる事があり、ユーザが改善を行わない限りコンディションが良くなる事はありません。

<<<<<<< HEAD
ISUから定期的に送信されるデータには複数のコンディションが含まれます。
コンディションにはコンディションが記録された時間が含まれますが、この時間は過去に戻ることはありません。また、1つのISUが同一時間のコンディションを送信することはありません。
=======
- `target_base_url` を変更することで ISU がコンディションを送る先を変更することが可能ですが、既に登録済みの ISU には反映されません。
- `target_base_url` には下記の3つの FQDN のみを指定することができます。それ以外を指定した場合は JIA から `400 Bad Request` が返され ISU のアクティベートに失敗します。
  - `isucondition-1.t.isucon.dev`
  - `isucondition-2.t.isucon.dev`
  - `isucondition-3.t.isucon.dev`
>>>>>>> 08.15.1

ISUCONDITIONは、ISUから送信されるコンディション (`POST /api/condition/:jia_isu_uuid`) データを保持しますが、アプリケーションが高負荷な場合データを保存せず `503 Service Unavailable` を返すことを許容しています。

<<<<<<< HEAD
### 3. 登録済みのISUの確認

ユーザは、一定の頻度でログイン後のトップページ (`/`) に表示される、自身が登録したISUの一覧を確認しています。ユーザはISUの一覧を受け取ったとき、各ISUの詳細 (`/isu/:jia_isu_uuid`) を確認します。
他のユーザのISUについて見ることできません。

### 4. ISUのコンディション確認

ユーザは、ISUの詳細を確認後に「グラフの確認」、「コンディションの確認」、「コンディションレベルが悪いISUの改善」を行います。
これらの3つ処理が正しく1回行われると、閲覧者が1人増えます。（MEMO: 現在の仕様であり変更が入るかも）
=======
ISU は、JIA から送信開始の指示を受け取った時点から、自身のコンディションを ISUCONDITION に送信するリクエスト (`POST /api/condition/:jia_isu_uuid`) を続けます。ISU のコンディションレベルは悪くなる事があり、ユーザが改善を行わない限り ISU のコンディションレベルが良くなる事はありません。

ISU から定期的に送信されるデータには複数のコンディションが含まれます。
コンディションにはコンディションが記録された時刻情報が含まれますが、この時刻情報は、既に送られたコンディションの時刻情報よりも過去の時刻となることはありません。また、1つの ISU が同一時刻のコンディションを複数送信することはありません。

ISUCONDITION は、ISU から送信されるコンディションのデータを保持しますが、アプリケーションの負荷を下げるためにデータを保存せず `503 Service Unavailable` を返すことがあります。
ユーザはコンディションのデータの欠損を許容しますが、理想的には全てのコンディションのデータが保存されることを期待しています。

### 3. 登録済みの ISU の確認

ユーザは、一定の間隔で自身が登録した ISU の一覧 (`GET /api/isu`) を確認しています。ユーザは ISU の一覧を受け取ったとき、各ISUの詳細 (`GET /api/isu/:jia_isu_uuid`) を確認します。
他のユーザの ISU について見ることできません。

### 4. ISU のコンディション確認

ユーザは、ISU の詳細ページから次のことを行っています。

- コンディションの確認、コンディションレベルが悪い ISU の検索 (`GET /api/condition/:jia_isu_uuid`) 
- グラフの確認 (`GET /api/isu/:jia_isu_uuid/graph`)

#### 4.1. コンディションの確認
>>>>>>> 08.15.1

ユーザは、最新の ISU のコンディションを確認後、まだ確認していないコンディションがある場合、過去に遡って ISU のコンディションを確認します。

<<<<<<< HEAD
ユーザは、当日のグラフを確認 (`/isu/:jia_isu_uuid/graph`) 後、まだ確認していない日付のグラフがある場合、過去日付に遡ってグラフを確認します。
未読のグラフ確認後、最後に見ていたグラフのデータが存在する時間をランダムに1時間選択し、コンディションを確認します。
=======
#### 4.2. コンディションレベルが悪い ISU の検索
>>>>>>> 08.15.1

ユーザは、コンディションレベルを指定して検索をする機能を利用し、状態の悪い ISU がいないかを調べます。コンディションレベルが悪い ISU があった場合、ユーザは掃除や修理などで ISU のコンディションを改善します。
この改善は速やかに完了し、 ISU は次のコンディション送信からは改善されたコンディションを送信します。

<<<<<<< HEAD
ユーザは、ISUのコンディションを確認 (`/isu/:jia_isu_uuid/condition`) 後、まだ確認していないコンディションがある場合、過去に遡ってコンディションを確認します。

#### 4.3. コンディションレベルが悪いISUの改善

ユーザは、ISUのコンディションレベルが悪く(`warning`, `critical`)なっていないかを確認し、コンディションレベルが悪いISUがあった場合に掃除や修理などでISUのコンディションを改善します。

ユーザによりコンディションの改善が行われたISUは、次のコンディション送信からは改善されたコンディションを送信します。
=======
#### 4.3. グラフの確認

ユーザは、最新のグラフを確認後、まだ確認していない日付のグラフがある場合に、過去日付に遡ってグラフを確認します。
まだ確認していなかったグラフを確認後、最後に確認したグラフの中からデータが存在する時間をランダムに 1 時間選択し、コンディションを確認します。

#### 5. トレンドの閲覧
>>>>>>> 08.15.1

トレンドは ISUCONDITION のサービスを知ってもらうための機能で、ログインしていないユーザ（閲覧者）が閲覧します。
トレンドでは ISUCONDITION に登録されているすべての ISU の最新のコンディションレベルが性格ごとにまとまっており、コンディションレベルの割合から、ISU が持つ性格ごとの傾向を見ることができます。

<<<<<<< HEAD
閲覧者は、**"未ログイン状態"** でISUCONDITIONのトップページ　(`/`) に表示されるトレンドを確認しています。
トレンドではISUCONDITIONに登録されているすべてのISUたちの最新のコンディションが取得できます。
トレンドで取得されるコンディションはISUが持つ性格ごとにまとまった形となっています。

サービスに興味を持った閲覧者はサービストップページに表示されるトレンドを一定間隔で閲覧し、トレンドの変化に注目しています。
閲覧者がトレンドの変化を1回確認するたびに、ユーザが1人増加します。（MEMO: 現在の仕様であり変更が入るかも）

トレンドはISUCONDITIONのサービスを知ってもらうために、ユーザ以外にも公開されているためログインは不要です。

## Japan ISU Association (JIA) の API 

現在JIAが公開しているAPIは以下の２つです
=======
閲覧者は、**"未ログイン状態"** で　ISUCONDITION　のトップページに表示されるトレンド (`GET /api/trend`) を確認しています。
サービスに興味を持っている閲覧者はサービストップページに表示されるトレンドを一定間隔で閲覧し、トレンドの変化に注目しています。
閲覧者たちがトレンドの変化を一定回数確認するたびに、ユーザが 1 人増加します。また、閲覧者の行動中にエラー(タイムアウトを含む)があった場合、 1 回のエラーにつき閲覧者は 1 人減ります。
閲覧者はユーザが「4. ISU のコンディション確認」に書かれた処理を正しく 1 回行うと 1 人増えます。

## Japan ISU Association (JIA) の API 

JIA はブラウザからトップページ (`GET /`) へアクセスすることができますが、それ以外にも API を提供しています。
現在 ISUCONDITION が利用している JIA　の API は以下の２つです。JIA のユーザ登録については ISUCONDITION 側では取り扱わないため、本アプリケーションマニュアルでは記載しません。
>>>>>>> 08.15.1

### `POST /api/activate`

JIAが管理するISUに対して指定のURLに対し、センサーデータを送るように指示するためのエンドポイント。
アクティベートに成功すると、ISUは `target_base_url` で指定されたURLに対しセンサーデータの送信を継続する。
レスポンスにはアクティベートされたISUの性格が含まれる。

+ Request (application/json)
<<<<<<< HEAD
    + Attributes (object)
        + target_base_url: `http://localhost:3000` (string, required) - ISUCONDITIONのサービスURL
        + isu_uuid: `0694e4d7-dfce-4aec-b7ca-887ac42cfb8f` (string, required) - JIAが発行するISUのID

=======
>>>>>>> 08.15.1
    + Schema

            {
                "target_base_url": "string",
                "isu_uuid": "string"
            }


    + Attributes (object)
<<<<<<< HEAD
        + character: `いじっぱり` (string) - アクティベートされたISUの性格
=======
>>>>>>> 08.15.1

        | Field           | Type   | Reqyured | Description                | Example                              |
        |-----------------|--------|----------|----------------------------|--------------------------------------|
        | target_base_url | string | true     | ISU のコンディション送信先     | https://isucondition-1.t.isucon.dev  |
        | isu_uuid        | string | true     | JIA が発行する ISU の 固有ID  | 0694e4d7-dfce-4aec-b7ca-887ac42cfb8f |


+ Response 202 (application/json)
    + Schema

            {
                "character": "string"
            }

    + Attributes (object)

        | Field     | Type   | Reqyured | Description                 | Example  |
        |-----------|--------|----------|-----------------------------|----------|
        | character | string | true     | アクティベートされた ISU の性格 | いじっぱり |



- Response 400 (text/plain)
- Response 403 (text/plain)
- Response 500 (text/plain)

### `POST /api/auth`

<<<<<<< HEAD
JIAへログインするためのエンドポイント。
ログインに成功をするとJWTを生成して返す。
=======
JIA から認証トークン(JWT)を発行するためのエンドポイント。
認証に成功をすると JWT を生成して返します。
>>>>>>> 08.15.1

+ Request (application/json)
    + Schema

            {
                "user": "string",
                "password": "string"
            }

    + Attributes (object)
        | Field    | Type   | Reqyured | Description        | Example |
        |----------|--------|----------|--------------------|---------|
        | user     | string | true     | ログインをするユーザ名 | isucon  |
        | password | string | true     | ログインパスワード    | isucon  |

+ Response 200 (text/plain)
    + Body

            eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjg1NjMxODksImlhdCI6MTYyODU2MTM4OSwiamlhX3VzZXJfaWQiOiJpc3Vjb24ifQ.MuIl1-kVe60DzwoGHj2yrck8QwYWDH_N20uCqNVR1IZiuo7ArYiBDbMdTbEzFbkN52x8SxGS3GvKoGuMmRfZXQ

+ Response 400 (text/plain)
+ Response 401 (text/plain)
