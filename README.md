# 🎵 Lyrics Library API

RESTful microservice in Go for receiving song lyrics with Russian translation.

## Features
- Getting song lyrics by artist and track title
- Automatic translation into Russian

## Stack
- **Language**: Go 1.24+
- **Database**: PostgreSQL
- **Caching**: Redis
- **Migrations**: golang-migrate
- **External APIs**:
- [LyricsOVH](https://lyricsovh.docs.apiary.io/#reference) - getting lyrics
- [Yandex.Translate](https://yandex.cloud/ru/docs/translate/quickstart) - translation into Russian
- **Containerization**: Docker
  
## TODO  
- [ ] Add integraion with auth service using gRPC  
- [ ] Use kafka
- [ ] Make frontend
- [ ] Deploy fullstack app on server
