  #!/bin/bash
  set -e

  if [ -f .env ]; then
      export $(grep -v '^#' .env | xargs)
  fi

  VM_NAME="foreman-base"
  FILE="${VM_NAME}.tar.zst"
  SERVER="root@${VPS}:/var/www/images/"

  ssh root@$VPS "mkdir -p /var/www/images"

  orb delete "$VM_NAME" 2>/dev/null || true
  orb create -a amd64 -c confs/orbstack-foreman-base.yml rocky:9 $VM_NAME

  # Wait for cloud-init to finish
  orb -m "$VM_NAME" -u root cloud-init status --wait

  orb export "$VM_NAME" "$FILE"
  rsync --progress "$FILE" "$SERVER"
  rm "$FILE"
