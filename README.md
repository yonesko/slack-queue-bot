A Slack bot to manage a queue of slack users in a channel.

## How?

This bot supports next commands:

* `add`   >   Add a user to the queue
* `del`   >   Delete user of the queue
* `show`  >   Show the queue 
* `clean` >   Delete all users in the queue 
* `pop`  >   Delete first user of the queue

## todo
* logs round robin
* direct notifications (your turn)
* estimate (dry add)
* require reason to add in the queue
* TTL to userInfoCache
* pass queue command
* i18n
* mutex to repo
* graceful shutdown
* run handler in goroutine and catch panic
* answer to the thread
* defensive limit msg per user


## Docs
https://api.slack.com/rtm
