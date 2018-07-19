const webpack = require('webpack');
const path = require('path');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const fs = require('fs');

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
      { from: '.web-extension-id', to: 'dist/' },
      { from: 'manifest.json', to: 'dist/',
        // Inject the current Browsh version into the manifest JSON
        transform(manifest, _) {
          const version_path = '../interfacer/src/browsh/version.go';
          let buffer = fs.readFileSync(version_path);
          let version_contents = buffer.toString();
          const matches = version_contents.match(/"(.*?)"/);
          return manifest.toString().replace('BROWSH_VERSION', matches[1]);
        }
      },
    ])
  ]
};
