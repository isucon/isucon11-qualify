#!/usr/bin/env bash

# for monitoring
mkdir ~ubuntu/.ssh
cat << _EOF_ >> ~ubuntu/.ssh/authorized_keys
$(
accounts="
 https://github.com/mirakui
 https://github.com/rosylilly
 https://github.com/sorah
 https://github.com/eagletmt
 https://github.com/sugyan
 https://github.com/Mahito
 https://github.com/nana4gonta
 https://github.com/kfly8
 https://github.com/okashoi
 https://github.com/hoto17296
 https://github.com/iwashi
 https://github.com/kinmemodoki
 https://github.com/ockie1729
 https://github.com/Nagarei
 https://github.com/buchy
 https://github.com/motoki317
 https://github.com/ShotaKitazawa
 https://github.com/hkws
 https://github.com/takonomura
 https://github.com/TakahashiKazuya
 https://github.com/yfujit
 https://github.com/ryoha000
 https://github.com/sapphi-red
 https://github.com/FujishigeTemma
 https://github.com/Hosshii
 https://github.com/oribe1115
 https://github.com/Osumi1125
 "
 for account in $accounts; do curl -s $account.keys; done
)
_EOF_
chmod 0700 ~ubuntu/.ssh
chmod 0600 ~ubuntu/.ssh/authorized_keys
chown ubuntu: ~ubuntu/.ssh -R
