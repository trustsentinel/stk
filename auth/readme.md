

```
yarn build:frontend
yarn copy:assets
yarn start
```

## Remote access

```
ssh -i stk.vrandkode.net.private.pem root@134.209.24.148
```

## Docker build instructions

### Auth/broker module
```
docker build -f dockerfiles/Dockerfile.auth -t vrandkode/auth . 
docker run -p 8081:8081 -it vrandkode/auth
```

### Web portal (frontend/backend)
```
docker build -t vrandkode/frontend -f Dockerfile.frontend web
docker run -p 9999:9999 -it vrandkode/frontend
```