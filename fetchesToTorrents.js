var request = require('request');

request.get('http://localhost:5984/trendingtorrents/_design/convert/_view/getTorrentsFromFetches', function (err, resp, body) {
  if (err || resp.statusCode !== 200) {
    return console.log(err, resp.statusCode);
  }

  var bigData = JSON.parse(body);
  console.log('bigData length', bigData);
});