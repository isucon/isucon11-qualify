### MEMO: ファイル種別

* suffix に `-dev` (例: `backend-go/Dockerfile-dev`) または `-dev-${lang}` (例: `docker-compose-dev-go.yml` : ローカル開発用
    * 基本的に Makefile 越しに操作
* suffix に `-ci` (例.  `Dockerfile-ci`, `docker-compose-ci.yml`) : CI 用ファイル
    * 環境変数 `target` に言語名 (例: `go`) を入れて実行する必要あり
