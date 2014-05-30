console.log('i\'m an happy worker, :)');

// libs
var gzipRequest = require('./lib/gzip-request.js');

// external deps
var cheerio = require('cheerio');

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
    case 'hour', 'hours':
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
  this.startTime = null;
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
      console.log('parsing torrent:', index);
      self.data[index] = rowData;
    });

    callback(null, self);

  });
};

function sendDataBack(err, data) {
  if (err) {
    process.send({
      err: err
    });
  } else {
    process.send({
      data: data
    });
  }
}

process.on('message', function (msg) {
  if (msg.url) {
    var singleFetch = new KatFetch();
    singleFetch.fetch(msg.url, sendDataBack);
  }
  console.log('dumping the message:', msg);
});
