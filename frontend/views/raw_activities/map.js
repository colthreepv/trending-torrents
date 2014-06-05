function (doc) {
  var magnetRegex = /btih:([0-9A-F]*)/;
  if (doc.startTime && !doc.done) {
    doc.data.forEach(function (torrent, index) {
      var matched;
      if (matched = torrent.magnet.match(magnetRegex), matched) {
        emit(matched[1], torrent);
      }
    });
  }
}
