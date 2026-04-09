# Security Deployment Guide

## Prerequisites

- Docker and Docker Compose installed
- Domain name pointing to server (api.todo.com)
- Ports 80 and 443 open in firewall

## Initial SSL Setup

1. **Run the Let's Encrypt initialization script:**
```bash
cd /Users/marvin/Projekte/todo
./scripts/init-letsencrypt.sh staging
```

2. **Test with staging certificate first**, then switch to production:
```bash
./scripts/init-letsencrypt.sh production
```

3. **Verify SSL configuration:**
```bash
# Check headers
curl -I https://api.todo.com/health

# Test with SSL Labs (after DNS propagation)
# Visit: https://www.ssllabs.com/ssltest/analyze.html?d=api.todo.com
```

## Environment Configuration

Create `.env` file with these security-related variables:

```bash
# JWT Configuration
JWT_SECRET=<generate-strong-secret-64-chars>
JWT_SECRET_PREVIOUS=<previous-secret-for-rotation>
JWT_ROTATION_WINDOW=86400

# Database (use SSL in production)
DB_URL=postgres://user:password@db:5432/todo?sslmode=require

# Redis
REDIS_URL=redis://redis:6379

# Logging
LOG_LEVEL=info
```

## Security Headers Verification

Expected response headers:
```
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; ...
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
```

## Rate Limiting Verification

Test rate limits:
```bash
# Auth endpoints - should allow 5 req/min
for i in {1..10}; do curl -X POST https://api.todo.com/auth/login; done

# API endpoints - should allow 100 req/min
for i in {1..110}; do curl https://api.todo.com/health; done
```

Expected: HTTP 429 Too Many Requests after limits exceeded

## Security Monitoring

### Check Logs for Security Events
```bash
# View failed authentication attempts
docker-compose logs api | grep "client error"

# View rate limit blocks
docker-compose logs api | grep "rate_limit"
```

### Monitor Redis for Rate Limits
```bash
docker-compose exec redis redis-cli KEYS "ratelimit:*"
```

## SSL Certificate Renewal

Automatic renewal is configured via cron. To verify:
```bash
crontab -l | grep certbot
```

Manual renewal:
```bash
docker-compose run --rm certbot renew
docker-compose exec nginx nginx -s reload
```

## JWT Secret Rotation

To rotate JWT secrets without downtime:

1. Set new secret:
```bash
export JWT_SECRET_PREVIOUS=$JWT_SECRET
export JWT_SECRET=<new-strong-secret>
```

2. Restart API:
```bash
docker-compose up -d api
```

3. After rotation window (default 24h), clear previous:
```bash
unset JWT_SECRET_PREVIOUS
docker-compose up -d api
```

## Incident Response

### Block Malicious IP
```bash
# Via Redis (temporary)
docker-compose exec redis redis-cli SET ratelimit:blocked:ip:<IP> 1 EX 3600
```

### Reset Rate Limit for User
```bash
# API endpoint (admin only)
curl -X POST https://api.todo.com/admin/ratelimit/reset \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"client_id": "user:<user-id>"}'
```

## Security Checklist Before Production

- [ ] SSL Labs A+ rating achieved
- [ ] All security headers present
- [ ] Rate limiting working
- [ ] Password policy enforced
- [ ] Sensitive data not in logs
- [ ] JWT secrets rotated
- [ ] API keys hashed in database
- [ ] Automated certificate renewal configured
- [ ] Firewall configured (only 80, 443 open)
- [ ] Environment variables secured
- [ ] Database using SSL
- [ ] Backup and recovery tested

## Security Contacts

For security issues, contact: security@todo.com
