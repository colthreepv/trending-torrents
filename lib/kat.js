// node api
var events = require('events'),
    util = require('util');

// libs
var gzipRequest = require('./gzip-request.js');

// external deps
var cheerio = require('cheerio');

/**
 * Code used by master
 */
function KatFetchCollection (howmany) {
  var board = [];
  for (var i = 0; i < howmany; i++) {
    board[i] = false;
  }

  this.activeFetchers = 0;
  this.current = 0;
  this.data = [];
  this.failures = {};
  this.board = board;
}
util.inherits(KatFetchCollection, events.EventEmitter);

KatFetchCollection.prototype.getPage = function () {
  var freePage;
  if (freePage = this.board.indexOf(false), freePage !== -1) {
    this.board[freePage] = true;
    this.activeFetchers++;
    if (this.activeFetchers < 8) {
      console.log('active fetchers:', this.activeFetchers);
    }
    return freePage + 1;
  }
  return null;
};
KatFetchCollection.prototype.success = function (data) {
  this.data.push(data);
  this.activeFetchers--;
  if (this.activeFetchers < 7) {
    console.log('active fetchers:', this.activeFetchers);
  }
  if (this.activeFetchers === 0) {
    this.emit('done');
  }
};

/**
 * Code used by workers
 */
function parseSize (size) {
  var splitSize = size.split(' ');
  var numericSize = parseFloat(splitSize[0]);

  switch (splitSize[1]) {
    case 'bytes':
      return numericSize;
    case 'KB':
      return numericSize * 1024;
    case 'MB':
      return numericSize * 1024 * 1024;
    case 'GB':
      return numericSize * 1024 * 1024 * 1024;
    default:
      console.log('size not handled:', splitSize[1]);
      return numericSize;
  }
}

function parseAge (timeAgo) {
  var splitTime = timeAgo.split(/\s/); // WARNING: ' ' didn't work as splitter, &nbsp in there!
  var numericTime = parseInt(splitTime[0], 10);
  var timeNow = Date.now();

  switch (splitTime[1]) {
    case 'sec.':
      timeNow = timeNow - (numericTime * 1000);
      break;
    case 'min.':
      timeNow = timeNow - (numericTime * 60 * 1000);
      break;
    case 'hour':
    case 'hours':
      timeNow = timeNow - (numericTime * 60 * 60 * 1000);
      break;
    case 'day':
      timeNow = timeNow - (numericTime * 24 * 60 * 60 * 1000);
      break;
    case 'week':
      timeNow = timeNow - (numericTime * 7 * 24 * 60 * 60 * 1000);
      break;
    default:
      console.log('timeAgo not handled:', splitTime[1]);
      return new Date();
  }

  return new Date(timeNow);
}

function KatFetch() {
  this.startTime = Date.now();
  this.elapsed  = null;
  this.data = [];
}

// callback must comply sendDataBack (err, data)
KatFetch.prototype.fetch = function (url, callback) {
  var self = this;
  gzipRequest(url, function (err, body) {
    if (err) {
      return callback(err);
    }

    var jq = cheerio.load(body);
    var torrentElements = jq('table .odd, table .even');
    torrentElements.each(function (index, torrent) {
      torrent = cheerio(torrent);

      var rowData = {
        name: torrent.find('.cellMainLink').text(),
        magnet: torrent.find('.imagnet').attr('href'),
        size: parseSize(torrent.find('.nobr.center').text()),
        files: parseInt(torrent.find('.nobr.center').next().text(), 10),
        age: parseAge(torrent.find('.nobr.center').next().next().text())
      };
      // console.log('parsing torrent:', index);
      self.data[index] = rowData;
    });

    // write elapsed time
    self.elapsed = Date.now() - self.startTime;
    callback(null, self);
  });
};


module.exports = {
  KatFetchCollection: KatFetchCollection,
  KatFetch: KatFetch,
  parseAge: parseAge,
  parseSize: parseSize
};