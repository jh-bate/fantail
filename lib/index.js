'use strict';

var app = require('http').createServer(handler);
var config = require('../env');

var notifications = require('./notificationClient')(app);

var mongoClient = require('./mongoClient')(
  config.mongoDbConnectionString,
  notifications
);

var fs = require('fs');

app.listen(config.httpPort);

function handler(req, res) {
  fs.readFile(__dirname + '/../index.html', function(err, data) {
    if (err) {
      console.log(err);
      res.writeHead(500);
      return res.end('Error loading index.html');
    }
    res.writeHead(200);
    res.end(data);
  });
}

/*
 * Now lets poll!
 */
mongoClient.pollEventNotifications();