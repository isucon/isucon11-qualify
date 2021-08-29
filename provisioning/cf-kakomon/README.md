# cf-kakomon

isucon11 予選の過去問環境を AWS 上に構築するための CloudFormation ファイルが配置されています。

## 利用手順

1. 事前に EC2 KeyPair の作成してください
2. CloudFormation より cf.yaml を利用してスタックを作成してください
    * パラメータに手順1で作成した KeyPair 名を指定
3. CloudFormation により作成される EC2 インスタンスには、手順1で作成した秘密鍵を利用して `ubuntu` ユーザでログイン可能です

## ベンチ実行手順

ベンチマーカインスタンスにログインした後 `isucon` ユーザで以下のコマンドを実行することでベンチマーカを実行可能です。

```
cd ~/bench
./bench -tls -target=192.168.0.11 -all-addresses=192.168.0.11,192.168.0.12,192.168.0.13 -jia-service-url http://192.168.0.10:5000
```
