A Slack bot to manage a queue of slack users in a channel.

## How?

Turnio expose a API with main functionalities of a queue. You only need to mention him with some of next commands:

* `add`   >   Add a user to the queue
* `del`   >   Delete user of the queue
* `show`  >   Show the queue 
* `clean` >   Delete all users in the queue 
* `pop`  >   Delete first user of the queue

##todo
* mention
* logs round robin
* refac : logic in service only
* notifications (your turn)
* estimate (dry add)


##Docs
https://api.slack.com/rtm