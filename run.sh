docker stop c4i-sso-supervisor
docker rm c4i-sso-supervisor

docker run --pull=always --name c4i-sso-supervisor -d --network bridge -p 11000:8080 -p 12222:4222 -v $(pwd)/app.yaml:/app.yaml registry.c4i.vn/c4i/c4i-sso-supervisor:dev

sleep 1

docker logs -f c4i-sso-supervisor
