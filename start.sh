go test -v "github.com/yonesko/slack-queue-bot/..." && \
go build -o run

if [ $? -eq 0 ]; then
  echo "Starting..."
  supervise . &
else
  echo "Failed to test and build"
fi

tail -f slack-queue-bot.log
