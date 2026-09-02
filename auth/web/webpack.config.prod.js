const { CleanWebpackPlugin } = require('clean-webpack-plugin')
const webpack = require('webpack');
const path = require('path')
module.exports = {
    mode: 'development',
    plugins: [
        new CleanWebpackPlugin({
            verbose: true,
        }),
        new webpack.DefinePlugin({
            'process.env': {
                BACKEND: JSON.stringify("stk.vrandkode.net:8081"),
                FRONTEND: JSON.stringify("stk.vrandkode.net:9999"), 
                ENCRYPTED: true
            }
          })
    ],
    entry: {
        main: './src/web/frontend/main.tsx',
    },

    output: {
        filename: '[name].bundle.js',
        chunkFilename: '[name].chunk.js',
        path: __dirname + '/dist/web/frontend',
        publicPath: '/assets/'
    },

    devtool: 'source-map',

    resolve: {
        extensions: ['.ts', '.tsx', '.js']
    },

    module: {
        rules: [
            {
                test: /\.tsx?$/,
                loader: 'ts-loader',
            },

            {enforce: 'pre', test: /\.js$/, loader: 'source-map-loader'},
            {
                test: /\.css$/,
                use: [{loader: 'style-loader'}, {loader: 'css-loader'}]
            },
        ]
    },

    optimization: {
        splitChunks: {
            chunks: 'all'
        },
        usedExports: true
    }
}