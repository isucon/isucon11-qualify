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

	var name string
	if len(os.Args) == 2 && os.Args[1] == "boot" {
		name = "test-boot"
	} else {
		name = "test-ssh"
		fmt.Println("SSH 接続が成功しました")
	}

	info, err := p.GetInfo(name)
	if err != nil {
		fmt.Printf("ポータルから情報の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("環境をチェックしています...")
	result := Check(CheckConfig{
		Name: name,
		AMI:  info.AMI,
		AZ:   info.AZ,
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
