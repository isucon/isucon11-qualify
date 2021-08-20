package main

import (
	"fmt"
	"os"

	"github.com/cenkalti/backoff/v4"
)

func main() {
	p, err := LoadPortalCredentials()
	if err != nil {
		fmt.Println("チェッカーの設定ファイルが読み込めませんでした")
		os.Exit(1)
	}

	var info EnvCheckInfo
	err = backoff.Retry(func() error {
		info, err = p.GetInfo("qualify")
		return err
	}, newBackoff())
	if err != nil {
		fmt.Printf("ポータルから情報の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("環境をチェックしています...")
	result := Check(CheckConfig{
		AMI: info.AMI,
		AZ:  info.AZ,
	})

	err = backoff.Retry(func() error {
		return p.SendResult(result)
	}, newBackoff())
	if err != nil {
		fmt.Printf("チェック結果の送信に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(result.Message)
	if !result.Passed {
		os.Exit(1)
	}
}
