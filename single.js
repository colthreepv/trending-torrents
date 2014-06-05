// node api
var fs = require('fs');

// libs
var gzipRequest = require('./lib/gzip-request.js'),
    kat = require('./lib/kat');

// external deps
var cheerio = require('cheerio'),
    request = require('request'),
    async = require('async'),
    extend = require('extend');

console.time('All done in');
var concurrency = 8; // 8 HTTP requests simultaneously
var socketsPool = { maxSockets: concurrency };
var couchDB = {
  url: 'http://localhost:5984/',
  db: 'trendingtorrents'
};

// KatScout with retry
async.retry(function (callback, results) {
  gzipRequest('http://kat.col3.me/new/', callback);
}, function scoutDone (err, body) {
  if (err) {
    return console.log('KatScout failed:', err);
  }

  var jq = cheerio.load(body);
  var pagesParsed = parseInt(jq('.turnoverButton.siteButton.bigButton:last-child').text(), 10);
  // pagesParsed = 100;
  var f = new kat.KatFetchCollection(pagesParsed);

  // task has 1 key: page
  var pagesQ = async.queue(function worker (task, callback) {
    // single thread logic
    var singleFetch = new kat.KatFetch();
    singleFetch.fetch({
      url: 'http://kickass.to/new/' + task.page + '/',
      pool: socketsPool
    }, function (err, data) {
      if (err) {
        return callback(err);
      }

      f.success(data);
      callback();
    });
  }, concurrency);

  function workDone (err) {
    if (err) {
      console.log(
        'some error happened processing page',
        pageToFetch
      );
    }

    var pageToFetch;
    if (pageToFetch = f.getPage(), pageToFetch) {
      pagesQ.push({
        page: pageToFetch,
      }, workDone);
    }
  }

  // bootstrap workers
  // NOTE: concurrency MUST be lesser than pageNumber, *ONLY* for the bootstrap
  for (var i = 0, bootstrapNum = Math.min(concurrency, pagesParsed); i < bootstrapNum; i++) {
    pagesQ.push({
      page: f.getPage()
    }, workDone);
  }

  f.on('done', function () {
    console.log('all is done!');
    request.post({
      url: couchDB.url + couchDB.db + '/' + '_bulk_docs',
      json: {
        docs: this.data
      }
    }, function (err, resp, body) {
      if (err) {
        throw err;
      }

      console.log('couchDB response:', resp.statusCode);
      console.timeEnd('All done in');
      process.exit();
    });
  });

});