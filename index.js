// node api
var cluster = require('cluster'),
    numCpu = require('os').cpus().length,
    fs = require('fs');

// libs
var gzipRequest = require('./lib/gzip-request.js'),
    kat = require('./lib/kat');

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
  cluster.fork();
}

// listen in case of child death, debugging reasons only.. for now
Array.prototype.forEach.call(cluster.workers, function (worker) {
  worker.on('exit', function (code, signal) {
    console.log('a child exited with code:', code, 'and signal', signal);
  });
});

// KatScout with retry
async.retry(function (callback, results) {
  gzipRequest('http://kickass.to/new/', callback);
}, function scoutDone (err, body) {
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
      fetchCollection = new kat.KatFetchCollection(20);

  // done, we start workers now.
  async.whilst(function testWhilst() { // OLOLOL, i don't have to use whilst, but .queue :(
    return workers > 0;
  }, function fn (callback) {
    var pageToFetch;
    // in case the pages to fetch are ended, we call callback immediately
    if (pageToFetch = fetchCollection.getPage(), !pageToFetch) {
      return callback();
    }

    // dispatch work to a worker
    cluster.workers[workers - 1].send({
      url: 'http://kickass.to/new/' + pageToFetch + '/'
    });

    // and listen once for completion
    cluster.workers[workers - 1].once('message', function (msg) {
      if (msg.data) {
        fetchCollection.success(msg.data);
        return callback();
      } else {
        return callback('no data received for page:' + pageToFetch);
      }
    });
  }, function (err) {
    if (err) {
      return console.log(err);
    }

    console.log('fetching done!');
    fs.writeFileSync('node-data.js', JSON.stringify(fetchCollection));
  });
});
