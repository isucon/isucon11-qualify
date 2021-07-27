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
  "zone_b": [
  ],
  "zone_c": [
  ]
}
```

* terraform の実行

```
export AMI_ID=<bench 用 AMI ID>
export ISUXPORTAL_SUPERVISOR_ENDPOINT_URL=<portal の gRPC エンドポイント>
export ISUXPORTAL_SUPERVISOR_TOKEN=<portal の supervisor 接続用トークン>
make all
```
