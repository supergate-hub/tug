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

2.  **Setup JWT Key**
    
    Copy your Slurm JWT private key to `/etc/tug/jwt_hs256.key` and set permissions.

    ```bash
    sudo cp /path/to/slurm/jwt.key /etc/tug/jwt_hs256.key
    sudo chown tug:tug /etc/tug/jwt_hs256.key
    sudo chmod 600 /etc/tug/jwt_hs256.key
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

MIT License. See [LICENSE](LICENSE) for details.
