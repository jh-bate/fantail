'use strict';

var fs = require('fs');

module.exports = (function(){
  var env = {};

  // The port to attach an HTTP listener, if null, no HTTP listener will be attached
  env.httpPort = process.env.PORT || 8000;

  env.mongoDbConnectionString = process.env.MONGO_CONNECTION_STRING || 'mongodb://localhost/notifications';

  return env;
})();
