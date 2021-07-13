local base = import './base.libsonnet';

base {
  arg_variant: 'contestant',
  provisioners_plus:: [
    {
      type: 'shell',
      inline: [
        'sudo sh -c "echo GRUB_CMDLINE_LINUX=\'maxcpus=2 mem=2G\' > /etc/default/grub.d/99-isucon.cfg"',
        'sudo update-grub',
      ],
    },
  ],
}
