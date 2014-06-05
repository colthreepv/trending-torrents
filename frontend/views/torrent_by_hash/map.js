function (doc) {
  if (doc.type === 'torrent' || doc.hash) {
    emit(doc.hash, doc);
  }
}
