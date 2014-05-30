// node api
var cluster = require('cluster'),
    numCpu = require('os').cpus().length;

// libs
var gzipRequest = require('./lib/gzip-request.js');

// external deps
var cheerio = require('cheerio'),
    async = require('async'),
    extend = require('extend');

// setup the code to execute to worker.js
cluster.setupMaster({
  exec: 'worker.js'
});
var workersArr = [];
for (var i = 0; i < numCpu; i++) {
  workersArr.push(cluster.fork());
}

// listen in case of child death, debugging reasons only.. for now
workersArr.forEach(function (worker) {
  worker.on('exit', function (code, signal) {
    console.log('a child exited with code:', code, 'and signal', signal);
  });
});

function KatFetchCollection (howmany) {
  var board = [];
  for (var i = 0; i < howmany; i++) {
    board[i] = false;
  }

  this.activeFetchers = 0;
  this.current = 0;
  this.data = {};
  this.failures = {};
  this.board = board;
}

KatFetchCollection.prototype.getPage = function () {
  var freePage;
  if (freePage = this.board.indexOf(false), freePage !== -1) {
    this.board[freePage] = true;
    this.activeFetchers++;
    console.log('active fetchers:', this.activeFetchers);
    return freePage + 1;
  }
  return null;
};

// KatScout.
gzipRequest('http://kickass.to/new/', function (err, body) {
  if (err) {
    return console.log('KatScout failed:', err);
  }

  console.time('parse');
  var jq = cheerio.load(body);
  var pagesParsed = parseInt(jq('.turnoverButton.siteButton.bigButton:last-child').text(), 10);
  // console.log('page parsed, numPages:', pagesParsed);
  console.timeEnd('parse');


  // processing
  var workers = numCpu,
      fetchCollection = new KatFetchCollection(pagesParsed);

  // done, we start workers now.
  async.whilst(function testWhilst() {
    return workers > 0;
  }, function fn (callback) {
    var pageToFetch;
    // in case the pages to fetch are ended, we call callback immediately
    if (pageToFetch = fetchCollection.getPage(), !pageToFetch) {
      return callback();
    }


    workersArr[workers - 1].send({
      url: 'http://kickass.to/new/' + pageToFetch + '/'
    });
    // workersArr[workers - 1].once('message', function (msg) {
    //   console.log('received data back from worker:', msg);
    // });
  }, function (err) {

  });
});
