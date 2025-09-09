echo "Загрузка последней версии backend'а из репозитория"
git pull
echo "Удаление старого контейнера"
docker rm -f minio-photo-service
echo "Удаление старого образа"
docker image rm -f minio-photo-service:latest
echo "Создание нового образа"
docker build -t minio-photo-service:latest .
echo "Запуск контейнера из нового образа"
docker run --name minio-photo-service -d -p 8180:8180 minio-photo-service:latest
docker update --restart=always minio-photo-service
echo "Deploy завершён"