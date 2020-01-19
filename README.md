A Slack bot to manage a queue of slack users in a channel.

## How?

This bot supports next commands:

* `add`   >   Add a user to the queue
* `del`   >   Delete user of the queue
* `show`  >   Show the queue 
* `clean` >   Delete all users in the queue 
* `pop`  >   Delete first user of the queue

## backlog
#### features
* direct notification on turn
* estimate on show
* require reason to add in the queue
* pass queue command
#### tech
* **cache file repository**
* **server test**
* i18n
* TTL to userInfoCache
* defensive limit msg per user


## Docs
https://api.slack.com/rtm
