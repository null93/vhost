# NGINX — vhost

> NGINX config structure for provisioning virtual hosts with a CLI tool

## About

This is a proposal for a new NGINX config structure that allows for provisioning virtual hosts with a CLI tool.
Please keep in mind that this is a proof of concept and the actual config files in the example nginx config directory are not complete.
It is the minimum viable product to demonstrate the CLI tool rather than a complete NGINX config.

![demo-1](assets/demo-1.svg)

## Requirements

All you need is docker installed with the docker compose plugin.

## Try It Out

On your host machine, run the following commands:

```
docker compose up -d
docker compose exec nginx bash
```

Once you are inside the docker container, you can create some virtual hosts:

```
vhost create catch-all default-backend
vhost enable default-backend

vhost create wordpress my-blog domain_names=wordpress-127-0-0-1.nip.io
vhost enable my-blog

vhost create magento-2 my-store magento_version=2.4.6.3 domain_names=magento-127-0-0-1.nip.io
vhost enable my-store
```

Finally you can reload nginx to apply the changes:

```
nginx -s reload
```

That's it! You can now visit the following URLs:

```
http://localhost
http://wordpress-127-0-0-1.nip.io
http://magento-127-0-0-1.nip.io
```

## Install

<details>
  <summary>Darwin</summary>

  ### Intel & ARM
  
  ```shell
  brew tap null93/tap
  brew install vhost
  ```
</details>

<details>
  <summary>Debian</summary>

  ### amd64
  
  ```shell
  curl -sL -o ./vhost_0.0.3_amd64.deb https://github.com/null93/vhost/releases/download/0.0.3/vhost_0.0.3_amd64.deb
  sudo dpkg -i ./vhost_0.0.3_amd64.deb
  rm ./vhost_0.0.3_amd64.deb
  ```

  ### arm64

  ```shell
  curl -sL -o ./vhost_0.0.3_arm64.deb https://github.com/null93/vhost/releases/download/0.0.3/vhost_0.0.3_arm64.deb
  sudo dpkg -i ./vhost_0.0.3_arm64.deb
  rm ./vhost_0.0.3_arm64.deb
  ```
</details>

<details>
  <summary>Red Hat</summary>
  
  ### aarch64

  ```shell
  rpm -i https://github.com/null93/vhost/releases/download/0.0.3/vhost-0.0.3-1.aarch64.rpm
  ```

  ### x86_64

  ```shell
  rpm -i https://github.com/null93/vhost/releases/download/0.0.3/vhost-0.0.3-1.x86_64.rpm
  ```
</details>

<details>
  <summary>Alpine</summary>
  
  ### aarch64

  ```shell
  curl -sL -o ./vhost_0.0.3_aarch64.apk https://github.com/null93/vhost/releases/download/0.0.3/vhost_0.0.3_aarch64.apk
  apk add --allow-untrusted ./vhost_0.0.3_aarch64.apk
  rm ./vhost_0.0.3_aarch64.apk
  ```

  ### x86_64

  ```shell
  curl -sL -o ./vhost_0.0.3_x86_64.apk https://github.com/null93/vhost/releases/download/0.0.3/vhost_0.0.3_x86_64.apk
  apk add --allow-untrusted ./vhost_0.0.3_x86_64.apk
  rm ./vhost_0.0.3_x86_64.apk
  ```
</details>

<details>
  <summary>Arch</summary>
  
  ### aarch64

  ```shell
  curl -sL -o ./vhost-0.0.3-1-aarch64.pkg.tar.zst https://github.com/null93/vhost/releases/download/0.0.3/vhost-0.0.3-1-aarch64.pkg.tar.zst
  sudo pacman -U ./vhost-0.0.3-1-aarch64.pkg.tar.zst
  rm ./vhost-0.0.3-1-aarch64.pkg.tar.zst
  ```

  ### x86_64

  ```shell
  curl -sL -o ./vhost-0.0.3-1-x86_64.pkg.tar.zst https://github.com/null93/vhost/releases/download/0.0.3/vhost-0.0.3-1-x86_64.pkg.tar.zst
  sudo pacman -U ./vhost-0.0.3-1-x86_64.pkg.tar.zst
  rm ./vhost-0.0.3-1-x86_64.pkg.tar.zst
  ```
</details>
