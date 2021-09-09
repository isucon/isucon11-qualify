# isucon11-qualify

## ディレクトリ構成

```
.
├── webapp       # 各言語の参考実装
├── docs         # 競技用マニュアル
├── bench        # ベンチマーカー
├── provisioning # セットアップ用
├── development  # 開発用資材置場
└── extra        # その他のファイル
```

## JWT で利用する公開鍵・秘密鍵

ISUCON11 予選ではウェブアプリケーションのログインに JWT を利用しています。
JWT を生成・検証するための公開鍵・秘密鍵はそれぞれ以下に配置されています。

* bench/key/ec256-private.pem
* bench/key/ec256-public.pem
* webapp/ec256-public.pem (bench/key/ec256-public.pemのコピー)
* extra/jiaapi-mock/ec256-private.pem (bench/key/ec256-private.pemのコピー)

## ISUCON11 予選のインスタンスタイプ

* 競技者 VM * 3
    * InstanceType: c5.large
    * VolumeType: gp3
* ベンチ VM * 1
    * InstanceType: c4.xlarge
    * VolumeType: gp3

## AWS 上での過去問環境の構築方法

### 用意された AMI を利用する場合

[provisioning/cf-kakomon](./provisioning/cf-kakomon) を参照してください。なお AMI は ISUCON11 運営の解散を目処に公開を停止する予定です。上記イメージが参照不可である場合ひとつ下の手順で構築してください。

### AMI を自前で作成し構築する場合

手順準備中です。

## Links

- [ISUCON11 予選レギュレーション](https://isucon.net/archives/55854734.html)
- [ISUCON11 予選当日マニュアル](./docs/manual.md)


