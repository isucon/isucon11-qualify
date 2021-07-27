terraform-bench
===

本番用 bench 環境を複数展開するための terraform ファイルが配置されています。

### requirement

* terraform v1.0.1

### 実行方法

```
export AMI_ID=<bench 用 AMI ID>
export ISUXPORTAL_SUPERVISOR_ENDPOINT_URL=<portal の gRPC エンドポイント>
export ISUXPORTAL_SUPERVISOR_TOKEN=<portal の supervisor 接続用トークン>
make all
```
