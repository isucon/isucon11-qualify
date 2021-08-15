ansible-replace-bench
===

ベンチ VM (bench) の bench 及び supervisor をグレースフルに差し替えるための Ansible Playbook が配置されています。
競技当日の緊急対応時に利用されます。

### requirement

* ansible 4.2.0 (core 2.11.2)

### 実行方法

* teams.json から hosts ファイルを生成

```
python generate_hosts.py > hosts
```

* `ssh_config` にて jump サーバの指定
    * 本番環境の構成上、手元から ansible を実行する場合は jump サーバ経由で ssh する必要があります

```diff
  Host *
    User isucon-admin
+   IdentityFile <秘密鍵のPATH>
  Host 192.168.*
    User isucon-admin
    ControlMaster auto
    ControlPath ~/.ssh/%C
    ControlPersist 60s
+   ProxyJump <jumpサーバのアドレス>
```

* 差し替えたいコンパイル済みのベンチバイナリ (`bench`) 、supervisorバイナリ (`isuxportal-supervisor`) をこのディレクトリに配置
    * 配置してない場合、task は skip されます

* Ansible の実行

```
ansible-playbook -i hosts tasks.yml
```
