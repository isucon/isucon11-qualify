packer
===

競技用 VM (contestant) 、 ベンチ VM (bench) の AMI を作成するための packer 用設定ファイルが配置されています。

### requirement

* packer 1.7.3
* aws-cli 2.2.17
* jsonnet v0.17.0

### 実行方法

* contestant VM 用 AMI の作成

```
make build-contestant
```

* bench VM 用 AMI の作成

```
make build-bench
```
