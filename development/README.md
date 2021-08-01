# development

開発用資材置場です。

開発する前に S3 から必要なデータを取ってくる必要があるため、 `初期データについて` の項はご一読ください。

## 初期データについて

benchmarkerを動かすのには初期データが必要です。

初期データはbenchmarker用とDB用にそれぞれあるのでS3から取得して配置してください。
なお、isucon11 AWS アカウントへのログイン方法は[こちら](https://scrapbox.io/ISUCON11/AWS%E3%82%A2%E3%82%AB%E3%82%A6%E3%83%B3%E3%83%88)を参考にしてください。

* https://s3.console.aws.amazon.com/s3/buckets/isucon11-qualify-dev?region=ap-northeast-1&tab=objects

ダウンロードしたデータはそれぞれ以下に配置する必要があります。

* initialize.json (benchmarker用の初期データ) : `bench/data/` 以下に `initialize.json` という名前で配置
* initial-data.sql (DB用の初期データ) : `webapp/sql` 以下に `1_InitData.sql` という名前で配置

### 実行方法

* bench 開発用 : `make up-bench`
* backend 開発用 (Go) : `make up-go`

### MEMO: ファイル種別

* prefix や suffix に `dev` (例: `backend-go/dev.dockerfile`, `docker-compose-dev.yml`) : ローカル開発用
    * 基本的に Makefile 越しに操作
* prefix や suffix に `ci` (例.  `ci.dockerfile`, `docker-compose-ci.yml`) : CI 用ファイル
    * 環境変数 `target` に言語名 (例: `go`) を入れて実行する必要あり

