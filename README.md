# Joe Portal Backend

This is the backend service for the Joe Portal application.

## Domain Configuration

### Recommended Setup

- Frontend: `app.yourdomain.com`
- Backend: `api.yourdomain.com`

### DNS Configuration

1. Add an A record in your domain's DNS settings:

   - Name: `api` (or your preferred subdomain)
   - Value: Your DigitalOcean Droplet's IP address
   - TTL: 3600 (or your preferred TTL)

2. Update Nginx configuration to use the subdomain:

   ```nginx
   server {
       listen 80;
       server_name api.yourdomain.com;  # Update this to your subdomain

       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

3. Update SSL certificate to include the subdomain:
   ```bash
   certbot --nginx -d api.yourdomain.com
   ```

## Deployment Guide (DigitalOcean)

### Prerequisites

- A DigitalOcean account
- Domain name (optional but recommended)
- SSH access to your local machine

### 1. Create a DigitalOcean Droplet

- Create a new Droplet on DigitalOcean
- Choose Ubuntu 22.04 LTS as the distribution
- Select a plan (Basic plan with 1GB RAM recommended for starting)
- Choose a datacenter region closest to your users
- Add your SSH keys for secure access
- Create the Droplet

### 2. Server Setup

SSH into your Droplet:

```bash
ssh root@your_droplet_ip
```

Install required dependencies:

```bash
apt update
apt install -y git golang-go
```

### 3. Deploy Application

```bash
# Create application directory
mkdir -p /opt/joe-portal
cd /opt/joe-portal

# Clone repository
git clone https://github.com/dramsay1996/joe-portal-backend.git .

# Build application
go build -o bin/server ./cmd/server
```

### 4. Environment Setup

Create a `.env` file with required environment variables:

```bash
nano .env
```

### 5. Create Systemd Service

Create service file:

```bash
nano /etc/systemd/system/joe-portal.service
```

Add this configuration:

```ini
[Unit]
Description=Joe Portal Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/joe-portal
Environment="GIN_MODE=release"
ExecStart=/opt/joe-portal/bin/server
Restart=always

[Install]
WantedBy=multi-user.target
```

Start and enable the service:

```bash
systemctl daemon-reload
systemctl enable joe-portal
systemctl start joe-portal
```

### 6. Nginx Setup (Optional but Recommended)

Install Nginx:

```bash
apt install -y nginx
```

Create Nginx configuration:

```bash
nano /etc/nginx/sites-available/joe-portal
```

Add this configuration:

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;  # Update this to your subdomain

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Enable the site:

```bash
ln -s /etc/nginx/sites-available/joe-portal /etc/nginx/sites-enabled/
nginx -t
systemctl restart nginx
```

### 7. SSL Setup (Optional but Recommended)

```bash
apt install -y certbot python3-certbot-nginx
certbot --nginx -d api.yourdomain.com
```

### 8. Security Setup

Set up firewall:

```bash
apt install -y ufw
ufw allow ssh
ufw allow http
ufw allow https
ufw enable
```

Create non-root user:

```bash
adduser deployer
usermod -aG sudo deployer
```

### 9. Monitoring

Check service status:

```bash
systemctl status joe-portal
```

View logs:

```bash
journalctl -u joe-portal -f
```

### Updating the Application

To update your application:

```bash
cd /opt/joe-portal
git pull
go build -o bin/server ./cmd/server
systemctl restart joe-portal
```

### Security Notes

- Always use environment variables for sensitive data
- Regularly update your system and dependencies
- Monitor your application logs for any issues
- Consider setting up automated backups
- Use strong passwords and SSH keys for authentication
