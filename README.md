# isucon11-qualify

## ディレクトリ構成

```
.
├── bench        # ベンチマーカー
├── development  # 開発用資材置場
├── extra        # その他のファイル
├── provisioning # セットアップ用
└── webapp       # 各言語の参考実装
```

## JWT で利用する公開鍵、秘密鍵

* bench/key/ec256-private.pem
* bench/key/ec256-public.pem
* webapp/ec256-public.pem (bench/key/ec256-public.pemのコピー)
* webapp/jiaapi_mock/cmd/standalone/ec256-private.pem (bench/key/ec256-private.pemのコピー)

となっています。

## Links
