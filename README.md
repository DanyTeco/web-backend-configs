# Web Backend Configs

A complete VPS server configuration for hosting Go applications with automated deployment through GitHub webhooks. This setup replaces traditional shared hosting PHP/HTML environments with a modern Docker-based infrastructure.

## Automatic Deployment via Webhook

When you push code to your repository on GitHub (or GitLab), if a webhook is configured, it will automatically trigger a deployment process. This process typically involves pulling the latest code and redeploying the application container using Docker. This enables continuous deployment, ensuring that your application is always up-to-date with the latest changes from your repository.

---

### How to Configure a Webhook in a GitHub Project

1. **Navigate to Your Repository**:  
    Go to your repository on GitHub.

2. **Go to Settings > Webhooks**:  
    Click on the **Settings** tab in your repository, then select **Webhooks** from the left sidebar.

3. **Click Add webhook**:  
    Click the **Add webhook** button to create a new webhook.

4. **Enter the Webhook URL**:  
    In the **Payload URL** field, enter the endpoint that should receive the webhook POST requests (e.g., `http://your-server:6666/webhook`).

5. **Set Content Type**:  
    Select **application/json** as the content type.

6. **Set Secret Token**:  
    In the **Secret** field, enter a secret token that your server will use to verify the webhook request (this should match the `webhookSecret` in your `github_webhook.go` file).

7. **Select Trigger Events**:  
    Choose **Just the push event** for code pushes, or select **Let me select individual events** for more specific triggers.

8. **Save the Webhook**:  
    Click **Add webhook** to save your configuration.

Now, whenever a push occurs to the repository, GitHub will send a POST request to your specified URL, which will trigger your deployment script or process.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Components](#components)
- [Setup Guide](#setup-guide)
- [Configuration](#configuration)
  - [How to Configure a Webhook in a GitHub Project](#how-to-configure-a-webhook-in-a-github-project)
- [Usage Examples](#usage-examples)
- [How It Works](#how-it-works)
- [Troubleshooting](#troubleshooting)

## Overview

This project provides a complete infrastructure setup for:
- **Automatic SSL certificate management** with Let's Encrypt
- **Reverse proxy** with automatic service discovery
- **GitHub webhook integration** for automated deployments
- **Docker-based deployment** for Go applications
- **Zero-downtime deployments** with health checks

Perfect for developers migrating from shared hosting to VPS environments who want modern DevOps practices without complex setup.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   GitHub        │    │   VPS Server     │    │   Your Apps     │
│   Repository    │───▶│   nginx-proxy    │───▶│   (Go/Docker)   │
│   + Webhooks    │    │   + SSL          │    │   Auto-deployed │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Components

### 1. HTTP Proxy (`http-proxy/`)
- **nginx-proxy**: Automatic reverse proxy with service discovery
- **acme-companion**: Automatic SSL certificate management
- **Port exposure**: 80 (HTTP) and 443 (HTTPS)

### 2. Webhook Server (`webhook/`)
- **GitHub webhook listener**: Receives push notifications
- **Deploy script**: Automated deployment pipeline
- **Security**: HMAC signature verification
- **Logging**: Comprehensive deployment logs

## Setup Guide

### Prerequisites
- Ubuntu/Debian VPS with Docker and Docker Compose installed
- Domain name pointed to your VPS IP
- GitHub repository with your Go application

### Step 1: Initial Server Setup

```bash
# Install Docker and Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Clone this repository
git clone https://github.com/DanyTeco/web-backend-configs.git
cd web-backend-configs
```

### Step 2: Configure the HTTP Proxy

```bash
cd http-proxy
# Create the external network
docker network create nginx-proxy
# Start the proxy services
docker-compose up -d
```

### Step 3: Setup Webhook Server

```bash
cd ../webhook

# Edit the webhook configuration
nano github_webhook.go
# Update webhookSecret with your GitHub webhook secret

# Edit the deploy script
nano deploy.sh
# Update GITHUB_USER, GITHUB_TOKEN, and PROJECTS_DIR

# Build and run the webhook server
go mod init webhook-server
go mod tidy
go build -o webhook-server github_webhook.go
./webhook-server
```

### Step 4: Configure Your Go Application

Create a `docker-compose.yml` in your Go project repository:

```yaml
version: "3"

services:
  app:
    build: .
    container_name: my-go-app
    working_dir: /app
    environment:
      - VIRTUAL_HOST=api.yourdomain.com
      - VIRTUAL_PORT=8080
      - LETSENCRYPT_HOST=api.yourdomain.com
      - LETSENCRYPT_EMAIL=your-email@domain.com
      - GO_ENV=production
    volumes:
      - .:/app
    expose:
      - "8080"
    command: ./your-app-binary
    networks:
      - nginx-proxy

networks:
  nginx-proxy:
    external: true
```

## Configuration

### Environment Variables for Your Apps

| Variable | Description | Example |
|----------|-------------|---------|
| `VIRTUAL_HOST` | Domain for your app | `api.yourdomain.com` |
| `VIRTUAL_PORT` | Port your app runs on | `8080` |
| `LETSENCRYPT_HOST` | Domain for SSL certificate | `api.yourdomain.com` |
| `LETSENCRYPT_EMAIL` | Email for Let's Encrypt | `admin@yourdomain.com` |

### Webhook Configuration

1. **GitHub Webhook Setup**:
   - Go to your repository → Settings → Webhooks
   - Add webhook URL: `http://your-server:6666/webhook`
   - Content type: `application/json`
   - Secret: Match the `webhookSecret` in `github_webhook.go`
   - Events: Select "Push events"

2. **Deploy Script Variables**:
   ```bash
   GITHUB_USER="your_github_username"
   GITHUB_TOKEN="your_personal_access_token"
   PROJECTS_DIR="/home/user/projects"  # Where projects are cloned
   ```

## Usage Examples

### Example 1: Simple Go API

```yaml
version: "3"

services:
  app:
    image: golang:1.24.3
    container_name: example-app
    working_dir: /app
    environment:
      - VIRTUAL_HOST=backend.example-app.com
      - VIRTUAL_PORT=8081
      - LETSENCRYPT_HOST=backend.example-app.com
      - LETSENCRYPT_EMAIL=email@youremailserver.com
      - GO_ENV=development
    volumes:
      - .:/app
    expose:
      - "8081"
    command: sh -c "go mod tidy && go run main.go"
    networks:
      - nginx-proxy

networks:
  nginx-proxy:
    external: true
```

## How It Works - Step by Step

### 1. **Initial Setup**
```bash
# The nginx-proxy container automatically detects new containers
# and creates reverse proxy rules based on VIRTUAL_HOST environment variable
```

### 2. **When You Push Code to GitHub**
```
GitHub Push → Webhook Triggered → Webhook Server Receives Event
```

### 3. **Automatic Deployment Process**
```bash
1. Webhook server verifies GitHub signature
2. Extracts repository name and clone URL
3. Calls deploy.sh script with project details
4. Deploy script:
   - Clones/updates project code
   - Stops existing containers: `docker-compose down`
   - Pulls latest code: `git pull`
   - Rebuilds images: `docker-compose build`
   - Starts containers: `docker-compose up -d`
```

### 4. **SSL Certificate Generation**
```bash
# When a new container starts with LETSENCRYPT_HOST:
1. acme-companion detects the new container
2. Requests SSL certificate from Let's Encrypt
3. Configures nginx with the new certificate
4. Your app is now available via HTTPS
```

### 5. **Traffic Flow**
```
User Request → nginx-proxy (Port 80/443) → Your App Container (Port 8080)
```

## Advanced Configuration

### Custom Nginx Configuration

Create `http-proxy/custom.conf`:
```nginx
client_max_body_size 100m;
proxy_connect_timeout 300s;
proxy_send_timeout 300s;
proxy_read_timeout 300s;
```

### Useful Commands

```bash
# View all containers and their proxy status
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Check nginx-proxy configuration
docker exec nginx-proxy cat /etc/nginx/conf.d/default.conf

# Monitor deployment logs in real-time
tail -f /var/log/webhook-deploy.log

# Restart proxy services
cd http-proxy && docker-compose restart

# Test webhook endpoint
curl -X POST http://your-server:6666/health
```