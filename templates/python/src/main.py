"""Automation handler: single entrypoint wrapped by stax-sdk decorator."""
import json
import sys

try:
    from stax_sdk import handler as stax_handler
except ImportError:
    def stax_handler(fn):
        """Stub when stax-sdk not installed: just call fn(event)."""
        return fn

@stax_handler
def handler(event):
    """User handler: receives CloudEvent (dict). Use event.get("data") for payload."""
    data = event.get("data") or {}
    print(json.dumps({"received": data}), file=sys.stderr)
    return {"ok": True}

if __name__ == "__main__":
    # Allow CLI to pass JSON CloudEvent via stdin
    payload = json.load(sys.stdin)
    result = handler(payload)
    print(json.dumps(result))
