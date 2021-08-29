{
  arg_arch:: 'amd64',
  arg_distro:: error '$.arg_distro not specified',
  arg_variant:: error '$.arg_variant not specified',
  arg_source_variant:: error '$.arg_source_variant not specified',

  variables: {
    revision: 'unknown',
    name: 'isucon11q-' + $.arg_arch + '-' + $.arg_variant + '-{{isotime "20060102-1504"}}-{{user "revision"}}',
    git_tag: "{{env `GIT_TAG`}}",
  },

  builder_ec2:: {
    type: 'amazon-ebs',

    region: 'ap-northeast-1',

    source_ami_filter: {
      filters: {
        'virtualization-type': 'hvm',
        'root-device-type': 'ebs',
        name: 'ubuntu/images/hvm-ssd/ubuntu-focal-20.04-' + $.arg_arch + '-server-*',
      },
      owners: ['099720109477'],
      most_recent: true,
    },

    tags: {
      Name: '{{user "name"}}',
      Packer: '1',
      Family: 'isucon11q-' + $.arg_arch + '-' + $.arg_variant,
      Project: 'qualify-dev',
      GitTag: '{{user "git_tag"}}'
    },

    // TODO: spot instance を利用する
    instance_type: 'c5.xlarge',
    // spot_price: 'auto',
    // spot_instance_types: [
    //   'c5.xlarge',
    //   'c5a.xlarge',
    //   'm5.xlarge',
    //   'm5a.xlarge',
    //   'r5.xlarge',
    //   'r5a.xlarge',
    // ],
    // spot_tags: self.tags,

    ssh_username: 'ubuntu',
    ssh_timeout: '5m',
    ssh_interface: 'public_ip',
    associate_public_ip_address: true,

    run_tags: {
      Name: 'packer-isucon11q-' + $.arg_arch + '-' + $.arg_variant,
      Project: 'qualify-dev',
      Ignore: '1',
      Packer: '1',
    },
    run_volume_tags: self.run_tags,

    ami_name: '{{user "name"}}',
    ami_regions: ['ap-northeast-1'],
    snapshot_tags: self.tags,

    launch_block_device_mappings: [
      {
        device_name: '/dev/sda1',
        volume_type: 'gp2',
        volume_size: 16,
        delete_on_termination: true,
      },
    ],
  },

  builders: [
    $.builder_ec2,
  ],

  common_provisioners:: {
    copy_files: {
      type: 'file',
      source: './files',
      destination: '/dev/shm/files',
    },
    copy_files_playbook: {
      type: 'file',
      source: '../ansible',
      destination: '/dev/shm/ansible',
    },
    copy_files_generated: {
      type: 'file',
      source: './files-generated',
      destination: '/dev/shm/files-generated',
      generated: true,
    },
    wait_cloud_init: {
      type: 'shell',
      inline: ['cloud-init status --wait'],
    },

    apt_source_ec2: {
      type: 'shell',
      inline: [
        'sudo install -o root -g root -m 0644 /dev/shm/files/sources-ec2.list /etc/apt/sources.list',
        'sudo apt-get update',
      ],
    },
    apt_source_generic: {
      type: 'shell',
      inline: [
        'sudo install -o root -g root -m 0644 /dev/shm/files/sources-generic.list /etc/apt/sources.list',
        'sudo apt-get update',
      ],
    },
    apt_upgrade: {
      type: 'shell',
      inline: [
        "sudo DEBIAN_FRONTEND=noninteractive apt-get -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confnew' upgrade",
      ],
    },

    install_ansible: {
      type: 'shell',
      inline: [
        'sudo apt-add-repository -y --update ppa:ansible/ansible',
        'sudo apt install -y ansible',
      ],
    },
    configurate_ansible: {
      type: 'shell',
      inline: [
        'sudo cp /dev/shm/files-generated/REVISION /etc/',
        'sudo mv /dev/shm/files-generated/isucon11-qualify.tar /dev/shm/ansible/roles/common/files/',
        'sudo mv /dev/shm/files-generated/isucon11-portal.tar /dev/shm/ansible/roles/bench.supervisor/files/',
        'sudo mv /dev/shm/files-generated/initial-data.sql /dev/shm/ansible/roles/contestant/files/',
        'sudo mv /dev/shm/files-generated/initialize.json /dev/shm/ansible/roles/bench/files/',
        'sudo mv /dev/shm/files-generated/images.tgz /dev/shm/ansible/roles/bench/files/',
        'sudo cp /dev/shm/files/tls-cert.pem /dev/shm/ansible/roles/contestant/files/etc/nginx/certificates',
        'sudo cp /dev/shm/files/tls-key.pem /dev/shm/ansible/roles/contestant/files/etc/nginx/certificates',
        'sudo echo "[target]\n127.0.0.1" >> /dev/shm/ansible/hosts',
        'sudo ssh-keygen -f /root/.ssh/id_rsa -t rsa -N ""',
        'sudo su -c "cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys"',
        'sudo chmod 0600 /root/.ssh/authorized_keys',
      ],
    },
    run_ansible: {
      type: 'shell',
      inline: [
        '( cd /dev/shm/ansible && sudo ansible-playbook -u root -i ' + $.arg_variant + '.hosts -t aws -t ' + $.arg_variant + ' site.yml )',
      ],
    },
    remove_ansible: {
      type: 'shell',
      inline: [
        'sudo apt remove -y ansible',
        'sudo apt-add-repository -y --remove ppa:ansible/ansible',
        'sudo rm -rf /etc/ansible',
        'sudo su -c "rm -rf /root/.ssh/*"',
      ],
    },

    remove_netplan: {
      type: 'shell',
      inline: [
        'if [ -e /etc/netplan ]; then sudo install -o root -g root -m 0644 /dev/shm/files/99_disable_netplan.cfg /etc/cloud/cloud.cfg.d/99_disable_netplan.cfg; fi',
        'if [ -e /etc/netplan ]; then sudo cp -pv /run/systemd/network/* /etc/systemd/network/; fi',
        'sudo rm -rf /etc/netplan || :',
      ],
    },

    update_grub: {
      type: 'shell',
      inline: [
        'sudo install -o root -g root -m 0644 /dev/shm/files/grub.default /etc/default/grub',
        'sudo update-grub',
      ],
    },

    sysprep: {
      type: 'shell',
      inline: [
        'sudo dpkg -l',
        'sudo systemctl list-unit-files',
        'sudo journalctl --rotate',
        'sudo journalctl --vacuum-time=1s',
        'sudo mkdir -p /var/log/journal',
        "sudo sh -c 'echo > /etc/machine-id'",
        "sudo sh -c 'echo > /home/ubuntu/.ssh/authorized_keys'",
        'sudo mv /etc/sudoers.d/*-cloud-init-users /root/ || :',
        'sudo rm -f /var/lib/systemd/timesync/clock || :',
        'sudo rm -rf /var/lib/dbus/machine-id',
        'sudo rm -rf /root/go',
        'sudo rm -rf /root/.ansible',
        'sudo rm -rf /home/isucon/.ansible',
        'sudo rm -rf /dev/shm/files',
        'sudo rm -rf /dev/shm/files-generated',
        'sudo rm -rf /dev/shm/ansible',
        // Rmdir deployed by ansible
        'sudo rm -rf /tmp/isucon11-qualify',
        // Cleanup cloud-init except scripts
        'sudo mv /var/lib/cloud/scripts /tmp/cloud-init.scripts',
        'sudo rm -rf /var/lib/cloud/*',
        'sudo mv /tmp/cloud-init.scripts /var/lib/cloud/scripts',
      ],
    },
  },

  provisioners_plus:: [],
  provisioners: [
    $.common_provisioners.copy_files,
    $.common_provisioners.copy_files_playbook,
    $.common_provisioners.copy_files_generated,

    $.common_provisioners.wait_cloud_init,
    #$.common_provisioners.apt_source_generic,
    $.common_provisioners.apt_source_ec2,
    $.common_provisioners.apt_upgrade,
    $.common_provisioners.install_ansible,
    $.common_provisioners.configurate_ansible,
    $.common_provisioners.run_ansible,
    // $.common_provisioners.remove_netplan,
  ] + $.provisioners_plus + [
    $.common_provisioners.remove_ansible,
    $.common_provisioners.sysprep,
  ],

  'post-processors': [
    {
      type: 'manifest',
      output: 'output/manifest-' + $.arg_arch + '-' + $.arg_variant + '.json',
      strip_path: true,
      custom_data: {
        family: 'isucon11q-' + $.arg_arch + '-' + $.arg_variant,
        name: '{{user "name"}}',
      },
    },
  ],
}
