// Code goes here


// functions to generate an object comprised of COMPLETELY FAKE DATAZ!
(function (exports) {

  // shortcode
  var mf = Math.floor;
  var mr = function (multiplier) { return mf(Math.random() * multiplier + 1); };
  var dd = function (jd) { return jd.getDate() + '/' + (jd.getMonth() + 1) + '/' + jd.getFullYear(); };

  var randomDate = function (day, month, year, avgLeech, avgSeed) {
    // distance from average 30%
    var distanceFromAvg = 0.3;

    // spec
    // { date: '14/02/2014 16:23', leechers: 58, seeders: 25 }
    return { date: dd(new Date(year, month, day, mr(23), mr(59), mr(59))), leechers: mr(300), seeders: mr(100) };
  };

  var randomMonth = function (monthIndex, days, year, maxDates) {
    var availableDays = [];
    for (var i = 1; i <= days; i++) {
      availableDays.push(i);
    }
    // availableDays = [1, 2, 3, 4, 5, 6, 7, 8.....];

    var probability = 100;
    var datesArray = [];
    var randomDay;
    for (i = 0; i < maxDates; i++) {

      // if the random number output surpasses the probability, generate a date
      if (mr(100) < probability) {
        probability = probability * 0.65;

        // take out a random day
        randomDay = availableDays[mr(availableDays.length - 1)];
        // generate a random date for THIS day, and push it
        datesArray.push(randomDate(randomDay, monthIndex, year));

        // remove that day from the available
        availableDays.splice(availableDays.indexOf(randomDay), 1);
      }
    }
    return datesArray;
  };

  var randomYear = function (year, flat) {
    // returns an object like this:
    // { jan: [{ date:  }], feb: [], mar: [].... }
    var months = ['jan', 'feb', 'mar', 'apr', 'may', 'jun', 'jul', 'aug', 'sep', 'oct', 'nov', 'dec'];
    var days = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
    var returnObj = {};
    var returnArray = [];

    months.forEach(function (month, index, array) {
      if (flat) {
        returnArray = returnArray.concat(randomMonth(index, days[index], year, 10));
      } else {
        returnObj[months[index]] = randomMonth(index, days[index], year, 10);
      }
    });
    return (flat) ? returnArray : returnObj;
  };

  /** jQuery starts here!
  $(function () {
    var randomGenerated = randomYear(2013, true);
    console.log(randomGenerated);

    $('#tabledata').handsontable({
      data: randomGenerated,
      colHeaders: ['Snapshot Date', 'Leechers', 'Seeders'],
      columns: [
        {data: 'date', type: 'date', dateFormat: 'dd/mm/yy'},
        {data: 'leechers' },
        {data: 'seeders' }
      ],
      contextMenu: true
    });

  });
  **/

  exports.random = {
    randomDate: randomDate,
    randomMonth: randomMonth,
    randomYear: randomYear
  };
})(window);
