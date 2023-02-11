/* global module require */

const { merge } = require('webpack-merge')
const common = require('./webpack.common.js')

module.exports = merge(common, {
  devtool: 'inline-source-map',
  mode: 'development',
  optimization: {
    minimize: false,
  },
  devServer: {
    host: '0.0.0.0',
    port: 8080,
    allowedHosts: ['all'],
    open: false,
    liveReload: true,

    client: {
      webSocketURL: 'auto://0.0.0.0:0/ws', // note the `:0` after `0.0.0.0`
    },
  },
})
