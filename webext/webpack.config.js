const webpack = require('webpack');
const path = require('path');
const CopyWebpackPlugin = require('copy-webpack-plugin');

module.exports = {
  mode: process.env['BROWSH_ENV'] === 'RELEASE' ? 'production' : 'development',
  target: 'node',
  entry: {
    content: './content.js',
    background: './background.js'
  },
  output: {
    path: __dirname,
    filename: 'dist/[name].js',
  },
  resolve: {
    modules: [
      path.resolve(__dirname, './src'),
      'node_modules'
    ]
  },
  devtool: 'source-map',
  plugins: [
    new webpack.DefinePlugin({
      DEVELOPMENT: JSON.stringify(true),
      TEST: JSON.stringify(false),
      // TODO: For production use a different webpack.config.js
      PRODUCTION: JSON.stringify(false)
    }),
    new CopyWebpackPlugin([
      { from: 'assets', to: 'dist/assets' },
      { from: 'manifest.json', to: 'dist/' },
      { from: '.web-extension-id', to: 'dist/' },
    ])
  ]
};
