A Slack bot to manage a queue of slack users in a channel written in Golang.

This bot is used in production at https://city-mobil.ru/

## How?

This bot supports next commands:

* `add`   >   Add a user to the queue
* `del`   >   Delete user of the queue
* `show`  >   Show the queue 
* `clean` >   Delete all users in the queue 
* `pop`  >   Delete first user of the queue

## backlog
#### features
* direct notification when you've been deleted
* estimate on show
* require reason to add in the queue
* pass queue command
* local command names
#### tech
* add logger with levels to grep errors
* add general fileRepository
* rename slack-queue-bot.db.json
* defensive limit msg per user
* log HTTP Response body


## Docs
https://api.slack.com/rtm
