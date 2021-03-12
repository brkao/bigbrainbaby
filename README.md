# bigbrainbaby (http://bigbrainbaby.herokuapp.com)
---------------------------
BigBrainBaby is a full stack application that is the user interface as well an API gateway.
It is the interface between the user and the data collected by babymachine (https://github.com/brkao/babymachine)

The application is self is a HTTP application that returns all the sentiment and post velocity
for the posts on reddit in table format as well a chart.  The computation of deltas of sentiment
and total count are done by bigbrainbaby using the raw data from Redis.

Additionally, a REST endpoint is also available for other applications that wishes to consume
the data collected for any other purposes.

This project includes deployment files to aid automatic deployment into Heroku Cloud.
