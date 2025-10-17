# Troubleshooting Guide

Common issues and solutions for the Demo App IoT Data Pipeline.

## Table of Contents

- [Connection Issues](#connection-issues)
- [Service Not Starting](#service-not-starting)
- [Message Queue Issues](#message-queue-issues)
- [Database Issues](#database-issues)
- [Performance Issues](#performance-issues)
- [Container Issues](#container-issues)
- [Kubernetes Issues](#kubernetes-issues)

## Connection Issues

### Cannot Connect to RabbitMQ

**Symptoms**:
```
Error: failed to connect to AMQP server: dial tcp: connection refused
```

**Checks**:
```bash
# 1. Verify RabbitMQ is running
docker ps | grep rabbitmq

# 2. Test connection
telnet localhost 5672

# 3. Check logs
docker logs rabbitmq

# 4. Verify URL format
# Correct: amqp://localhost:5672
# Wrong: localhost:5672 (missing protocol)
```

**Solutions**:

**If RabbitMQ is not running**:
```bash
docker start rabbitmq
# or
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine
```

**If connection string is wrong**:
```bash
# Correct format
./demo-app backend --rabbitmq-url=amqp://guest:guest@localhost:5672

# With credentials
./demo-app backend --rabbitmq-url=amqp://user:pass@rabbitmq.example.com:5672
```

**Check RabbitMQ is accessible**:
```bash
# Web UI
curl http://localhost:15672
# Should return HTML

# AMQP port
nc -zv localhost 5672
# Should say "succeeded"
```

---

### Cannot Connect to PostgreSQL

**Symptoms**:
```
Error: failed to connect to database: dial tcp: connection refused
```

**Checks**:
```bash
# 1. Verify PostgreSQL is running
docker ps | grep postgres

# 2. Test connection
psql -h localhost -U postgres -d iot_db

# 3. Check logs
docker logs postgres
```

**Solutions**:

**If PostgreSQL is not running**:
```bash
docker start postgres
# or
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=iot_db \
  -p 5432:5432 \
  postgres:16-alpine
```

**Test connection manually**:
```bash
# Using psql
psql -h localhost -U postgres -d iot_db -c "SELECT 1"

# Using docker exec
docker exec -it postgres psql -U postgres -d iot_db -c "SELECT 1"
```

**Check connection string**:
```bash
# Correct format
./demo-app backend \
  --db-host=localhost \
  --db-port=5432 \
  --db-user=postgres \
  --db-password=postgres \
  --db-name=iot_db \
  --db-sslmode=disable
```

**SSL Mode Issues**:
- Local development: Use `sslmode=disable`
- Production: Use `sslmode=require` or `sslmode=verify-full`

---

### Frontend Cannot Connect to Backend

**Symptoms**:
```
Error: rpc error: code = Unavailable desc = connection error
```

**Checks**:
```bash
# 1. Verify backend is running
curl http://localhost:9090/metrics

# 2. Test gRPC connection
grpcurl -plaintext localhost:50051 list

# 3. Check backend logs
# Look for "gRPC server started"
```

**Solutions**:

**If backend is not running**:
```bash
./demo-app backend
```

**Check backend URL**:
```bash
# Correct (localhost)
./demo-app frontend --backend-url=localhost:50051

# Correct (Docker network)
./demo-app frontend --backend-url=demo-backend:50051

# Wrong (missing port)
./demo-app frontend --backend-url=localhost
```

**Network connectivity**:
```bash
# From frontend container
docker exec -it demo-frontend ping demo-backend

# Port check
nc -zv localhost 50051
```

## Service Not Starting

### Backend Fails to Start

**Symptoms**:
```
Error: failed to auto-migrate database
```

**Cause**: Database migration failure

**Solution**:
```bash
# Check database is accessible
psql -h localhost -U postgres -d iot_db

# Manually run migrations (if needed)
# The backend will auto-migrate on startup

# Check for existing tables
docker exec -it postgres psql -U postgres -d iot_db -c "\dt"

# Drop and recreate (WARNING: loses data)
docker exec -it postgres psql -U postgres -d iot_db -c "DROP TABLE sensor_readings CASCADE;"
docker exec -it postgres psql -U postgres -d iot_db -c "DROP TABLE iot_devices CASCADE;"
```

---

### Generator Fails to Start

**Symptoms**:
```
Error: maximum retry attempts exceeded
```

**Cause**: Cannot connect to RabbitMQ after 5 retry attempts

**Solution**:
```bash
# 1. Verify RabbitMQ is running and accessible
docker ps | grep rabbitmq

# 2. Check connection string
echo $APP_GENERATOR_RABBITMQ_URL

# 3. Test connection manually
curl http://localhost:15672/api/vhosts
# Should return JSON (use guest/guest for auth)

# 4. Restart generator after RabbitMQ is ready
./demo-app generator --rabbitmq-url=amqp://localhost:5672
```

---

### Port Already in Use

**Symptoms**:
```
Error: bind: address already in use
```

**Solution**:
```bash
# Find process using port
lsof -i :8080
lsof -i :50051
lsof -i :9090
lsof -i :9091

# Kill process
kill -9 <PID>

# Or use different port
./demo-app frontend --http-port=8081
./demo-app backend --grpc-port=50052 --metrics-port=9092
./demo-app generator --metrics-port=9093
```

## Message Queue Issues

### Messages Not Being Consumed

**Symptoms**:
- RabbitMQ queue depth increasing
- No database entries

**Checks**:
```bash
# 1. Check RabbitMQ management UI
open http://localhost:15672
# Login: guest/guest
# Go to "Queues" tab

# 2. Check backend consumer logs
# Should see "device consumer started" and "sensor consumer started"

# 3. Check active consumers
curl http://localhost:9090/metrics | grep active_consumers
```

**Solutions**:

**If consumers are not running**:
```bash
# Restart backend
./demo-app backend
```

**If consumers are stuck**:
```bash
# Check for errors in logs
# Look for database connection issues
# Look for message deserialization errors

# Restart backend to reset consumers
```

**Purge queue** (WARNING: loses messages):
```bash
# Via management UI: Go to queue → "Purge Messages"

# Via CLI
docker exec rabbitmq rabbitmqctl purge_queue sensor-data
docker exec rabbitmq rabbitmqctl purge_queue device-data
```

---

### Queue Not Declared

**Symptoms**:
```
Error: channel/connection is not open
```

**Cause**: Queue doesn't exist or was deleted

**Solution**:
```bash
# Queues are auto-declared by generator and backend
# Restart both services:

# 1. Start backend first (declares queues as consumer)
./demo-app backend

# 2. Start generator (publishes to queues)
./demo-app generator
```

## Database Issues

### Migration Failures

**Symptoms**:
```
Error: relation "iot_devices" already exists
```

**Solution**:
```bash
# GORM migrations are idempotent
# This error usually means tables exist but schema changed

# Check current schema
docker exec -it postgres psql -U postgres -d iot_db -c "\d iot_devices"
docker exec -it postgres psql -U postgres -d iot_db -c "\d sensor_readings"

# Force recreate (WARNING: loses data)
docker exec -it postgres psql -U postgres -d iot_db << EOF
DROP TABLE IF EXISTS sensor_readings CASCADE;
DROP TABLE IF EXISTS iot_devices CASCADE;
EOF

# Restart backend to recreate tables
./demo-app backend
```

---

### Foreign Key Constraint Violation

**Symptoms**:
```
Error: foreign key constraint "fk_sensor_readings_device" violated
```

**Cause**: Trying to insert sensor reading for non-existent device

**Solution**:
```bash
# Ensure device is created first
# The system automatically creates devices before readings

# If you're manually inserting data:
# 1. Insert device first
# 2. Then insert sensor reading

# Check existing devices
docker exec -it postgres psql -U postgres -d iot_db \
  -c "SELECT device_id FROM iot_devices"

# Delete orphaned readings (if any)
docker exec -it postgres psql -U postgres -d iot_db \
  -c "DELETE FROM sensor_readings WHERE device_id NOT IN (SELECT device_id FROM iot_devices)"
```

---

### Database Connection Pool Exhausted

**Symptoms**:
```
Error: too many clients already
```

**Solution**:
```bash
# Check current connections
docker exec -it postgres psql -U postgres -d iot_db \
  -c "SELECT count(*) FROM pg_stat_activity"

# Check max connections
docker exec -it postgres psql -U postgres -d iot_db \
  -c "SHOW max_connections"

# Increase max connections (PostgreSQL config)
# Edit postgresql.conf: max_connections = 200

# Or use connection pooling (PgBouncer)
# Or scale backend horizontally
```

## Performance Issues

### High Memory Usage

**Symptoms**:
- Service consuming >1GB RAM
- OOMKilled in Kubernetes

**Checks**:
```bash
# Check memory usage
docker stats demo-backend

# Check Go metrics
curl http://localhost:9090/metrics | grep go_memstats

# Check goroutines
curl http://localhost:9090/metrics | grep go_goroutines
```

**Solutions**:

**Set Go memory limit**:
```bash
export GOMEMLIMIT=512MiB
./demo-app backend
```

**Reduce concurrent operations**:
```bash
# Reduce number of devices
./demo-app generator --num-devices=10

# Reduce producer interval
./demo-app generator --interval=30s
```

**Check for goroutine leaks**:
```bash
# Download pprof profile
curl http://localhost:9090/debug/pprof/goroutine > goroutine.prof

# Analyze with go tool
go tool pprof goroutine.prof
```

---

### High CPU Usage

**Symptoms**:
- Service consuming >80% CPU
- Slow response times

**Checks**:
```bash
# Check CPU usage
docker stats demo-backend

# Check request rate
curl http://localhost:9090/metrics | grep requests_total
```

**Solutions**:

**Reduce load**:
```bash
# Reduce generation rate
./demo-app generator --interval=60s

# Scale horizontally (Kubernetes)
kubectl scale deployment backend --replicas=3
```

**Profile CPU**:
```bash
# Download CPU profile (30 seconds)
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze
go tool pprof cpu.prof
```

---

### Slow gRPC Responses

**Symptoms**:
- Frontend timeouts
- High p95 latency

**Checks**:
```bash
# Check latency metrics
curl http://localhost:9090/metrics | grep grpc_request_duration

# Check database query time
# Enable query logging in PostgreSQL
```

**Solutions**:

**Add database indexes**:
```sql
-- Already included in migrations
CREATE INDEX idx_device_timestamp ON sensor_readings(device_id, timestamp);
CREATE INDEX idx_timestamp ON sensor_readings(timestamp);
```

**Optimize queries**:
```bash
# Use pagination for large result sets
# Default page_size is 50, max is 1000

# Avoid fetching all devices if not needed
# Use GetDevice instead of GetAllDevice when possible
```

**Scale backend**:
```bash
# Multiple backend instances
# Load balance with Kubernetes Service
```

## Container Issues

### Container Crashes on Startup

**Symptoms**:
```
Error: standard_init_linux.go: exec user process caused: no such file or directory
```

**Cause**: Binary compiled for wrong architecture

**Solution**:
```bash
# Build for correct architecture
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/demo-app ./cmd

# Or use Docker buildx for multi-arch
docker buildx build --platform linux/amd64,linux/arm64 -t demo-app:latest .
```

---

### Cannot Pull Image

**Symptoms**:
```
Error: failed to pull image "ghcr.io/procodus/demo-app:latest": unauthorized
```

**Solution**:
```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Or use public image (if available)
docker pull ghcr.io/procodus/demo-app:latest

# Or build locally
docker build -t demo-app:latest -f deployments/Dockerfile .
```

## Kubernetes Issues

### Pods in CrashLoopBackOff

**Checks**:
```bash
# Check pod status
kubectl get pods -n demo-app

# View pod logs
kubectl logs -n demo-app -l app=backend --tail=50

# Describe pod
kubectl describe pod -n demo-app <pod-name>

# Check events
kubectl get events -n demo-app --sort-by='.lastTimestamp'
```

**Common Causes**:
1. Cannot connect to database
2. Cannot connect to RabbitMQ
3. Missing environment variables
4. Out of memory (OOMKilled)

**Solutions**:

**Check configuration**:
```bash
kubectl get configmap -n demo-app
kubectl get secret -n demo-app
```

**Verify environment variables**:
```bash
kubectl exec -n demo-app -it <pod-name> -- env | grep APP_
```

**Check resource limits**:
```bash
kubectl describe pod -n demo-app <pod-name> | grep -A5 Limits
```

---

### Service Not Accessible

**Symptoms**:
- Cannot access frontend from browser
- Services cannot communicate

**Checks**:
```bash
# Check services
kubectl get svc -n demo-app

# Check endpoints
kubectl get endpoints -n demo-app

# Test from within cluster
kubectl run -n demo-app -it --rm debug --image=busybox --restart=Never -- wget -O- http://backend:50051
```

**Solutions**:

**For external access** (LoadBalancer or Ingress):
```bash
# Check ingress
kubectl get ingress -n demo-app

# Port forward for testing
kubectl port-forward -n demo-app svc/frontend 8080:8080
```

**For internal communication**:
```bash
# Use service DNS name
# Format: <service-name>.<namespace>.svc.cluster.local
APP_FRONTEND_BACKEND_URL=backend.demo-app.svc.cluster.local:50051
```

---

### Persistent Volume Issues

**Symptoms**:
```
Error: pod has unbound immediate PersistentVolumeClaims
```

**Solution**:
```bash
# Check PVCs
kubectl get pvc -n demo-app

# Check PVs
kubectl get pv

# Describe PVC to see error
kubectl describe pvc -n demo-app postgres-pvc

# For local testing, use hostPath or local provisioner
# For production, use cloud provider storage classes
```

## Debugging Tools

### Log Analysis

```bash
# Grep for errors
kubectl logs -n demo-app -l app=backend | grep ERROR

# Follow logs
kubectl logs -n demo-app -l app=backend -f

# Previous container logs
kubectl logs -n demo-app <pod-name> --previous

# All containers in pod
kubectl logs -n demo-app <pod-name> --all-containers
```

### Network Debugging

```bash
# Test DNS resolution
kubectl run -n demo-app -it --rm debug --image=busybox --restart=Never -- nslookup backend

# Test connectivity
kubectl run -n demo-app -it --rm debug --image=busybox --restart=Never -- nc -zv backend 50051

# Curl test
kubectl run -n demo-app -it --rm debug --image=curlimages/curl --restart=Never -- curl http://backend:9090/metrics
```

### Metrics Inspection

```bash
# Check Prometheus targets
curl http://localhost:9000/api/v1/targets

# Query specific metric
curl 'http://localhost:9000/api/v1/query?query=demo_app_backend_active_consumers'

# Export all metrics
curl http://localhost:9090/metrics > backend-metrics.txt
```

## Getting Help

If you're still experiencing issues:

1. **Check logs**: Enable debug logging with `--log-level=debug`
2. **Check metrics**: View Prometheus metrics for service health
3. **Search issues**: [GitHub Issues](https://github.com/procodus/demo-app/issues)
4. **Ask community**: [GitHub Discussions](https://github.com/procodus/demo-app/discussions)
5. **Report bug**: Create a new issue with:
   - Service version
   - Error message
   - Steps to reproduce
   - Environment (Docker, K8s, local)

## Next Steps

- [Monitoring Guide](./monitoring.md) - Set up monitoring to prevent issues
- [Performance Guide](./performance.md) - Optimize for better performance
- [Configuration Guide](./configuration.md) - Tune settings
