'use strict';

module.exports = function(app) {

  var io = require('socket.io').listen(app);
  var usedSocket;

  init();

  function init(){
    io.sockets.on('connection', function(socket) {
      usedSocket = socket;
    });
  }

  function getSocket(){
    return usedSocket;
  }

  return {
    send: function(content) {
      var socket = getSocket();
      socket.volatile.emit('notification', content);
    }
  };
};