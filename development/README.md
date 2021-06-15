### MEMO: ファイル種別

* prefix や suffix に `dev` (例: `backend-go/dev.dockerfile`, `docker-compose-dev.yml`) : ローカル開発用
    * 基本的に Makefile 越しに操作
* prefix や suffix に `ci` (例.  `ci.dockerfile`, `docker-compose-ci.yml`) : CI 用ファイル
    * 環境変数 `target` に言語名 (例: `go`) を入れて実行する必要あり
