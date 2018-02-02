// `npm install -g jsonwebtoken`
var jwt = require('jsonwebtoken');

var key = 'user:13243312:78';
var secret = process.env.MDN_KEY;

var issuedAt = Math.floor(Date.now() / 1000);
var payload = {
  iss: key,
  jti: Math.random().toString(),
  iat: issuedAt,
  exp: issuedAt + 60,
};

var token = jwt.sign(payload, secret, {
  algorithm: 'HS256',  // HMAC-SHA256 signing algorithm
});

var auth = 'JWT ' + token;
var path = '848208/browsh-0.2.3-an+fx.xpi';
var base = 'https://addons.mozilla.org/api/v3/file/';
var uri = base + path;

process.stdout.write('curl -H "Authorization: ' + auth + '" ' + uri);
