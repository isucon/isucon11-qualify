# ISUCONDITION アプリケーションマニュアル

![ISUCONDITIONロゴ](https://s3-ap-northeast-1.amazonaws.com/hackmd-jp1/uploads/upload_f7b1f89768906f3a6d8280f4814af567.png)


## ISUCONDITION とは

**"ISU とつくる新しい明日"**

あなたの大事なパートナーである ISU が教えてくれるコンディションから、そのメッセージやコンディションレベルを知ることで、大事な ISU と長く付き合っていくためのサービスです。

### ストーリー

20xx 年、政府が働く人々にリモートワークを推奨したことにより、家での仕事を支える存在として ISU が大事にされるようになりました。
働く人々が ISU に愛着を持って大事にするようになった結果、大事な ISU のコンディションを知ることで ISU を理解し、ISU と長く付きあっていきたいと人々は願うようになりました。
ISUCONDITION はこうした人々のニーズに応えるサービスとしてリリース目前です。しかし、パフォーマンスに大きな問題を抱えていました。
あなたは ISUCONDITION の開発者としてリリースまでにこの問題を改善し、人と ISU が作る新しい明日を支えなければなりません。

## 用語

- **ISU**: この世界で愛されるあなたの大事なパートナー。いろいろな性格を持っていて、その時その時のコンディションを教えてくれる。すべての ISU にはあらかじめ JIA ISU ID が割り当てられている。
- **ユーザ**: ISUCONDITION に登録している人。
- **閲覧者**: ISUCONDITION に登録していないが、トップページでトレンドを見ている人。
- **JIA**: Japan ISU Association の略。この世界において日本の ISU を取りまとめる団体。すべての ISU に固有の JIA ISU ID を割り当てて管理を行っている。
- **コンディション**: ISU からのメッセージやその時の ISU の状態に関する情報。
- **コンディションレベル**: コンディション内の `is_dirty`、`is_overweight`、`is_broken` という 3 つの情報から決まる ISU の状態。それぞれの情報は、問題が発生している場合に `true` となる。以下の 3 つのレベルが存在する。
    - **Info**: 一切、問題が発生していない状態。
    - **Warning**: 1〜2 つの問題が発生している状態。
    - **Critical**: 3 つの問題が発生している状態。
- **グラフ**: 24 時間分の ISU の状態を 1 時間単位で可視化したもの。
- **トレンド**: 性格ごとに、最新の ISU のコンディションを累計したもの。 

## ISUCONDITIONの機能とユーザ、ISU、閲覧者について

### ログイン

ISUCONDITION は認証を JIA に委ねており、ユーザは JIA の認証サイトで認証成功時に発行されるトークンを使って ISUCONDITION にログインします。

ログインの処理は以下のような流れになります。

![ログインの動き](https://user-images.githubusercontent.com/210692/130006129-293ac048-30c3-4a4b-815c-67572ef3b44e.png)

1. ユーザは、ISUCONDITION のトップページにアクセスします。
2. ユーザが、ISUCONDITION のトップページにある "JIA のアカウントでログイン" のボタンをクリックすると JIA 認証サイトへ遷移します。
3. ユーザが、JIA 認証サイトで JIA のアカウント情報を入力します。
4. JIA 認証サイトは認証成功時にトークン（JWT: JSON Web Token）を発行し、ユーザを ISUCONDITION にリダイレクトします。リダイレクトの URL にはクエリパラメータにトークンが設定されています。
5. ISUCONDITION はトークンが妥当なものかを検証します。
6. トークンの妥当性が確認された場合、ユーザは ISUCONDITION のログインに成功します。

### ISUの登録

ユーザが、ISUCONDITION に ISU を登録することで、ISU から ISUCONDITION へのコンディション送信が開始されます。
コンディション送信を開始するには JIA による ISU のアクティベートが必要です。

ISU の登録は以下のような流れになります。

![ISUのアクティベートイメージ](https://user-images.githubusercontent.com/210692/130178391-005ee191-202d-45d9-95d5-d8f4405d355d.png)

1. ユーザが ISUCONDITION に JIA ISU ID を入力します。
2. ISUCONDITION は JIA の ISU 管理サービスに対して JIA ISU ID を送信します。
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
| 0694e4d7-dfce-4aec-b7ca-887ac42cfb8f |
| f012233f-c50e-4349-9473-95681becff1e |

### ISUのコンディション送信処理

コンディションの送信先 URL はアクティベート時に、 ISUCONDITION が JSON で送信する `target_base_url` と `isu_uuid` により以下のように決定されます。

```
${target_base_url}/api/condition/${isu_uuid}
```

ISU はアクティベートされると、自身のコンディションを送信先 URL へ継続的に送信するようになります。

ISU から送信されるデータには 1 つ以上のコンディションが含まれます。送信されるコンディションは ISU 単位で下記が保証されています。

- コンディションの時刻情報が重複することはない。
- コンディションの時刻情報は、既に送られたコンディションの時刻情報よりも常に先の時刻となる。

ユーザはコンディションのデータ欠損を許容しますが、理想的には全てのコンディションのデータが保存されることを期待しています。

### 登録済みの ISU の確認

ユーザは、自身が登録した ISU の一覧（`GET /api/isu`）を確認しています。ユーザは ISU の一覧を見て、各 ISU の詳細（`GET /api/isu/:jia_isu_uuid`）を確認します。

ユーザは一覧中の、最新コンディションに変化がない ISU には興味を持ちません。

### ISU の詳細確認

ISU の詳細ページでは、次のことが行えます。

- コンディションの確認（`GET /api/condition/:jia_isu_uuid`） 
- グラフの確認（`GET /api/isu/:jia_isu_uuid/graph`）

#### コンディションの確認

コンディションの確認では ISU に登録されたコンディションの履歴を確認できます。このとき、コンディションレベルや時刻情報で表示するデータを絞り込めます。

ISU のコンディションレベルは悪くなる事があり、ユーザは悪いコンディションレベル（ `Info` 以外）を確認すると速やかに問題を改善します。

#### グラフの確認

グラフは、ある日の 24 時間分の ISU の状態を可視化したものです。過去のグラフも遡って確認できます。

データポイントは 0 時から 24 時までの 1 時間単位で集計されています。

### トレンドの閲覧

閲覧者は、ISUCONDITION のトップページに表示されるトレンド（`GET /api/trend`）を確認しています。

トレンドでは ISUCONDITION に登録されているすべての ISU の最新のコンディションレベルが性格ごとにまとまっており、ISU が持つ性格ごとの傾向を見ることができます。ISUCONDITION に興味を持っている閲覧者は、トレンドの変化に注目しています。

## JIA ISU 管理サービス API 

JIA ISU 管理サービスが提供する API は以下の通りです。
同様の機能が、開発/検証用に JIA API Mock という名前で提供されています（後述）。

### `POST /api/activate`

JIA が管理する ISU をアクティベートするためのエンドポイントです。

`target_base_url` には下記の制約があります。これに違反した場合、 JIA から `400 Bad Request` が返され ISU のアクティベートに失敗します。

- ホスト部は下記の 3 つのみを指定できる。
  - `isucondition-1.t.isucon.dev`
  - `isucondition-2.t.isucon.dev`
  - `isucondition-3.t.isucon.dev`
- スキームには `https` のみが利用できる。
- ポート番号は指定できない。

また、同一の ISU に対する 2 度目以降のリクエストは成功しますが `target_base_url` は 1 度目の内容が利用されます。

+ Request（application/json）
    + Schema

            {
                "target_base_url": "string",
                "isu_uuid": "string"
            }


    + Attributes

        | Field           | Type   | Required | Description          | Example                                |
        |-----------------|--------|----------|----------------------|----------------------------------------|
        | target_base_url | string | true     | ISU のコンディション送信先 | `https://isucondition-1.t.isucon.dev`  |
        | isu_uuid        | string | true     | JIA ISU ID           | `0694e4d7-dfce-4aec-b7ca-887ac42cfb8f` |


+ Response 202（application/json）
    + Schema

            {
                "character": "string"
            }

    + Attributes

        | Field     | Type   | Required | Description                 | Example    |
        |-----------|--------|----------|-----------------------------|------------|
        | character | string | true     | アクティベートされた ISU の性格 | `いじっぱり` |


+ Other Responsess
    + 400（text/plain）
    + 404（text/plain）


### JIA API Mock について

JIA API Mock は、ISUCONDITION の開発に利用できる JIA の API モックとして、選手に提供される各サーバーのポート 5000 番で待ち受けています。
JIA API Mock は以下の機能を持っています。

- ISU 管理サービス(`POST /api/activate`)
  - ただし、先述した `target_base_url` の制約は存在しない
- 登録した ISU から ISUCONDITION へ向けたテスト用コンディションの送信

JIA API Mock は ISU がアクティベートされると、JIA API Mock が停止されるまで ISUCONDITION へテスト用コンディションの送信を行います。JIA API Mock の操作は永続化されません。
そのため、負荷走行前には JIA API Mock を停止または再起動することでテスト用コンディションの送信を停止することをお勧めします。

JIA API Mock のサービスを停止または再起動するには、 以下のコマンドを利用してください。

```shell
sudo systemctl [stop|restart] jiaapi-mock.service
```

なお、負荷走行後に JIA API Mock を利用する際は、下記のように `POST /initialize` で JIA API Mock のエンドポイントを再設定してください。

```shell
curl -fk -H 'content-type: application/json' https://localhost/initialize -d '{"jia_service_url": "http://localhost:5000"}'
```

なお、各サーバーのポート 5000 番は外部から接続できないよう、AWS のセキュリティグループで設定されています。セキュリティグループを変更すると当日マニュアルに記載されている環境確認が失敗するようになり、失格となる恐れがあるのでご注意ください。
