#!/bin/bash

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
print_step() { echo -e "${BLUE}▶${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_info() { echo -e "${YELLOW}ℹ${NC} $1"; }

echo "==========================================="
echo "🚀 RESERV-SERVICE DEPLOYMENT TO MARKET NAMESPACE"
echo "==========================================="

# 1. Проверка Minikube
print_step "Checking Minikube status..."
if ! minikube status | grep -q "Running"; then
    print_error "Minikube is not running. Starting Minikube..."
    minikube start --cpus=2 --memory=4096
    if [ $? -ne 0 ]; then
        print_error "Failed to start Minikube"
        exit 1
    fi
fi
print_success "Minikube is running"

# 2. Настройка Docker для Minikube
print_step "Configuring Docker for Minikube..."
eval $(minikube docker-env)
if [ $? -ne 0 ]; then
    print_error "Failed to configure Docker"
    exit 1
fi
print_success "Docker configured for Minikube"

# 3. Сборка Docker образа
print_step "Building Docker image..."
if docker build -t reserv-service:latest .; then
    print_success "Image built successfully: reserv-service:latest"
else
    print_error "Docker build failed"
    exit 1
fi

# 4. Проверка namespace market
print_step "Checking namespace market..."
if ! kubectl get namespace market >/dev/null 2>&1; then
    print_error "Namespace 'market' not found"
    echo "Available namespaces:"
    kubectl get namespaces
    exit 1
fi
print_success "Namespace 'market' exists"

# 5. Удаляем старые deployments reserv-service в namespace market
print_step "Cleaning up OLD reserv-service deployments in market namespace..."
RESERV_DEPLOYMENTS=$(kubectl get deployments -n market --no-headers 2>/dev/null | awk '{print $1}' | grep "^reserv-service")

if [ -n "$RESERV_DEPLOYMENTS" ]; then
    echo "Found old reserv-service deployments to delete:"
    for DEPLOY in $RESERV_DEPLOYMENTS; do
        echo "  - $DEPLOY"
        kubectl delete deployment -n market "$DEPLOY" --ignore-not-found
    done
    print_success "Old deployments deleted"
    
    # Ждем удаления подов
    print_info "Waiting for old pods to terminate..."
    sleep 5
    
    # Форсированное удаление старых подов
    kubectl delete pods -n market -l app=reserv-service --ignore-not-found --force --grace-period=0 2>/dev/null
    sleep 2
else
    print_success "No old reserv-service deployments found"
fi

# 6. Создаём новый deployment в namespace market
TIMESTAMP=$(date +%Y%m%d%H%M%S)
NEW_DEPLOYMENT_NAME="reserv-service-v${TIMESTAMP}"

print_step "Creating NEW deployment in market namespace: ${NEW_DEPLOYMENT_NAME}..."
cat <<YAML | kubectl apply -n market -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${NEW_DEPLOYMENT_NAME}
  namespace: market
spec:
  replicas: 1
  selector:
    matchLabels:
      app: reserv-service
      version: "v${TIMESTAMP}"
  template:
    metadata:
      labels:
        app: reserv-service
        version: "v${TIMESTAMP}"
    spec:
      containers:
      - name: reserv-service
        image: reserv-service:latest
        imagePullPolicy: Never
        ports:
        - containerPort: 8074
        env:
        - name: DB_HOST
          value: "postgres"  # имя сервиса PostgreSQL в namespace market
        - name: DB_PORT
          value: "5432"
        - name: DB_NAME
          value: "marketdb"
        - name: DB_USER
          value: "admin"
        - name: DB_PASSWORD
          value: "admin123"
        - name: DB_SSLMODE
          value: "disable"
        - name: APP_PORT
          value: "8074"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        readinessProbe:
          httpGet:
            path: /reserv/health
            port: 8074
          initialDelaySeconds: 15
          periodSeconds: 10
          timeoutSeconds: 3
        livenessProbe:
          httpGet:
            path: /reserv/health
            port: 8074
          initialDelaySeconds: 30
          periodSeconds: 20
YAML
print_success "Deployment ${NEW_DEPLOYMENT_NAME} created in market namespace"

# 7. Service в namespace market (создаем или обновляем)
print_step "Creating/Updating service in market namespace..."
cat <<YAML | kubectl apply -n market -f -
apiVersion: v1
kind: Service
metadata:
  name: reserv-service
  namespace: market
spec:
  selector:
    app: reserv-service
    version: "v${TIMESTAMP}"
  ports:
  - port: 8074
    targetPort: 8074
    protocol: TCP
  type: ClusterIP
YAML
print_success "Service reserv-service ready in market namespace"

# 8. Ждём запуска новой поды
print_step "Waiting for NEW pod in market namespace..."
MAX_WAIT=90
POD_READY=false
for i in $(seq 1 $MAX_WAIT); do
    POD_NAME=$(kubectl get pods -n market -l version=v${TIMESTAMP} -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [ -n "$POD_NAME" ]; then
        POD_STATUS=$(kubectl get pod -n market "$POD_NAME" -o jsonpath='{.status.phase}' 2>/dev/null)
        POD_READY_STATE=$(kubectl get pod -n market "$POD_NAME" -o jsonpath='{.status.containerStatuses[0].ready}' 2>/dev/null)
        
        if [[ "$POD_STATUS" == "Running" ]] && [[ "$POD_READY_STATE" == "true" ]]; then
            print_success "✅ New pod $POD_NAME is running and ready!"
            POD_READY=true
            break
        fi
    fi
    
    if [ $i -eq $MAX_WAIT ]; then
        print_error "❌ Timeout waiting for pod"
        echo "Current pods in market namespace:"
        kubectl get pods -n market
        echo ""
        echo "Checking deployment status:"
        kubectl describe deployment -n market ${NEW_DEPLOYMENT_NAME} | tail -30
        echo ""
        echo "Pod logs:"
        if [ -n "$POD_NAME" ]; then
            kubectl logs -n market $POD_NAME --tail=20
        fi
        exit 1
    fi
    
    if [ $((i % 10)) -eq 0 ]; then
        echo -n "${i}s"
    else
        echo -n "."
    fi
    sleep 1
done
echo ""

# 9. Проверка статуса
print_step "Checking status in market namespace..."
echo ""
echo "📊 CURRENT PODS IN MARKET:"
kubectl get pods -n market -o wide
echo ""
echo "📊 CURRENT DEPLOYMENTS IN MARKET:"
kubectl get deployments -n market
echo ""
echo "📊 SERVICES IN MARKET:"
kubectl get services -n market

# 10. Тестирование приложения
print_step "Testing application..."
# Запускаем port-forward в фоне
kubectl port-forward -n market svc/reserv-service 8074:8074 > /dev/null 2>&1 &
PF_PID=$!
sleep 5

echo "Testing health endpoint..."
if curl -s --max-time 10 http://localhost:8074/reserv/health > /dev/null 2>&1; then
    print_success "✅ App is responding!"
    echo ""
    echo "   Health check response:"
    curl -s http://localhost:8074/reserv/health | jq . 2>/dev/null || curl -s http://localhost:8074/reserv/health
    echo ""
else
    print_error "❌ App not responding on health endpoint"
    echo ""
    echo "Checking logs..."
    kubectl logs -n market -l version=v${TIMESTAMP} --tail=20
    echo ""
    echo "Checking pod events..."
    POD_NAME=$(kubectl get pods -n market -l version=v${TIMESTAMP} -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -n "$POD_NAME" ]; then
        kubectl describe pod -n market $POD_NAME | grep -A 10 "Events:"
    fi
fi

# Убиваем port-forward
kill $PF_PID 2>/dev/null

# 11. Очистка других deployment reserv-service (кроме текущего)
print_step "Final cleanup of other reserv-service deployments in market..."
OTHER_DEPLOYMENTS=$(kubectl get deployments -n market --no-headers 2>/dev/null | awk '{print $1}' | grep "^reserv-service" | grep -v "^${NEW_DEPLOYMENT_NAME}$")

if [ -n "$OTHER_DEPLOYMENTS" ]; then
    echo "Found other reserv-service deployments to clean up:"
    for DEPLOY in $OTHER_DEPLOYMENTS; do
        echo "  - $DEPLOY"
        kubectl delete deployment -n market "$DEPLOY" --ignore-not-found
    done
    print_success "Other reserv-service deployments cleaned"
else
    print_success "No other reserv-service deployments found"
fi

echo ""
echo "==========================================="
echo "✅ DEPLOYMENT COMPLETE!"
echo "==========================================="
echo ""
echo "📌 Summary:"
echo "   • Namespace: market"
echo "   • New Deployment: ${NEW_DEPLOYMENT_NAME}"
echo "   • Version: v${TIMESTAMP}"
echo "   • Image: reserv-service:latest"
echo "   • Service: reserv-service:8074"
echo "   • Health: /reserv/health"
echo ""
echo "🌐 Access from Postman:"
echo "   Method 1 (Port-forward):"
echo "     1. kubectl port-forward -n market svc/reserv-service 8074:8074"
echo "     2. Use: http://localhost:8074/reserv/health"
echo ""
echo "   Method 2 (Minikube IP):"
echo "     1. minikube service -n market reserv-service --url"
echo ""
echo "📊 Commands to check resources:"
echo "   kubectl get pods -n market"
echo "   kubectl get deployments -n market"
echo "   kubectl get services -n market"
echo "   kubectl logs -n market -l app=reserv-service --tail=20"
echo ""
echo "🔄 To update/rollback:"
echo "   Run this script again!"
echo ""
echo "🔧 Troubleshooting:"
echo "   If DB connection fails, check PostgreSQL service in market namespace:"
echo "   kubectl get services -n market | grep postgres"
echo ""