#!/bin/bash
# Let's Encrypt SSL Certificate Setup Script
# Usage: ./init-letsencrypt.sh [staging|production]

set -e

# Configuration
DOMAIN="api.todo.com"
EMAIL="admin@todo.com"  # Replace with actual admin email
NGINX_CONTAINER="nginx"
CERTBOT_CONTAINER="certbot"

# Mode: staging or production
MODE="${1:-staging}"

if [ "$MODE" == "staging" ]; then
    echo "Running in STAGING mode (test certificates)"
    STAGING_ARG="--staging"
else
    echo "Running in PRODUCTION mode"
    STAGING_ARG=""
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker first."
    exit 1
fi

# Create required directories
log_info "Creating required directories..."
mkdir -p ./nginx/certbot/www/.well-known/acme-challenge
mkdir -p ./nginx/certbot/conf
mkdir -p ./nginx/certbot/logs

# Download recommended TLS parameters
if [ ! -f ./nginx/certbot/conf/options-ssl-nginx.conf ]; then
    log_info "Downloading recommended TLS parameters..."
    curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > ./nginx/certbot/conf/options-ssl-nginx.conf
fi

if [ ! -f ./nginx/certbot/conf/ssl-dhparams.pem ]; then
    log_info "Generating DH parameters (this may take a while)..."
    openssl dhparam -out ./nginx/certbot/conf/ssl-dhparams.pem 2048
fi

# Check if certificates already exist
if [ -d "./nginx/certbot/conf/live/$DOMAIN" ]; then
    log_warn "Certificates for $DOMAIN already exist."
    read -p "Do you want to renew them? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Exiting..."
        exit 0
    fi
fi

# Start nginx with temporary configuration for Let's Encrypt challenge
log_info "Starting nginx with temporary configuration..."

# Create temporary nginx config for certificate acquisition
cat > ./nginx/nginx.temp.conf << 'EOF'
server {
    listen 80;
    server_name api.todo.com;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 200 "Let's Encrypt challenge server";
        add_header Content-Type text/plain;
    }
}
EOF

# Start nginx if not running
if ! docker-compose ps | grep -q "nginx.*Up"; then
    log_info "Starting nginx container..."
    docker-compose up -d nginx
    sleep 2
fi

# Request certificate
log_info "Requesting Let's Encrypt certificate for $DOMAIN..."
docker-compose run --rm --entrypoint " \
  certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    $STAGING_ARG \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    -d $DOMAIN \
    --rsa-key-size 4096 \
    --cert-name $DOMAIN \
" certbot

# Check if certificate was created
if [ ! -f "./nginx/certbot/conf/live/$DOMAIN/fullchain.pem" ]; then
    log_error "Certificate creation failed!"
    exit 1
fi

log_info "Certificate created successfully!"

# Reload nginx with production configuration
log_info "Reloading nginx with SSL configuration..."
docker-compose exec nginx nginx -s reload

# Setup auto-renewal
log_info "Setting up auto-renewal cron job..."
CRON_JOB="0 3 * * * cd $(pwd) && docker-compose run --rm certbot renew --quiet && docker-compose exec nginx nginx -s reload"

if ! crontab -l 2>/dev/null | grep -q "certbot renew"; then
    (crontab -l 2>/dev/null; echo "$CRON_JOB") | crontab -
    log_info "Cron job added for automatic renewal"
else
    log_info "Cron job already exists"
fi

# Clean up temporary config
rm -f ./nginx/nginx.temp.conf

# Test SSL configuration
log_info "Testing SSL configuration..."
sleep 2

if command -v openssl &> /dev/null; then
    echo "SSL Certificate Info:"
    openssl s_client -connect $DOMAIN:443 -servername $DOMAIN </dev/null 2>/dev/null | openssl x509 -noout -text | grep -E "Subject:|Issuer:|Not Before|Not After"
fi

log_info "SSL Setup complete!"
log_info ""
log_info "To test with SSL Labs:"
log_info "  https://www.ssllabs.com/ssltest/analyze.html?d=$DOMAIN"
log_info ""
log_info "Remember to:"
log_info "  1. Update your DNS to point to this server"
log_info "  2. Open ports 80 and 443 in your firewall"
log_info "  3. For production, run: ./init-letsencrypt.sh production"
