module.exports = function(grunt) {
  'use strict';

  // Project configuration.
  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),
    jshint: {
      options: {
        jshintrc: '.jshintrc'
      },
      all: ['Gruntfile.js', 'lib/**/*.js', 'test/**/*.js']
    },
    shell: {
      startMongo: {
        command: [
          'mongod',
          'mongo'
        ].join('&&'),
        options: {
          async: false,
          failOnError: false
        }
      },
      startAPI: {
        command: [
          'node lib/index.js'
        ]
      }
    }
  });

  // Load the plugins
  grunt.loadNpmTasks('grunt-contrib-jshint');
  grunt.loadNpmTasks('grunt-shell-spawn');

  // Default task(s).
  grunt.registerTask('default', ['test']);
  // Standard tasks
  grunt.registerTask('test', ['jshint']);
  grunt.registerTask('start-mongo', ['shell:startMongo']);
  grunt.registerTask('start', ['test','start-mongo','shell:startAPI']);

};
