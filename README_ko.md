# Tug

[![Release](https://img.shields.io/github/v/release/supergate-hub/tug?style=flat-square)](https://github.com/supergate-hub/tug/releases)
[![License](https://img.shields.io/github/license/supergate-hub/tug?style=flat-square)](LICENSE)

**Tug**는 Slurm 클러스터와의 상호작용을 단순화하는 경량 데몬입니다. `slurmrestd`를 위한 보안 프록시 역할을 수행하며, JWT 인증을 자동으로 처리하고 작업 제출 및 관리를 위한 간소화된 API를 제공합니다.

[🇺🇸 English](README.md)

---

## 주요 기능

- **JWT 인증 자동화**: 개인 키를 사용하여 `slurmrestd`용 JWT 토큰을 자동으로 생성하고 관리합니다.
- **보안 프록시**: 인증 헤더(`X-SLURM-USER-NAME`, `X-SLURM-USER-TOKEN`)를 주입하여 `slurmrestd`로 요청을 전달합니다.
- **간편한 설정**: 직관적인 YAML 설정 파일을 사용합니다.
- **Systemd 통합**: 손쉬운 배포를 위한 systemd 서비스 파일이 포함되어 있습니다.

## 설치 방법

### Linux (Debian/Ubuntu)

[Releases](https://github.com/supergate-hub/tug/releases) 페이지에서 `.deb` 패키지를 다운로드합니다.

```bash
sudo dpkg -i tug_x.y.z_linux_amd64.deb
```

### Linux (RHEL/CentOS)

[Releases](https://github.com/supergate-hub/tug/releases) 페이지에서 `.rpm` 패키지를 다운로드합니다.

```bash
sudo rpm -ivh tug_x.y.z_linux_amd64.rpm
```

### 바이너리 직접 설치

Releases 페이지에서 아키텍처에 맞는 바이너리를 다운로드합니다.

```bash
# 예시
chmod +x tug
sudo mv tug /usr/local/bin/
```

## 빠른 시작

1.  **설정 파일 생성**

    `/etc/tug/config.yaml` 파일을 생성합니다. (디렉토리가 없다면 생성하세요)

    ```yaml
    # /etc/tug/config.yaml
    listenAddr: ":8080"

    slurmrestd:
      uri: "http://localhost:6820"
      version: "v0.0.40"
      jwtMode: "auto"
      jwtUser: "slurm"
      jwtLifespan: 360
      jwtKey: "/etc/tug/jwt_hs256.key" # Slurm JWT 개인 키 경로
    ```

2.  **JWT 키 설정 (Auto 모드 필수)**

    `jwtMode: "auto"`를 사용하는 경우 Slurm JWT 개인 키가 필요합니다.
    키 파일을 안전한 위치로 복사하고, `tug` 사용자만 읽을 수 있도록 권한을 제한하세요.

    ```bash
    # 키 파일 복사 (원본 경로는 Slurm 설정에 따라 다를 수 있음)
    sudo cp /var/spool/slurm/statesave/jwt_hs256.key /etc/tug/jwt_hs256.key

    # 소유권을 tug 사용자로 변경
    sudo chown tug:tug /etc/tug/jwt_hs256.key

    # 권한 제한 (소유자만 읽기 가능)
    sudo chmod 0400 /etc/tug/jwt_hs256.key
    ```

3.  **서비스 시작**

    ```bash
    sudo systemctl enable --now tug
    sudo systemctl status tug
    ```

## 사용법

`slurmrestd`에 직접 요청하는 대신 Tug 데몬으로 요청을 보냅니다. Tug가 필요한 인증 토큰을 자동으로 주입합니다.

**작업 제출 예시:**

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

## 기여하기

기여는 언제나 환영합니다! Pull Request를 보내주세요.

## 라이선스

자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.
