package main

import (
	"fmt"
	"os"
)

func main() {
	p, err := LoadPortalCredentials()
	if err != nil {
		fmt.Println("チェッカーの設定ファイルが読み込めませんでした")
		os.Exit(1)
	}

	info, err := p.GetInfo("qualify")
	if err != nil {
		fmt.Printf("ポータルから情報の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("環境をチェックしています...")
	result := Check(CheckConfig{
		AMI: info.AMI,
		AZ:  info.AZ,
	})

	if err := p.SendResult(result); err != nil {
		fmt.Printf("チェック結果の送信に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(result.Message)
	if !result.Passed {
		os.Exit(1)
	}
}
