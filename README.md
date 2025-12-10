# Tug

[![Release](https://img.shields.io/github/v/release/supergate-hub/tug?style=flat-square)](https://github.com/supergate-hub/tug/releases)
[![License](https://img.shields.io/github/license/supergate-hub/tug?style=flat-square)](LICENSE)

**Tug** is a lightweight daemon that simplifies interaction with Slurm clusters. It acts as a secure proxy for `slurmrestd`, handling JWT authentication automatically and providing a simplified API for job submission and management.

[🇰🇷 한국어 (Korean)](README_ko.md)

---

## Features

- **Automated JWT Authentication**: Automatically generates and manages JWT tokens for `slurmrestd` using a private key.
- **Secure Proxy**: Proxies requests to `slurmrestd` with proper authentication headers (`X-SLURM-USER-NAME`, `X-SLURM-USER-TOKEN`).
- **Simple Configuration**: Easy-to-read YAML configuration.
- **Systemd Integration**: Ready-to-use systemd service file for easy deployment.

## Limitations

- **Supported Slurm REST API Version**: Currently, only `v0.0.40` is supported.
  - This limitation is due to the underlying [`slurm-client`](https://github.com/supergate-hub/slurm-client) SDK, which currently implements only `v0.0.40`.
  - Support for additional versions (e.g., `v0.0.41`, `v0.0.42`) is planned for future releases.

## Installation

### Linux (Debian/Ubuntu)

Download the `.deb` package from the [Releases](https://github.com/supergate-hub/tug/releases) page.

```bash
sudo dpkg -i tug_x.y.z_linux_amd64.deb
```

### Linux (RHEL/CentOS)

Download the `.rpm` package from the [Releases](https://github.com/supergate-hub/tug/releases) page.

```bash
sudo rpm -ivh tug_x.y.z_linux_amd64.rpm
```

### Binary

Download the binary for your architecture from the Releases page.

```bash
# Example
chmod +x tug
sudo mv tug /usr/local/bin/
```

## Quick Start

1.  **Create Configuration File**

    Create `/etc/tug/config.yaml`. (Ensure the directory exists)

    ```yaml
    # /etc/tug/config.yaml
    listenAddr: ":8080"

    slurmrestd:
      uri: "http://localhost:6820"
      version: "v0.0.40"
      jwtMode: "auto"
      jwtUser: "slurm"
      jwtLifespan: 360
      jwtKey: "/etc/tug/jwt_hs256.key" # Path to your Slurm JWT private key
    ```

2.  **Setup JWT Key (Auto Mode Only)**

    If you are using `jwtMode: "auto"`, you must provide the Slurm JWT private key.
    Copy the key to a secure location and restrict permissions so only the `tug` user can read it.

    ```bash
    # Copy the key (source path may vary depending on your Slurm config)
    sudo cp /var/spool/slurm/statesave/jwt_hs256.key /etc/tug/jwt_hs256.key

    # Set ownership to tug user
    sudo chown tug:tug /etc/tug/jwt_hs256.key

    # Restrict permissions (read-only for owner)
    sudo chmod 0400 /etc/tug/jwt_hs256.key
    ```

3.  **Start Service**

    ```bash
    sudo systemctl enable --now tug
    sudo systemctl status tug
    ```

## Usage

Send requests to the Tug daemon instead of directly to `slurmrestd`. Tug will inject the required authentication tokens.

**Submit a Job:**

```bash
curl -X POST http://localhost:8080/job/submit \
  -H "X-SLURM-USER-NAME: myuser" \
  -H "Content-Type: application/json" \
  -d '{
    "script": "#!/bin/bash\n#SBATCH -J test\nsrun hostname",
    "job": {
      "name": "test-job",
      "current_working_directory": "/home/myuser"
    }
  }'
```

## Contributing

Contributions are welcome! Please submit a Pull Request.

## License

See [LICENSE](LICENSE) for details.
