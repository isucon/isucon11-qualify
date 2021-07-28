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

* 差し替えたいコンパイル済みのベンチバイナリ・supervisorバイナリを配置
    * 配置してない場合、task は skip されます

```
```

* Ansible の実行
