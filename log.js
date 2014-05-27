'use strict';

var _ = require('lodash');
var baseLog = require('bunyan').createLogger({name: 'fantail'});

function createLogger(filename, extraObjects)
{

  //null needs to be tested in this way
  if (extraObjects == null) {
    extraObjects = {};
  }

  var extras = _.cloneDeep(extraObjects);
  extras.srcFile = filename;

  return baseLog.child(extras);
}

module.exports = createLogger;