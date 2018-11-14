# KarmaBoi
KarmaBoi was originally a python project for developing a cloud native python app.

[Visit the original defunct application if you like bad code!](https://github.com/tylarb/KarmaBoi-PCF)

It's now fully re-written in Golang, and features more features! Keep reading to see bad code in Golang!!1!


## Requirements and installation:

Add a [bot name and token](https://api.slack.com/bot-users) as environemtn variables `SLACK_BOT_NAME` and `SLACK_BOT_TOKEN`. 

KarmaBoi requires a Postgres Database. Currently it's hard coded to an [elephantsql](https://docs.run.pivotal.io/marketplace/services/elephantsql.html) instance. So make sure you have that. Or fix/improve it.


Once you have those two things, just push the app to Cloud Foundry:

~~~
cf push
~~~


## Basic Usage

You can give any user or name karma by adding '++' to the end of the word:

  ![alt text](https://github.com/tylarb/KarmaBoi-Go/blob/master/screenshots/karmaup.png "up")

Subtracting karma is just as simple - simply add '--':
  
  ![alt text](https://github.com/tylarb/KarmaBoi-Go/blob/master/screenshots/karmadown.png "down")

The bot uses user IDs, so if a user's display name changes, their karma will remain.

There's a timer to prevent vote spam - karma can't be added or subtracked during this time:

  ![alt text](https://github.com/tylarb/KarmaBoi-Go/blob/master/screenshots/timer.png "timer")

You can also give name shame. Be intentional - shame cannot be decreased, it stays for the life of the user!

  ![alt text](https://github.com/tylarb/KarmaBoi-Go/blob/master/screenshots/shame.png "shame")

You can see full leaderboards by messaging the bot and give it one of the following commands: rank (for highest karma leaderboard), !rank (for lowest karma leaderboard), or ~rank (for shame leaderboard):

  ![alt text](https://github.com/tylarb/KarmaBoi-Go/blob/master/screenshots/rank.png "rank")
  
The bot also has a memory feature - you can tag any word using the keyword "is also":

  ![alt text](https://github.com/tylarb/KarmaBoi-Go/blob/master/screenshots/memoryset.png "memory set")

and display what is remembered with the "keyword + ?". If a keyword has multiple inputs, the bot will choose a random one to display:

  ![alt text](https://github.com/tylarb/KarmaBoi-Go/blob/master/screenshots/memoryask.png "memory ask")


### Additional information

Submit an issue or for any questions. I welcome contributions via pull requests as well.



### License

Released under MIT license, copyright 2018 Tyler Ramer

