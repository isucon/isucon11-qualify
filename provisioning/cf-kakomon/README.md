# cf-kakomon

isucon11 予選の過去問環境を AWS 上に構築するためのファイルが配置されています。

## 構築手順

1. AWS にて事前に EC2 KeyPair の作成してください
2. packer を実行し AMI を作成してください
    * 詳しくは [AMI の構築](#ami-の構築) を参照
3. cf.yaml ファイルを開き、競技者・ベンチ用 AMI の ID を更新してください
    * 詳しくは [cf.yaml に書かれた AMI ID の更新](#cf.yaml-に書かれた-ami-id-の更新) を参照
4. CloudFormation より cf.yaml を利用してスタックを作成してください
    * パラメータに手順2で作成した KeyPair 名を指定
5. CloudFormation により作成される EC2 インスタンスには、手順2で作成した秘密鍵を利用して `ubuntu` ユーザでログイン可能です

### AMI の構築

以下のコマンドを事前にインストールする必要があります。

* https://github.com/hashicorp/packer
* https://github.com/google/jsonnet

AMI の構築手順は以下のとおりです。

1. https://github.com/isucon/isucon11-qualify/releases/tag/public の各ファイル (`1_InitData.sql`, `initialize.json`, `images.tgz`) を `provisioning/packer/files-generated/` 以下に配置
    * **`1_InitData.sql` は `initial-data.sql` に rename してください**
2. aws-cli を用い AMI のビルドと登録を行う AWS アカウントにログイン
3. `provisioning/packer/` 以下に移動し、AMI の構築を行うコマンドの実行
    * ベンチ用イメージの構築: `make build-bench_kakomon`
    * 競技者用イメージの構築: `make build-contestant_kakomon`

### cf.yaml に書かれた AMI ID の更新

* 以下のコマンドを実行することで AMI ID が更新されます。
    * `${CONTESTANT_AMI_ID}` , `${BENCH_AMI_ID}` にはそれぞれ [AMI の構築](#ami-の構築) 手順で構築した AMI の ID を指定してください

```shell
sed -i \
  -e 's|__CONTESTANT_AMI_ID__|'${CONTESTANT_AMI_ID}'|g' \
  -e 's|__BENCH_AMI_ID__|'${BENCH_AMI_ID}'|g' \
  provisioning/cf-kakomon/cf.yaml
```

## 構築された環境の利用方法

大方針は [当日マニュアル](../../docs/manual.md) 及び [アプリケーションマニュアル](../../docs/isucondition.md) をご参照ください。

### ベンチ実行手順

ベンチマーカインスタンスにログインした後 `isucon` ユーザで以下のコマンドを実行することでベンチマーカを実行可能です。

```
cd ~/bench
./bench -tls -target=192.168.0.11 -all-addresses=192.168.0.11,192.168.0.12,192.168.0.13 -jia-service-url http://192.168.0.10:5000
```

### ブラウザでのアクセスにおける留意点

過去問環境において競技用インスタンスで動作している isucondition にブラウザからアクセスする際の留意点です。

#### ログイン

「JIAのアカウントでログイン」を押すと http://localhost:5000 に遷移するようになっています。
このアクセスは競技用サーバ上で動作する `jiaapi-mock.service` が受ける想定です。

以下のコマンドより localhost:5000 が競技用サーバ上の 5000 番ポートにローカルフォワードされるようにした上でブラウザ操作を行ってください。

```
ssh isucon@<競技用サーバのグローバルアドレス> -L 5000:localhost:5000
```

#### ISU の登録

ブラウザより ISU の登録を行う際にも JIA API Mock が必要です。
こちらについては [アプリケーションマニュアル](../../docs/isucondition.md) をご確認ください。
