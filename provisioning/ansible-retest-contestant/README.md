ansible-retest-contestant
===

競技者 VM の チェックスクリプトを置き換え後、VM 再起動を行うための Ansible Playbook が配置されています。
競技当日の競技終了後に利用されます。

### requirement

* ansible 4.2.0 (core 2.11.2)

### 実行方法

* `ssh_config` にて jump サーバの指定
    * 本番環境の構成上、手元から ansible を実行する場合は jump サーバ経由で ssh する必要があります

```diff
  Host *
    User isucon-admin
    ControlMaster auto
    ControlPath ~/.ssh/%C
    ControlPersist 60s
+   IdentityFile <秘密鍵のPATH>
+   ProxyJump <jumpサーバのアドレス>
```

* 再配置するチェックスクリプトは envchecker に配置
    * isucon-env-checker は必要に応じてビルドを行う
    * 配置してない場合、task は fail されます

```
cp -a ../../extra/envchecker .
pushd envchecker/isucon-env-checker
env GOOS=linux GOARCH=amd64 go build
popd
```

* Ansible の実行

```
ansible-playbook -i hosts tasks.yml
```
