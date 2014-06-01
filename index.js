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
  workersArr[i] = cluster.fork();
}

// listen in case of child death, debugging reasons only.. for now
workersArr.forEach(function (worker) {
  worker.on('exit', function (code, signal) {
    console.log('a child exited with code:', code, 'and signal', signal);
  });
});

// KatScout with retry
async.retry(function (callback, results) {
  gzipRequest('http://kat.col3.me/new/', callback);
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
  var workersWorking = [],
      f = new kat.KatFetchCollection(100);

  // init the workers
  for (var i = 0; i < workersArr.length; i++) {
    workersWorking[i] = false;
  }

  // done, prepare queue executor
  var workersQ = async.queue(function worker (task, callback) {
    // task has 2 keys: page, workerId
    if (task.workerId === -1) {
      console.log('this is gonna crash', workersWorking);
    }

    // dispatch work to a worker
    workersArr[task.workerId].send({
      url: 'http://kat.col3.me/new/' + task.page + '/',
      page: task.page,
      worker: task.workerId
    });

    // and listen once for completion
    workersArr[task.workerId].once('message', function (msg) {
      if (msg.data) {
        f.success(msg.data);
        return callback();
      } else {
        return callback('no data received for page: ' + task.page);
      }
    });
  }, workersArr.length);

  function createWorkerCallback() {
    var selectedWorker;
    // finds the first idling worker if it's not the first time the fn gets called
    if (selectedWorker = workersWorking.indexOf(false), selectedWorker == -1) {
      console.log('bad! i have called createWorkerClosure without an idle worker!!!');
      return process.exit();
    }
    workersWorking[selectedWorker] = true;
    return function workDone (err) {
      if (err) {
        console.log(
          'some error happened processing page',
          pageToFetch,
          'i will requeue it, i am worker number',
          selectedWorker
        );
      }
      workersWorking[selectedWorker] = false;

      var pageToFetch;
      if (pageToFetch = f.getPage(), pageToFetch) {
        workersQ.push({
          page: pageToFetch,
          workerId: selectedWorker
        }, createWorkerCallback());
      }
    };
  }

  // bootstrap workers
  for (i = 0; i < workersArr.length; i++) {
    workersQ.push({
      page: f.getPage(),
      workerId: i
    }, createWorkerCallback());
  }

  f.on('done', function () {
    console.log('active processes reached 0, we should have done it');
    fs.writeFileSync('katfetch1.json', JSON.stringify(this.data));
    process.exit();
  });
});
