#!/usr/bin/env python3
"""
Usage: whisper_transcribe.py <model_size> <audio_file>
Prints transcript to stdout. model_size: tiny / base / small / medium / large-v3
"""
import sys, os

model_size = sys.argv[1]
audio_file = sys.argv[2]

if not os.path.exists(audio_file):
    print(f"file not found: {audio_file}", file=sys.stderr)
    sys.exit(1)

try:
    from faster_whisper import WhisperModel
    model = WhisperModel(model_size, device="cpu", compute_type="int8")
    segments, _ = model.transcribe(audio_file, beam_size=5)
    parts = [s.text.strip() for s in segments if s.text.strip()]
    print(" ".join(parts))
except Exception as e:
    print(f"whisper error: {e}", file=sys.stderr)
    sys.exit(1)
