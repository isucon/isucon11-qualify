terraform-bench
===

本番用 bench 環境を複数展開するための terraform ファイルが配置されています。

### requirement

* terraform v1.0.1

### 実行方法

* team ごとにどの AZ に属するかを `teams.json` に記述
    * Region は `ap-northeast-1`

```json
{
  "zone_a": [
    "1",
    "2"
  ],
  "zone_c": [
  ],
  "zone_d": [
  ]
}
```

* terraform の実行

```
export GIT_TAG=<git tag>
export ISUXPORTAL_SUPERVISOR_ENDPOINT_URL=<portal の gRPC エンドポイント>
export ISUXPORTAL_SUPERVISOR_TOKEN=<portal の supervisor 接続用トークン>
make all
```

### 備考

* 現状の terraform ファイルだと各リージョン 254 台までしか立てられない
