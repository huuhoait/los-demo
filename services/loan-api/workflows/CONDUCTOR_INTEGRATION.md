# Netflix Conductor Integration Guide

This guide explains how to set up and use Netflix Conductor with the loan service for workflow orchestration.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Loan Service  │    │ Conductor Server │    │ Conductor UI    │
│   (Port 8081)   │◄──►│   (Port 8080)    │◄──►│  (Port 3000)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   PostgreSQL     │    │   Workflow      │
│  (Port 5433)    │    │   (Port 5432)    │    │   Definitions   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### 1. Start All Services

```bash
# Start all services with Docker Compose
./docker-run.sh compose

# Check health of all services
./docker-run.sh health
```

### 2. Deploy Workflows

```bash
# Deploy workflows to Conductor server
./docker-run.sh deploy-workflows

# Or manually deploy
./workflows/deploy-conductor.sh -s http://localhost:8080
```

### 3. Test Integration

```bash
# Test Conductor integration
./workflows/test-conductor.sh

# Check Conductor status
./docker-run.sh check-conductor
```

### 4. Access Services

- **Loan Service**: http://localhost:8081
- **Conductor Server**: http://localhost:8082
- **Conductor UI**: http://localhost:3000 (when available)
- **Swagger Docs**: http://localhost:8081/swagger/index.html

## 🔧 Configuration

### Conductor Server Configuration

The Conductor server is configured via `workflows/conductor.properties`:

```properties
# Database Configuration
conductor.db.url=jdbc:postgresql://postgres:5432/conductor
conductor.db.username=postgres
conductor.db.password=password

# Server Configuration
conductor.server.port=8080
conductor.server.host=0.0.0.0

# Performance Configuration
conductor.workflow.max.parallel.workers=10
conductor.task.max.parallel.workers=20
```

### Environment Variables

Key environment variables for the loan service:

```bash
# Conductor Configuration
CONDUCTOR_BASE_URL=http://conductor-server:8080
CONDUCTOR_TIMEOUT=30

# Environment
APP_ENV=docker
```

## 📋 Available Workflows

### 1. Pre-qualification Workflow

**Purpose**: Quick loan eligibility assessment
**Duration**: ~30 seconds
**Tasks**: 5 sequential tasks

```bash
# Start pre-qualification workflow
curl -X POST http://localhost:8082/api/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prequalification_workflow",
    "version": 1,
    "input": {
      "userId": "user123",
      "loanAmount": 30000,
      "annualIncome": 60000,
      "monthlyDebt": 500,
      "employmentStatus": "full_time"
    }
  }'
```

### 2. Loan Processing Workflow

**Purpose**: Complete loan application processing
**Duration**: 1-3 days (includes human tasks)
**Tasks**: 8 tasks with sub-workflow integration

```bash
# Start loan processing workflow
curl -X POST http://localhost:8082/api/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "name": "loan_processing_workflow",
    "version": 1,
    "input": {
      "applicationId": "app123",
      "userId": "user123",
      "loanAmount": 30000,
      "loanPurpose": "personal",
      "annualIncome": 60000,
      "monthlyIncome": 5000,
      "monthlyDebt": 500,
      "requestedTerm": 48,
      "currentState": "initiated"
    }
  }'
```

### 3. Underwriting Workflow

**Purpose**: Risk assessment and decision making
**Duration**: 15 minutes - 2 days
**Tasks**: Complex decision tree with manual review paths

## 🔌 API Endpoints

### Loan Service Workflow Endpoints

```bash
# Get workflow status
GET /v1/workflows/{id}/status

# Pause workflow
POST /v1/workflows/{id}/pause

# Resume workflow
POST /v1/workflows/{id}/resume

# Terminate workflow
POST /v1/workflows/{id}/terminate
```

### Conductor Server Endpoints

```bash
# Health check
GET /health

# Start workflow
POST /api/workflow

# Get workflow status
GET /api/workflow/{id}

# Get workflow metadata
GET /api/metadata/workflow

# Get task definitions
GET /api/metadata/taskdefs
```

## 🧪 Testing

### Manual Testing

```bash
# Test workflow status endpoint
curl -X GET http://localhost:8081/v1/workflows/test123/status

# Test workflow pause
curl -X POST http://localhost:8081/v1/workflows/test123/pause

# Test workflow resume
curl -X POST http://localhost:8081/v1/workflows/test123/resume

# Test workflow termination
curl -X POST http://localhost:8081/v1/workflows/test123/terminate \
  -H "Content-Type: application/json" \
  -d '{"reason": "Testing termination"}'
```

### Automated Testing

```bash
# Run comprehensive integration tests
./workflows/test-conductor.sh
```

## 📊 Monitoring

### Health Checks

```bash
# Check all services health
./docker-run.sh health

# Check Conductor specifically
./docker-run.sh check-conductor
```

### Logs

```bash
# View all service logs
./docker-run.sh compose-logs

# View specific service logs
docker-compose logs -f conductor-server
docker-compose logs -f loan-service
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Conductor Server Not Starting

```bash
# Check PostgreSQL connection
docker-compose exec postgres psql -U postgres -d conductor

# Check Conductor logs
docker-compose logs conductor-server

# Verify configuration
docker-compose exec conductor-server cat /app/conductor.properties
```

#### 2. Workflow Deployment Fails

```bash
# Check Conductor health
curl -f http://localhost:8082/health

# Check API endpoints
curl -f http://localhost:8080/api/metadata/workflow

# Redeploy workflows
./workflows/deploy-conductor.sh -s http://localhost:8080
```

#### 3. Loan Service Cannot Connect to Conductor

```bash
# Check network connectivity
docker-compose exec loan-service ping conductor-server

# Check environment variables
docker-compose exec loan-service env | grep CONDUCTOR

# Restart loan service
docker-compose restart loan-service
```

### Debug Commands

```bash
# Check container status
docker-compose ps

# Check network
docker network ls
docker network inspect loan_loan-network

# Check volumes
docker volume ls
docker volume inspect loan_conductor_data
```

## 🔄 Workflow Lifecycle

### 1. Workflow Creation

```bash
# Deploy workflow definitions
./workflows/deploy-conductor.sh

# Verify deployment
curl -s http://localhost:8082/api/metadata/workflow | jq 'length'
```

### 2. Workflow Execution

```bash
# Start workflow
curl -X POST http://localhost:8080/api/workflow \
  -H "Content-Type: application/json" \
  -d @workflow-input.json

# Monitor execution
curl -s http://localhost:8082/api/workflow/{workflowId}
```

### 3. Workflow Management

```bash
# Pause workflow
curl -X POST http://localhost:8081/v1/workflows/{id}/pause

# Resume workflow
curl -X POST http://localhost:8081/v1/workflows/{id}/resume

# Terminate workflow
curl -X POST http://localhost:8081/v1/workflows/{id}/terminate \
  -d '{"reason": "User requested"}'
```

## 📈 Performance Tuning

### Conductor Server Tuning

```properties
# Increase worker threads
conductor.workflow.max.parallel.workers=20
conductor.task.max.parallel.workers=40

# Optimize caching
conductor.workflow.def.cache.size=200
conductor.workflow.def.cache.ttl.seconds=1200

# Queue optimization
conductor.queue.workflow.size=2000
conductor.queue.task.size=2000
```

### Database Tuning

```sql
-- Create indexes for better performance
CREATE INDEX idx_workflow_execution_status ON workflow_execution(status);
CREATE INDEX idx_task_execution_status ON task_execution(status);
CREATE INDEX idx_workflow_execution_created ON workflow_execution(created_on);
```

## 🔐 Security

### Current Configuration

- **Authentication**: Disabled (development mode)
- **Authorization**: Disabled (development mode)
- **HTTPS**: Not configured

### Production Security

```properties
# Enable security
conductor.security.enabled=true
conductor.security.authentication.enabled=true
conductor.security.authorization.enabled=true

# Configure authentication
conductor.security.authentication.type=jwt
conductor.security.authentication.jwt.secret=your-secret-key

# Configure authorization
conductor.security.authorization.type=rbac
```

## 📚 Additional Resources

### Documentation

- [Netflix Conductor Documentation](https://conductor.netflix.com/)
- [Conductor API Reference](https://conductor.netflix.com/api-docs/)
- [Workflow Design Patterns](https://conductor.netflix.com/workflow-patterns/)

### Examples

- [Workflow Definitions](./)
- [Task Definitions](./tasks/)
- [Integration Examples](./examples/)

### Support

- [Conductor GitHub](https://github.com/Netflix/conductor)
- [Conductor Community](https://conductor.netflix.com/community/)
- [Issue Tracker](https://github.com/Netflix/conductor/issues)

---

## 🎯 Next Steps

1. **Deploy to Production**: Configure security and monitoring
2. **Add More Workflows**: Implement additional business processes
3. **Performance Optimization**: Monitor and tune based on usage
4. **Integration Testing**: Add comprehensive test coverage
5. **Monitoring & Alerting**: Set up production monitoring

For questions or issues, please refer to the troubleshooting section or create an issue in the project repository.
