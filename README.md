# 🚀 Glimpse Backend (Серверная Часть)

Этот проект является серверной частью (Backend) мобильного приложения [Glimpse](https://github.com/GAZIZlikesLollipop/Glimpse). Инструкции ниже позволяют быстро развернуть полный кластер сервисов на вашем локальном компьютере с использованием **Minikube** и **Kong Ingress Controller**.

---

## ⚡️ Быстрый Старт (Локальное Развертывание)

Для успешного развертывания проекта выполните следующие шаги:

### 1. Подготовка Инструментов

Убедитесь, что у вас установлены необходимые инструменты:

| **Инструмент** | **Команда для проверки**   | **Назначение**                         |
| -------------- | -------------------------- | -------------------------------------- |
| **Minikube**   | `minikube version`         | Запуск локального кластера Kubernetes. |
| **kubectl**    | `kubectl version --client` | Управление кластером.                  |
| **Helm**       | `helm version`             | Управление пакетами (чартами).         |

### 2. Запуск Кластера Minikube

Запустите кластер с драйвером Docker.

```bash
minikube start --driver docker
```

---

### 3. Установка Kong Ingress Controller

Мы используем Kong в качестве API Gateway для маршрутизации и применения политик безопасности (JWT).

1. **Добавьте репозиторий Helm и обновите:**
```bash
helm repo add kong https://charts.konghq.com
helm repo update
```
    
2. **Создайте Namespace и установите Kong:**
```bash
kubectl create namespace kong
helm install kong kong/ingress --namespace kong
```

---

### 4. Подготовка Секретов (TLS и JWT)

Кластер требует наличие двух секретов для работы Ingress (HTTPS) и валидации токенов (JWT).

1. Создайте TLS Secret (для HTTPS Ingress):
    
    Замените <путь_к_вашему_tls.key> и <путь_к_вашему_tls.crt> на реальные пути.
    
```bash
kubectl create secret tls my-tls-secret \
--key <путь_к_вашему_tls.key> \
--cert <путь_к_вашему_tls.crt>
```
    
2. Создайте JWT Secret (для Kong Validator):
    
    Этот секрет должен содержать ваш публичный ключ (public_key) для проверки токенов, выпущенных вашим сервисом авторизации.
    
```bash
kubectl create secret generic my-jwt-secret \
--from-literal=kongCredType=jwt \
--from-literal=rsa_public_key='<ВСТАВЬТЕ_ВАШ_ПУБЛИЧНЫЙ_КЛЮЧ_СЮДА>'
    ```
    
    (Или используйте файл ключа: `--from-file=rsa_public_key=./jwt-public.pem`)

---

### 5. Развертывание Сервисов Проекта

1. Скачайте архив:
    
    Получите ZIP-архив glimpse-k8s.zip (содержит манифесты) из раздела Releases проекта и распакуйте его.
    
2. Примените Манифесты:
    
    Перейдите в папку с манифестами (k8s-manifests) и примените их к кластеру.
    
```bash
# Предполагается, что манифесты находятся в папке k8s-manifests
kubectl apply -f k8s-manifests/
```
    
3. **Убедитесь, что все Поды запущены:**
    
```bash
kubectl get pods
```
    

---

### 6. Активация Сетевого Туннеля

Запустите туннель в **отдельном окне терминала** с правами администратора. Это позволит вам получить доступ к `Ingress` (Kong) через локальный IP-адрес.

Bash

```bash
minikube tunnel
```

> **Внимание:** Не закрывайте это окно!

---

## 🎉 Все Готово!

Ваш кластер **Glimpse Backend** полностью развернут и готов к работе. Теперь вы можете обращаться к API через IP-адрес, предоставленный `minikube tunnel`.