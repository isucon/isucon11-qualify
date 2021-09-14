# development

**このディレクトリは予選問題開発用です。**

開発する前に 必要なデータを取ってくる必要があるため、 `初期データについて` の項はご一読ください。

なお、8/30現在作問者向けの案内のままになっています。

## 初期データについて

benchmarkerを動かすのには初期データが必要です。

初期データはbenchmarker用とDB用にそれぞれあるので[releases](https://github.com/isucon/isucon11-qualify/releases/tag/public)から取得して配置してください。
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

