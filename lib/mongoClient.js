'use strict';

var mongojs = require('mongojs');
var _ = require('lodash');
var log = require('../log')('mongoClient.js');

var notificationsCollectionName = 'tlNotifications';

/*
 * CRUD opertaions via Mongo instance
 */
module.exports = function(connectionString, emitter) {

  var dependencyStatus = { running: false, deps: { up: [], down: [] } };

  var dbInstance = mongojs(connectionString, [notificationsCollectionName],function(err){
    log.error('error opening mongo');
    dependencyStatus = isDown(dependencyStatus);
  });

  dependencyStatus = isUp(dependencyStatus);

  dbInstance.on('error',function (err) {
    log.error('error with mongo connection',err);
    dependencyStatus = isDown(dependencyStatus);
  });

  dbInstance.on('disconnected', function () {
    log.warn('we lost the mongo connection');
    dependencyStatus = isDown(dependencyStatus);
  });

  var notificationsCollection = dbInstance.collection(notificationsCollectionName);

  /*
    Mongo is down
  */
  function isDown(status){
    status.deps.up = _.without(status.deps.up, 'mongo');
    status.deps.down = _.union(status.deps.down, ['mongo']);
    return status;
  }

  /*
    Mongo is up
  */
  function isUp(status){
    status.deps.down = _.without(status.deps.down, 'mongo');
    status.deps.up = _.union(status.deps.up, ['mongo']);
    return status;
  }

  function emitNotifictaions(collection, doc, cb){
    //  using tailable cursor get reference to our very first doc
    var query = { _id: { $gt: doc._id } };
    var options = { tailable: true, awaitdata: true, numberOfRetries: -1 };
    var cursor = collection.find(query, options).sort({ $natural: 1 });
    // This function will take cursor to next doc from current as soon as 'notifications' database is updated
    function next() {
      cursor.next(function(err, message) {
        if (err) { return cb(err); }
        emitter.send(message);
        next();
      });
    }
    // what you need to do is: call it first time
    next();
  }

  return {
    status: function(callback) {
      log.debug('checking status');
      return callback(null, dependencyStatus);
    },
    pollEventNotifications: function() {
      notificationsCollection
        .find()
        .sort({ $natural: -1 })
        .limit(1)
        .next(function(err, doc) {
          if(_.isEmpty(err) && !_.isEmpty(doc) ) {
            emitNotifictaions(notificationsCollection, doc);
          } else {
            if( err) {
              console.error('error polling notifictions',err);
            } else {
              console.log('no notifications yet');
            }
          }
        });
    }
  };
};