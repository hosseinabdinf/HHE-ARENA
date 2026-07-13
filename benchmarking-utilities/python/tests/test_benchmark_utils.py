import gc
import unittest

from utils.benchmark_utils import (
    _memory_delta,
    _proportional_warmups,
    _set_gc,
    _signed_delta,
)


class BenchmarkUtilityTests(unittest.TestCase):
    def test_proportional_warmups(self):
        self.assertEqual(_proportional_warmups(1), 1)
        self.assertEqual(_proportional_warmups(100), 10)
        self.assertEqual(_proportional_warmups(1000), 10)

    def test_signed_delta_is_not_clamped(self):
        self.assertEqual(_signed_delta(10, 3), -7)

    def test_memory_delta_retains_absolute_ru_maxrss(self):
        memory = _memory_delta((100, 200, 300), (110, 220, 350))
        self.assertEqual(memory["peak_rss_bytes_delta"], 50)
        self.assertEqual(memory["ru_maxrss_bytes"], 350)
    def test_gc_modes_are_explicit(self):
        was_enabled = gc.isenabled()
        try:
            _set_gc("disabled")
            self.assertFalse(gc.isenabled())
            _set_gc("enabled")
            self.assertTrue(gc.isenabled())
        finally:
            gc.enable() if was_enabled else gc.disable()


if __name__ == "__main__":
    unittest.main()
