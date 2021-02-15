var randomColorGenerator = function () {
    return '#' + (Math.random().toString(16) + '0000000').slice(2, 8);
};

var allCharts = {};

function drawSentimentChart(chartName, dataTableName, title) {
    var table = document.getElementById(dataTableName);
    var json = []; // First row needs to be headers
    var headers =[];
    for (var i = 0; i < table.rows[0].cells.length; i++) {
        headers[i] = table.rows[0].cells[i].innerHTML.toLowerCase().replace(/ /gi, '');
    }

    // Go through cells
    for (var i = 1; i < table.rows.length; i++) {
        var tableRow = table.rows[i];
        var rowData = {};
        for (var j = 0; j < tableRow.cells.length; j++) {
            rowData[headers[j]] = tableRow.cells[j].innerHTML;
        }

        json.push(rowData);
    }

    console.log(json);
    // Map JSON values back to label array
    var labels = json.map(function (e) {
        return e.symbol;
    });
    console.log(labels); // ["2016", "2017", "2018", "2019"]

    // Map JSON values back to values array
    var values = json.map(function (e) {
        return e.compound;
    });
    console.log(values); // ["10", "25", "55", "120"]
    var chart = BuildChart(chartName, labels, values, title);
    allCharts[chartName] = chart;
}

function BuildChart(chartName, labels, values, chartTitle) {
  if (allCharts[chartName] != null) {
      allCharts[chartName].destroy();
  }
  var ctx = document.getElementById(chartName).getContext('2d');
  var colors = new Array();

  values.forEach(function (item, index) {
    colors.push(randomColorGenerator());
  })
  console.log(labels);
  var myChart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labels, // Our labels
      datasets: [{
        label: chartTitle, // Name the series
        data: values, // Our values
        backgroundColor: colors,
      }]
    },
    options: {
      responsive: true, // Instruct chart js to respond nicely.
      maintainAspectRatio: false, // Add to prevent default behavior of full-width/height
    }
  });
  return myChart;
}
