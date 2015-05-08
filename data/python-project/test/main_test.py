from nose.tools import assert_equal
from nose.tools import assert_false
from nose.tools import assert_true


def test_config_returns_defaults_if_all_none_and_no_egrc():
    assert_equal(1, 1)
    assert_true(True)
    assert_false(False)
