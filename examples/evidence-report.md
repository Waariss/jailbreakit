# jailbreakit Lab Readiness Evidence

- Tool: jailbreakit
- Version: v1.3.0
- Timestamp: 2026-06-25T12:00:00Z
- Host: darwin/arm64
- Safety note: Generated for authorized iOS security testing only.

## Available Host Dependencies
- ideviceinfo: available at /opt/homebrew/bin/ideviceinfo
- iproxy: available at /opt/homebrew/bin/iproxy
- ssh: available at /usr/bin/ssh
- scp: available at /usr/bin/scp
- ideviceinstaller: available at /opt/homebrew/bin/ideviceinstaller
- curl: available at /usr/bin/curl
- palera1n: optional, missing

## Connected Device
- product_type: iPhone10,6
- model: iPhone X
- chip: A11
- ios: 16.7.10

## Recommended Jailbreak / Testing Route
- palera1n 2.x: rootless or rootful fakefs, semi-tethered. A8-A11/checkm8 device with iOS 15+; good when USB/DFU flow is acceptable.

## IPA Install Readiness
- ssh: available at /usr/bin/ssh
- scp: available at /usr/bin/scp
- iproxy: available at /opt/homebrew/bin/iproxy
- ideviceinstaller: available at /opt/homebrew/bin/ideviceinstaller
- appinst: optional, missing
- ipainstaller: optional, missing

## Frida / Objection Readiness
- frida: available at /usr/local/bin/frida (17.0.0)
- frida-ps: available at /usr/local/bin/frida-ps (17.0.0)
- objection: available at /usr/local/bin/objection (1.11.0)
- Next: Install host tools if missing: python3 -m pip install frida-tools objection
- Next: List USB-connected processes: frida-ps -U
- Next: With a jailbroken device and SSH, ensure frida-server is installed and running on the device.

## SSH Readiness
- Skipped: pass --ssh-host 127.0.0.1 --ssh-port 2222 --ssh-user root after starting iproxy
