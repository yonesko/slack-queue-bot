go test -v "github.com/yonesko/slack-queue-bot/..." && \
go build && \
nohup ./slack-queue-bot&

tail -f slack-queue-bot.log