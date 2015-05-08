var add = require('./index');

var assert = require('assert');

var expected = add(1,2);
assert( expected === 3, 'one plus two is three');