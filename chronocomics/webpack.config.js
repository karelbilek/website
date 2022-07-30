const WebpackShellPlugin = require('webpack-shell-plugin');
const webpack = require('webpack');

const config = {
  entry: [
    './src/index.jsx'
  ],
  module: {
    loaders: [{
      test: /\.jsx?$/,
      exclude: /node_modules/,
      loader: 'babel-loader',
      query: {
        cacheDirectory: true,
      },
    }],

  },
  resolve: {
    extensions: ['.js', '.jsx']
  },
  output: {
    path: __dirname + '/dist',
    publicPath: '/',
    filename: 'bundle.js'
  },
  devServer: {
    contentBase: './dist',
  },
};

if (process.env.DEV_ENV === 'npmtest') {
    config.plugins.push(
        new WebpackShellPlugin({
            onBuildEnd: ['npm run test-true'],
            dev: false
        })
    )
}

module.exports = config;
