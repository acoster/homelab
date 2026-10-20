# homelab

This is a collection of configs for my homelab. A few notes for my future self:

  - Secrets are encrypted with sops + age. To run docker compose commands, use
    `sops exec-env secrets.enc.env 'docker compose ...'`
  - Backups use restic. The keys are somewhere safe
