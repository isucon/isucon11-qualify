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

- **ISU**: この世界で愛されるあなたの大事なパートナー。いろいろな性格を持っていて、その時その時のコンディションを通知してくれる。すべての ISU にはあらかじめ JIA ISU ID が割り当てられている。
- **ユーザ**: ISUCONDITION に登録をしている人。
- **閲覧者**: ISUCONDITION に登録はしていないが、トップページでトレンドを見ている人。
- **JIA**: Japan ISU Association の略。この世界において日本の　ISU を取りまとめる団体。すべての ISU に固有の JIA ISU ID を割り当てて管理を行っている。
- **コンディション**: ISU から送られてくる情報。ISU が誰かに座られているか耐荷重を超えていないかと言った情報や、汚れていないか、壊れていないかなどを教えてくれる。
- **コンディションレベル**: ISU から送られた `is_dirty`、`is_overweight`、`is_broken` という 3 つの情報から決まる ISU の状態。それぞれの情報は問題が発生している場合は `true` になる。以下の 3 つのレベルが存在する。
    - **Info**: 一切、問題が発生していない状態。
    - **Warning**: 1〜2 つの問題が発生している状態。
    - **Critical**: 3 つの問題が発生している状態。
- **スコア**: コンディションから計算された1時間毎の点数。0 以上 100 以下の整数値をとる。
- **グラフ**: 24 時間分の ISU の状態を可視化したもの。
- **トレンド**: ISUCONDITION に登録されている ISU たちの、性格ごとの最新の "Info", "Warning", "Critical" の割合を表示したもの。

## ISUCONDITIONの機能とユーザ、ISU、閲覧者について

### ログイン

ISUCONDITION は認証を JIA に委ねており、ユーザは JIA の認証サイトで認証成功時に発行されるトークンを使って ISUCONDITION へのログインを行います。

ログインの処理は以下のような流れになります。

![ログインの動き](https://user-images.githubusercontent.com/210692/129871067-c6e56bb9-546b-4743-979e-859ab6ca8e0b.png)

1. ユーザは、ISUCONDITION のトップページにアクセスします。
2. ユーザが、ISUCONDITION のトップページにある "JIA のアカウントでログイン" のボタンを押下すると JIA 認証サイトへ遷移します。
3. ユーザが、JIA 認証サイトで JIA のアカウント情報を入力します。
4. JIA 認証サイトは認証成功時にトークン（JWT: JSON Web Token）を発行し、ユーザを ISUCONDITION にリダイレクトします。リダイレクトの URL にはクエリパラメータにトークンが設定されています。
5. ISUCONDITION はトークンが妥当なものかを検証します。
6. トークンの妥当性が確認された場合、ユーザは ISUCOONDITON のログインに成功します。

### ISUの登録

ユーザが、ISUCONDITION に ISU を登録することで、ISU から ISUCONDITION へのコンディション送信が開始されます。
コンディション送信を開始するには JIA による ISU のアクティベートが必要です。

ISU の登録は以下のような流れになります。

![ISUのアクティベートイメージ](https://user-images.githubusercontent.com/210692/129871576-269324a6-14be-452e-8826-2bb421196d37.png)

1. ユーザが ISUCONDITION に ISU の ID を入力します。
2. ISUCONDITION は　JIA の ISU 管理サービスに対して ISU の ID を送信します。
3. ISU 管理サービスは、対象の ISU にコンディション送信を開始するよう指示します（アクティベート）。
4. ISUCONDITION は ISU の情報を保存し登録が完了します。

#### JIA ISU ID

アプリケーションの動作確認には以下の JIA ISU ID を利用できます。

| JIA ISU ID                           |
|--------------------------------------|
| 3a8ae675-3702-45b5-b1eb-1e56e96738ea |
| 3efff0fa-75bc-4e3c-8c9d-ebfa89ecd15e |
| f67fcb64-f91c-4e7b-a48d-ddf1164194d0 |
| 32d1c708-e6ef-49d0-8ca9-4fd51844dcc8 |
| af64735c-667a-4d95-a75e-22d0c76083e0 |
| cb68f47f-25ef-46ec-965b-d72d9328160f |
| 57d600ef-15b4-43bc-ab79-6399fab5c497 |
| aa0844e6-812d-41d2-908a-eeb82a50b627 |

### ISUのコンディション送信処理

コンディションの送信先 URL はアクティベート時に、 ISUCONDITION が JSON で送信する `target_base_url` と `isu_uuid` により以下のように決定されます。

```
$target_base_url/api/condition/$isu_uuid
```

ISU はアクティベートされると、自身のコンディションを送信先 URL へ定期的に送信するようになります。

ISU から送信されるデータには 1 つ以上のコンディションが含まれます。送信されるコンディションは ISU 単位で下記が保証されています。

- コンディションの時刻情報が重複することはない。
- コンディションの時刻情報は、既に送られたコンディションの時刻情報よりも常に先の時刻となる。

ユーザはコンディションのデータの欠損を許容しますが、理想的には全てのコンディションのデータが保存されることを期待しています。

### 登録済みの ISU の確認

ユーザは、一定の間隔で自身が登録した ISU の一覧 (`GET /api/isu`) を確認しています。ユーザは ISU の一覧を見て、各ISUの詳細 (`GET /api/isu/:jia_isu_uuid`) を確認します。

### ISU の詳細確認

ユーザは、ISU の詳細ページから次のことを行います。

- コンディションの確認 (`GET /api/condition/:jia_isu_uuid`) 
- グラフの確認 (`GET /api/isu/:jia_isu_uuid/graph`)

#### コンディションの確認

ユーザは、最新の ISU のコンディションを確認後、まだ確認していないコンディションがある場合、過去に遡って ISU のコンディションを確認します。
またこのときコンディションレベルで表示するものを絞り込むことができます。

ISU のコンディションレベルは悪くなる事があり、ユーザは悪いコンディションレベル（`Info` 以外）を確認すると速やかに問題を改善します。
コンディションレベルが悪い ISU があった場合、ユーザは掃除や修理などで ISU のコンディションを改善します。
この改善は速やかに完了し、 ISU は次のコンディション送信からは改善されたコンディションを送信します。

#### グラフの確認

グラフは、指定した時刻から 24 時間分の ISU の状態を可視化したものです。
グラフは24時間のデータで構成されており、1つのデータは1時間ごとのコンディションを元に計算されます。グラフのデータには以下の情報が含まれます。


- スコア ( ( (Infoの数 * 3) + (Warning の数 * 2) + (Critical の数 * 1) ) * 100 / 3 / 含まれているコンディション数)
- ISU に座っていた時間の割合 (`is_sitting=true` の数 * 100 / 含まれているコンディション数)
- コンディションレベルの割合 
  - `is_dirty` の割合 (`is_dirty=true` の数 * 100 / 含まれているコンディション数)
  - `is_overweight` の割合 (`is_overweight=true` の数 * 100 / 含まれているコンディション数)
  - `is_broken` の割合 (`is_broken=true` の数 * 100 / 含まれているコンディション数)

ユーザは、最新のグラフを確認後、まだ確認していない過去のグラフがある場合に、過去に遡ってグラフを確認します。
まだ確認していなかったグラフを確認後、最後に確認したグラフの中からデータが存在する時間をランダムに 1 時間選択し、コンディションを確認します。

### トレンドの閲覧

トレンドは ISUCONDITION のサービスを知ってもらうための機能で、ログインしていないユーザ（閲覧者）が閲覧します。
トレンドでは ISUCONDITION に登録されているすべての ISU の最新のコンディションレベルが性格ごとにまとまっており、コンディションレベルの割合から、ISU が持つ性格ごとの傾向を見ることができます。

閲覧者は、**未ログイン状態** で　ISUCONDITION　のトップページに表示されるトレンド (`GET /api/trend`) を確認しています。
サービスに興味を持っている閲覧者はサービストップページに表示されるトレンドを一定間隔で閲覧し、トレンドの変化に注目しています。
閲覧者たちがトレンドの変化を一定回数確認するたびに、ユーザが 1 人増加します。また、閲覧者の行動中にエラー(タイムアウトを含む)があった場合、 1 回のエラーにつき閲覧者は 1 人減ります。
閲覧者はユーザが「[ISU のコンディション確認](#isu-のコンディション確認)」に書かれた処理を正しく 1 回行うと 1 人増えます。

## Japan ISU Association (JIA) の API 

現在 ISUCONDITION が利用している JIA の API は２つです。

- `POST /api/auth`
- `POST /api/activate`

詳細は下記のとおりです。

### `POST /api/auth`

JIA から認証トークン(JWT)を発行するためのエンドポイントであり、JIA の認証サイト `https://jia-auth.xi.isucon.dev` で提供されています。
認証に成功をすると JWT を生成して返します。

+ Request (application/json)
    + Schema

            {
                "user": "string",
                "password": "string"
            }

    + Attributes (object)
        | Field    | Type   | Required | Description        | Example   |
        |----------|--------|----------|--------------------|-----------|
        | user     | string | true     | ログインをするユーザ名 | `isucon`  |
        | password | string | true     | ログインパスワード    | `isucon`  |

+ Response 200 (text/plain)
    + Body

            eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjg1NjMxODksImlhdCI6MTYyODU2MTM4OSwiamlhX3VzZXJfaWQiOiJpc3Vjb24ifQ.MuIl1-kVe60DzwoGHj2yrck8QwYWDH_N20uCqNVR1IZiuo7ArYiBDbMdTbEzFbkN52x8SxGS3GvKoGuMmRfZXQ

+ Other Responsess
    + 400 (text/plain)
    + 401 (text/plain)

### `POST /api/activate`

JIA が管理する ISU に対して指定の URL に向けて、センサーデータを送るように指示するためのエンドポイントです。
アクティベートに成功すると、ISU は `target_base_url` で指定された URL に対しセンサーデータの送信を継続します。
レスポンスにはアクティベートされた ISU の性格が含まれます。

+ Request (application/json)
    + Schema

            {
                "target_base_url": "string",
                "isu_uuid": "string"
            }


    + Attributes (object)

        | Field           | Type   | Required | Description                | Example                                |
        |-----------------|--------|----------|----------------------------|----------------------------------------|
        | target_base_url | string | true     | ISU のコンディション送信先     | `https://isucondition-1.t.isucon.dev`  |
        | isu_uuid        | string | true     | JIA が発行する ISU の 固有ID  | `0694e4d7-dfce-4aec-b7ca-887ac42cfb8f` |


+ Response 202 (application/json)
    + Schema

            {
                "character": "string"
            }

    + Attributes (object)

        | Field     | Type   | Required | Description                 | Example    |
        |-----------|--------|----------|-----------------------------|------------|
        | character | string | true     | アクティベートされた ISU の性格 | `いじっぱり` |


+ Other Responsess
    + 400 (text/plain)
    + 403 (text/plain)
    + 500 (text/plain)


このエンドポイントは、開発/検証用に JIA API Mock という名前で開発者向けに同様の機能が提供されています。

注意点として、以下があります。

- `target_base_url` には下記の3つの FQDN のみを指定することができます。それ以外を指定した場合は JIA から `400 Bad Request` が返され ISU のアクティベートに失敗します。
  - `isucondition-1.t.isucon.dev`
  - `isucondition-2.t.isucon.dev`
  - `isucondition-3.t.isucon.dev`

### JIA API Mock について

JIA API Mock は、ISUCONDITION の開発用に用いられる JIA の API モックとして、サーバーのポート 5000 番で待ち受けます。
JIA API Mock は以下のエンドポイントと、 ISU からのコンディション送信を模擬したリクエストを送る機能を持っています。

- `POST /api/activate` - ISU の登録（アクティベート）
- 登録した ISU から ISUCONDITION へ向けたコンディションの送信

JIA API Mock は ISU を登録（アクティベート）すると、指定された URL へのコンディションの送信を開始します。
登録したISUからのコンディション送信は JIA API Mock を停止/再起動するまで続きますので、負荷走行前には JIA API Mock を停止/再起動することをお勧めします。
JIA API Mock のサービスを停止/再起動する場合は、 以下のコマンドを利用してください。

```shell
$ sudo systemctl [stop|restart] jiaapi-mock.service
```

負荷走行後の ISUCONDITION はベンチマーカーが設定した JIA のエンドポイントが設定されているため、上記の設定を行っていても ISUCONDITIONから `500 Internal Server Error` が返されるエンドポイントがあります。
負荷走行後に JIA API Mock を利用する際は、下記のように `POST /initialize` で JIA API Moc のエンドポイントを設定してください。

```
curl -sf -H 'content-type: application/json' https://<Elastic IP アドレス>/initialize -d '{"jia_service_url": "http://<Elastic IP アドレス>:5000"}'
```

##　ターミナルからの ISUCONDITION の動作確認

JIA からトークンを取得し、ISUCONDITION へ取得したトークンを使いログインを行い、cookie の設定を行うことで、ターミナルからも ISUCONDITION の動作確認が可能です。
以下、ターミナルからの動作確認方法の一例を示します。

### JIA API からの トークンの取得

```
$ TOKEN=`curl -sf -H 'content-type: application/json' https://jia-auth.xi.isucon.dev/api/auth -d '{"user": "isucon", "password": "isucon"}'`
```

JIA API に送信する `user` と `password` には、当日マニュアルに記載されているものを用いてください。
JIA API から発行されるトークンの有効期限は、発行から30分となります。

###  トークンを使った cookie の設定

```
$ curl -c cookie.txt -vf -XPOST -H "authorization: Bearer ${TOKEN}" https://<Elastic IP アドレス>/api/auth
```

一例として、cookie を使い `GET /api/isu` にアクセスします。

```
$ curl -b cookie.txt https://<Elastic IP アドレス>/api/isu
```

### ISU からのコンディションを受け取る `POST /api/condition/:jia_isu_uuid` の検証

ISU からのコンディションを受け取る `POST /api/condition/:jia_isu_uuid` は、
アクティベートを行った ISU であれば、以下のようにコンソールからもコンディションを受け取ることが可能です。
JIA ISU ID は、予選当日マニュアル 2 アプリケーションの動作確認 に記載されているものを利用してください。

```
$ export JIA_ISU_UUID=0694e4d7-dfce-4aec-b7ca-887ac42cfb8f
$ curl -XPOST -H 'content-type: application/json' https://<Elastic IP アドレス>/api/condition/${JIA_ISU_UUID} \
-d '[{"is_sitting": true, "condition": "is_dirty=true,is_overweight=true,is_broken=true","message":"test","timestamp": 1628492991}]'
```

#### JIA API Mock を用いたローカルでの開発方法

ローカル環境で開発している ISUCONDITION に対して JIA API Mock を利用した検証を行うためには、サーバー上の 5000 番ポートで待ち受けている JIA API Mock へアクセスできるようにするため、ポートフォワーディングが必要です。

以下に [SSH におけるローカルポートフォワーディング](https://help.ubuntu.com/community/SSH/OpenSSH/PortForwarding#Local_Port_Forwarding) を実行するコマンドを例示します。
これは「リモートホスト `isucon-server1` に SSH 接続をした上で」「ローカルの `localhost:5000` への TCP 接続を」「リモートホストを通して `isucondition-1.t.isucon.dev` へ転送する」というコマンドです。

```shell
$ ssh -L localhost:5000:isucondition-1.t.isucon.dev:5000 isucon-server1
```

上記のローカルポートフォワーディングの設定を行った場合、ログインや ISU の登録（アクティベート）の機能は利用可能となりますが、登録した ISU からローカル環境で開発している ISUCONDITION へ向けたコンディションの送信はできません。[ISU からのコンディションを受け取る `POST /api/condition/:jia_isu_uuid` の検証](#isu-からのコンディションを受け取る-post-apiconditionjia_isu_uuid-の検証) を参考に、コンソールから ISU のコンディション送信を行った検証を行ってください。
