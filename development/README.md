### MEMO: ファイル種別

* prefix や suffix に `dev` (例: `backend-go/dev.dockerfile`, `docker-compose-dev.yml`) : ローカル開発用
    * 基本的に Makefile 越しに操作
* prefix や suffix に `ci` (例.  `ci.dockerfile`, `docker-compose-ci.yml`) : CI 用ファイル
    * 環境変数 `target` に言語名 (例: `go`) を入れて実行する必要あり

### 初期データについて
benchmarkerを動かすのに初期データが必要です。  
初期データはbenchmarker用とDB用にそれぞれあるのでS3から取得して配置してください。  
S3へのログイン方法はこちらを参考にしてください。
* https://s3.console.aws.amazon.com/s3/object/isucon11-qualify-dev?region=ap-northeast-1&prefix=initialize.json
* ログイン方法: https://scrapbox.io/ISUCON11/AWS%E3%82%A2%E3%82%AB%E3%82%A6%E3%83%B3%E3%83%88
  
データの配置場所はこちらを参考にしてください。 
* benchmarker用の初期データ（initialize.json）は/bench/data以下に配置
* DB用の初期データ（initial-data.sql）/webapp/sql以下などに配置しdocker-composer-xxxx.ymlのbackendのvolumesに指定してください

例えば/webapp/sqlにinitial-data.sqlを配置した場合は、1_InitData.sqlに対しては以下のような記載に変更
```
- "../webapp/sql/initial-data.sql:/webapp/sql/1_InitData.sql"
```
