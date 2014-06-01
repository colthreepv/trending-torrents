
console.log('i\'m an happy worker, :)');

// libs
var kat = require('./lib/kat');

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
    var singleFetch = new kat.KatFetch();
    singleFetch.fetch(msg.url, sendDataBack);
  }
  console.log(
    'worker:',
    msg.worker,
    'fetching page:',
    msg.page
  );
});
