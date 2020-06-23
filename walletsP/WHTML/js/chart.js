var widgetOptions;
var widgetOptionsTwo;
var widgetOptionsThree;
// import Datafeed from './datafeed.js';

// window.tvWidget = new TradingView.widget({
// 	symbol: 'Bitfinex:BTC/USD', // default symbol
// 	interval: '1D', // default interval
// 	fullscreen: true, // displays the chart in the fullscreen mode
// 	container_id: 'chart-widget',
// 	datafeed: Datafeed,
// 	// library_path: '../charting_library_clonned_data/charting_library/',
// });

var Datafeed = {};

widgetOptions = {
  debug: false,
  datafeed: Datafeed, // our datafeed object
  interval: "1M",
  container_id: "chart-widget",
  // library_path: 'charting_library/charting_library/',
  locale: "en",
  style: "2",
  enable_publishing: false,
  hide_top_toolbar: true,
  hide_legend: true,
  save_image: false,
  timezone: "Europe/Athens",
  disabled_features: [
    "use_localstorage_for_settings",
    "left_toolbar",
    "border_around_the_chart",
    "remove_library_container_border",
    "control_bar",
    "create_volume_indicator_by_default",
    "volume_force_overlay"
  ],
  enabled_features: ["study_templates"],
  client_id: "test",
  user_id: "public_user_id",
  fullscreen: false,
  width: "100%",
  height: "100%",
  plotLineColorGrowing: "rgb(122, 194, 49)",
  //тут стили всего чего нужно.
  overrides: {
    "paneProperties.background": "#fff", //dark background
    "paneProperties.vertGridProperties.color": "transparent",
    "paneProperties.horzGridProperties.color": "#d8d8d8",
    "symbolWatermarkProperties.transparency": 0,
    "scalesProperties.textColor": "rgba(73, 133, 231, 1)"
    // "volumePaneSize": "tiny"
  }
};

widgetOptionsTwo = {
  debug: false,
  datafeed: Datafeed, // our datafeed object
  interval: "1M",
  container_id: "markets-dash",
  // library_path: 'charting_library/charting_library/',
  locale: "en",
  style: "2",
  enable_publishing: false,
  hide_top_toolbar: true,
  hide_legend: true,
  save_image: false,
  timezone: "Europe/Athens",
  disabled_features: [
    "use_localstorage_for_settings",
    "left_toolbar",
    "border_around_the_chart",
    "remove_library_container_border",
    "control_bar",
    "create_volume_indicator_by_default",
    "volume_force_overlay",
    "timezone_menu",
    "countdown",
    "timeframes_toolbar"
  ],
  // enabled_features: ["study_templates"],
  client_id: "test",
  user_id: "public_user_id",
  fullscreen: true,
  width: 100,
  height: 35,
  plotLineColorGrowing: "rgb(122, 194, 49)",
  //тут стили всего чего нужно.
  overrides: {
    "paneProperties.background": "#fff", //dark background
    "paneProperties.vertGridProperties.color": "transparent",
    "paneProperties.horzGridProperties.color": "#d8d8d8",
    "symbolWatermarkProperties.transparency": 0,
    "scalesProperties.textColor": "rgba(73, 133, 231, 1)"
    // "volumePaneSize": "tiny"
  }
};

widgetOptionsThree = {
  debug: false,
  datafeed: Datafeed, // our datafeed object
  interval: "1M",
  container_id: "markets-eth",
  // library_path: 'charting_library/charting_library/',
  locale: "en",
  style: "2",
  enable_publishing: false,
  hide_top_toolbar: true,
  hide_legend: true,
  save_image: false,
  timezone: "Europe/Athens",
  disabled_features: [
    "use_localstorage_for_settings",
    "left_toolbar",
    "border_around_the_chart",
    "remove_library_container_border",
    "control_bar",
    "create_volume_indicator_by_default",
    "volume_force_overlay",
    "timezone_menu",
    "countdown",
    "timeframes_toolbar"
  ],
  // enabled_features: ["study_templates"],
  client_id: "test",
  user_id: "public_user_id",
  fullscreen: false,
  width: 100,
  height: 35,
  plotLineColorGrowing: "rgb(122, 194, 49)",
  //тут стили всего чего нужно.
  overrides: {
    "paneProperties.background": "#fff", //dark background
    "paneProperties.vertGridProperties.color": "transparent",
    "paneProperties.horzGridProperties.color": "#d8d8d8",
    "symbolWatermarkProperties.transparency": 0,
    "scalesProperties.textColor": "rgba(73, 133, 231, 1)"
    // "volumePaneSize": "tiny"
  }
};

TradingView.onready(function() {
  var widget = (window.tvWidget = new TradingView.widget(widgetOptions));
  var widgetTwo = (window.tvWidget = new TradingView.widget(widgetOptionsTwo));
  var widgetThree = (window.tvWidget = new TradingView.widget(widgetOptionsThree));
});


// const main_chart = document.querySelector("#chart-widget");

// const chart = LightweightCharts.createChart(main_chart, {
//   localization: {
//     locale: "en-US"
//   },
//   priceScale: {
//     position: "right",
//     mode: 2,
//     autoScale: false,
//     invertScale: true,
//     alignLabels: false,
//     borderVisible: false,
//     borderColor: "#555ffd",
//     scaleMargins: {
//       top: 0.3,
//       bottom: 0.25
//     }
//   }
// });

// chart.resize("100%", "100%");

// const lineSeries = chart.addLineSeries();
// lineSeries.setData([
//   { time: "2019-04-11", value: 80.01 },
//   { time: "2019-04-12", value: 96.63 },
//   { time: "2019-04-13", value: 76.64 },
//   { time: "2019-04-14", value: 81.89 },
//   { time: "2019-04-15", value: 74.43 },
//   { time: "2019-04-16", value: 80.01 },
//   { time: "2019-04-17", value: 96.63 },
//   { time: "2019-04-18", value: 76.64 },
//   { time: "2019-04-19", value: 81.89 },
//   { time: "2019-04-20", value: 74.43 }
// ]);
