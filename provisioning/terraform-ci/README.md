terraform-ci
===

CI で以下を行うための AWS リソースを作成する terraform ファイルが配置されています。

* initial-data.sql (1_InitData.sql) を S3 に配置 / S3 から取得
* packer で AMI を作成

### requirement

* terraform v1.0.1
