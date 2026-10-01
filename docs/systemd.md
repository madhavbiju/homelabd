# Systemd Deployment

If you prefer to run `homelabd` as a native service instead of using Docker, you can manage it with systemd.

## Installation

1. Move the binary to `/usr/local/bin`:
   ```bash
   sudo mv homelabd /usr/local/bin/
   sudo chmod +x /usr/local/bin/homelabd
   ```

2. Create a data directory:
   ```bash
   sudo mkdir -p /var/lib/homelabd
   ```

3. Create the configuration file `/etc/homelabd.yaml`:
   ```yaml
   server:
     host: 127.0.0.1
     port: 8080
   
   database:
     path: /var/lib/homelabd/homelabd.db
   ```

4. Create the systemd service file at `/etc/systemd/system/homelabd.service`:

```ini
[Unit]
Description=homelabd API
After=network.target docker.service

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/homelabd
Environment="HOMELABD_CONFIG=/etc/homelabd.yaml"
Restart=on-failure
RestartSec=5s

# Hardening
ProtectSystem=full
PrivateTmp=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

5. Enable and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now homelabd
   ```

## Sudoers Configuration

If you do not run `homelabd` as root, but you want to enable power operations (Reboot/Shutdown), you must configure `sudoers`.

Add the following file to `/etc/sudoers.d/homelabd`:

```text
homelabd ALL=(ALL) NOPASSWD: /sbin/shutdown
```
