const webpack = require('webpack');
const path = require('path');

module.exports = {
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
  plugins: [
    new webpack.DefinePlugin({
      // TODO: For production use a different webpack.config.js
      DEVELOPMENT: JSON.stringify(true),
      PRODUCTION: JSON.stringify(false)
    })
  ]
};
