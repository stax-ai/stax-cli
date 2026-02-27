"""Tests for the automation handler."""
import sys
import os

# Project root on path for main
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

def test_import_handler():
    from templates.python.src.main import handler
    assert callable(handler)
