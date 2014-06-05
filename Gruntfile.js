module.exports = function (grunt) {
  grunt.loadNpmTasks('grunt-couch');

  grunt.initConfig({
    'couch-compile': {
      app: {
        files: {
          'tmp/simple.json': 'frontend'
        }
      }
    },
    'couch-push': {
      localhost: {
        files: {
          'http://localhost:5984/trendingtorrents': 'tmp/simple.json'
        }
      }
    },
    backend: {
      url: 'http://localhost:5984/',
      database: 'trendingtorrents',
      designDoc: 'frontend'
    }
  });

  grunt.registerTask('backend', 'gets the list of fetches, converts it in torrent elements', function () {
    var request = require('request'),
        async = require('async');
    var config = grunt.config('backend');

    var done = this.async();

    /**
     * 1 - get raw activities
     * 2 - understand which torrents are present already
     *   a - query torrent_by_hash
     *   b - remove 
     *   c1 - update those with more informations
     *   c2 - add the others with an INSERT
     * 5 - update activities, marking them as done
     */
    async.auto({
      // 1
      getActivities: function (callback) {
        request({
          url: config.url + config.database + '/_design/' + config.designDoc + '/_view/' + 'raw_activities',
          json: true
        }, function (err, resp, body) {
          if (err) callback(err);
          // body contains A LOT of torrent hashes
          grunt.log.oklns('read ' + body.total_rows + ' torrents, checking their presence');
          callback(null, body);
        });
      },
      // 2
      checkTorrentPresence: ['getActivities', function (callback, data) {
        var torrentHashes = data.getActivities.rows.map(function (torrent, index) {
          return torrent.key;
        });

        request({
          url: config.url + config.database + '/_design/' + config.designDoc + '/_view/' + 'torrent_by_hash',
          method: 'POST',
          json: {
            keys: torrentHashes
          }
        }, function (err, resp, torrentsPresent) {
          if (err) callback(err);

          grunt.log.oklns('of the above ' + torrentsPresent.rows.length + ' are in the need of an UPDATE');
          callback(null, { allHashes: torrentHashes, onlyUpdate: torrentsPresent });
        });
      }],
      torrentBulkDocs: ['getActivities', 'checkTorrentPresence', function (callback, data) {
        // var torrentsToUpdate = data.checkTorrentPresence.onlyUpdate

        // the first one is the present torrent, the second "the new occurrence"
        console.log(data.checkTorrentPresence.onlyUpdate.rows[2]);
        console.log(data.getActivities.rows.filter(function (torrent) {
          if (torrent.key === data.checkTorrentPresence.onlyUpdate.rows[2].key) return true;
        }));
      }]
    }, function (err, results) {
      if (err) {
        grunt.log.error(err);
      }
    });

    /**
    // function in step 4
    function doneActivity(activities, callback) {
      var activitiesToUpdate = activities.rows.map(function (fetchIdx, index) {
        return fetchIdx.id;
      }).filter(function onlyUnique(value, index, self) {
        return self.indexOf(value) === index;
      });

      request({
        url: config.url + config.database + '/_design/' + config.designDoc + '/_view/' + 'activities_by_id',
        method: 'POST',
        json: {
          keys: activitiesToUpdate
        }
      }, function (err, resp, fetchesSelected) {
        if (err) { grunt.log.error(err); }

        var updateFetchObjects = fetchesSelected.rows.map(function (fetchObj, index) {
          fetchObj.value.done = true;
          return fetchObj.value;
        });

        request({
          url: config.url + config.database + '/_bulk_docs',
          method: 'POST',
          json: {
            docs: updateFetchObjects
          }
        }, function (err, resp, updatedActivities) {
          if (err) { grunt.log.error(err); }

          grunt.log.oklns('Closed', updatedActivities.length, 'Fetch objects');
          callback();
        });
      });
    }

    // 1 - get raw activities
    request({
      url: config.url + config.database + '/_design/' + config.designDoc + '/_view/' + 'raw_activities',
      json: true
    }, function (err, resp, fetches) {
      if (err) { grunt.log.error(err); }

      // fetches contains A LOT of torrent hashes
      grunt.log.oklns('read ' + fetches.total_rows + ' torrents, checking their presence');

      // 2 - understand which torrents are present already
      var torrentHashes = fetches.rows.map(function (torrent, index) { return torrent.id; });
      request({
        url: config.url + config.database + '/_design/' + config.designDoc + '/_view/' + 'torrent_by_hash',
        method: 'POST',
        json: {
          keys: torrentHashes
        }
      }, function (err, resp, torrentsPresent) {
        if (err) { grunt.log.error(err); }
        grunt.log.oklns('of the above ' + torrentsPresent.total_rows + ' are in the need of an UPDATE');

        // 3 - update those with more informations
        // function () {}

        var torrentsToAdd = fetches.rows.map(function (torrent, index) {
          return {
            hash: torrent.key,
            name: torrent.value.name,
            magnet: torrent.value.magnet,
            size: torrent.value.size,
            files: torrent.value.files,
            age: torrent.value.age
          };
        });

        // 4 - add the others with an INSERT
        request({
          url: config.url + config.database + '/_bulk_docs',
          method: 'POST',
          json: {
            docs: torrentsToAdd
          }
        }, function (err, resp, inserted) {
          if (err) { grunt.log.error(err); }

          grunt.log.ok('inserted', inserted.length, ' torrents.');

          // 5 - update activities, marking them as done
          doneActivity(fetches, done);
        });

      });

    });
    **/
  });

  grunt.registerTask('default', ['couch-compile', 'couch-push']);
};