ssh g.danichev@test.stl.msk.city-srv.ru <<'ENDSSH'
cd /home/g.danichev/gopath/src/slack-queue-bot && \
git pull && \
nohup /home/g.danichev/go/bin/go run main.go
ENDSSH
