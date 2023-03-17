const path = require('path')
const HtmlWebpackPlugin = require('html-webpack-plugin')

const srcDir = path.join(__dirname, '..', 'src')
const distDir = path.join(__dirname, '..', 'dist')

module.exports = {
  mode: 'none',
  entry: path.join(srcDir, 'index.ts'),
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: 'ts-loader',
        include: srcDir,
      },
    ],
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.js'],
    symlinks: false,
    cacheWithContext: false,
  },
  plugins: [
    new HtmlWebpackPlugin({ // https://github.com/jantimon/html-webpack-plugin#options
      title: 'Dashboard',
      filename: 'index.html',
      // favicon: 'favicon.ico', // TODO: Add favicon
      meta: {
        robots: 'noindex, nofollow',
      },
    }),
  ],
  output: {
    path: distDir,
    filename: 'bundle.js',
    clean: true,
  },
}
