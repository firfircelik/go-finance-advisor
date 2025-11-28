# Go Finance Advisor - Production Deployment Guide

## Overview
This guide provides instructions for deploying the Go Finance Advisor application to production environments.

## Prerequisites
- Docker and Docker Compose installed
- Domain name (optional, for HTTPS)
- SSL certificate (optional, for HTTPS)

## Quick Start

### 1. Clone and Configure
```bash
git clone <your-repo-url>
cd go-finance-advisor
```

### 2. Environment Configuration
Create a `.env` file in the root directory:
```env
JWT_SECRET=your-secure-jwt-secret-key-minimum-32-characters
JWT_REFRESH_SECRET=your-secure-refresh-secret-key-minimum-32-characters
PORT=8081
DATA_PATH=/data/finance.db
```

**Security Note**: Generate strong random secrets for JWT keys:
```bash
openssl rand -base64 32
```

### 3. Deploy with Docker Compose
```bash
docker-compose up -d
```

The application will be available at `http://localhost:8081`

## Production Deployment Options

### Option 1: Docker Compose (Recommended)
Use the provided `docker-compose.yml` for easy deployment with persistent data storage.

### Option 2: Kubernetes
For scalable deployments, use the provided Kubernetes manifests:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: finance-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: finance-app
  template:
    metadata:
      labels:
        app: finance-app
    spec:
      containers:
      - name: finance-app
        image: your-registry/go-finance-advisor:latest
        ports:
        - containerPort: 8081
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: finance-secrets
              key: jwt-secret
        - name: JWT_REFRESH_SECRET
          valueFrom:
            secretKeyRef:
              name: finance-secrets
              key: jwt-refresh-secret
        volumeMounts:
        - name: data-volume
          mountPath: /data
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: finance-data-pvc
```

### Option 3: Standalone Binary
Build and run the binary directly:

```bash
# Build
go build -o finance-advisor ./api

# Run
./finance-advisor
```

## Security Considerations

### 1. HTTPS Configuration
For production, always use HTTPS:

**Using Nginx Reverse Proxy:**
```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Using Let's Encrypt:**
```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d your-domain.com
```

### 2. Environment Variables
Never commit sensitive data to version control. Use:
- Docker secrets
- Environment variables
- Secure key management systems

### 3. Database Security
- SQLite database is stored in `/data/finance.db`
- Ensure proper file permissions (600)
- Regular backups recommended

## Monitoring and Maintenance

### Health Checks
The application includes health check endpoint:
```bash
curl http://localhost:8081/health
```

### Logs
View logs with Docker:
```bash
docker-compose logs -f finance-app
```

### Backup Strategy
Regular database backups:
```bash
# Create backup
docker exec finance-app cp /data/finance.db /data/finance-backup-$(date +%Y%m%d).db

# Restore from backup
docker cp finance-backup-20240101.db finance-app:/data/finance.db
```

### Updates
Update the application:
```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## Performance Optimization

### 1. Database Optimization
- SQLite is suitable for small to medium deployments
- Consider PostgreSQL for larger deployments
- Regular database maintenance (VACUUM)

### 2. Caching
- Implement Redis for session management
- Add CDN for static assets
- Browser caching headers

### 3. Scaling
- Horizontal scaling with load balancer
- Database read replicas
- Microservices architecture for larger deployments

## Troubleshooting

### Common Issues

**Port Already in Use:**
```bash
# Change port in docker-compose.yml
ports:
  - "8082:8081"
```

**Database Permission Issues:**
```bash
# Fix permissions
docker exec finance-app chmod 600 /data/finance.db
```

**JWT Secret Issues:**
- Ensure secrets are at least 32 characters
- Use only alphanumeric characters and symbols
- Regenerate if authentication fails

### Support
For issues and questions:
- Check application logs
- Verify environment configuration
- Test API endpoints with curl
- Check database connectivity

## API Documentation
Complete API documentation is available at:
```
http://your-domain.com/swagger/
```

## Features Summary
- User authentication and authorization
- Financial portfolio management
- Transaction tracking and categorization
- Budget management with progress tracking
- Goal setting and monitoring
- Real-time market data (WebSocket)
- Data export functionality
- Comprehensive reporting
- Mobile-responsive design