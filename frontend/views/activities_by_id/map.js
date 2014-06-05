function (doc) {
  if (doc.startTime && !doc.done) {
    emit(doc._id, doc);
  }
}
