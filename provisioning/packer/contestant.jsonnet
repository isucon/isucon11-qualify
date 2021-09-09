local base = import './base.libsonnet';

base {
  arg_variant: 'contestant',
  provisioners_plus:: [
    {
      type: 'shell',
      inline: [
        '( cd /dev/shm/ansible && sudo ansible-playbook -u root -i ' + $.arg_variant + '.hosts -c local -t aws -t qualify  site.yml )',
      ],
    },
  ],
}
