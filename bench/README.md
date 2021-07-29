

## ディレクトリ構成

```
.
├── main.go      # エントリーポイント 引数処理とか
├── key          # JWT用の鍵
├── logger       # 
├── model        # 内部データのデータ構造の定義
├── scenario     # シナリオの実行
├── service      # ネットワークとのインターフェース回り
├── gen          # 静的ファイルのhash生成用
```

## 静的ファイルチェック用のデータ更新

get/assets.goでjsなどのhash値を事前計算したscenario/assets.goを作成する。webapp/public以下のファイルから生成するので、webapp/publicに最新のファイルを配置する必要があります。  