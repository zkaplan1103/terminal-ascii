#!/usr/bin/env python3
"""
slice.py — Phase 1 of the build-frames pipeline.

Reads RPGMaker-format sprite sheets from assets/animals/farm_animals_4.18.24/
and slices every defined animal×state into individual frame PNGs under
/tmp/build-frames/<species>_<variant>/<state>/frame_NNN.png

Sheet layout (all sheets):
  384×256 px total
  4 character blocks per row, 2 rows = 8 characters per sheet
  Each character block: 96×128 px
  Walk frames: 3 columns × 4 direction rows = 32×32 px per frame
  Direction rows: 0=down, 1=left, 2=right, 3=up

We use direction row 2 (facing right / side profile) for all animations
since that reads best in a left-aligned terminal.

For eating states the pack provides separate "head down grazing" character
blocks on the same sheet (slots 4-7 mirror slots 0-3 but with head down).
"""

from PIL import Image
import os, shutil

ASSETS = os.path.join(os.path.dirname(__file__), '../../assets/animals/farm_animals_4.18.24')
OUT    = '/tmp/build-frames'

# RPGMaker sheet constants
CHAR_W, CHAR_H   = 96, 128
FRAME_W, FRAME_H = 32, 32
CHARS_PER_ROW    = 4
DIRECTION_ROW    = 2   # facing right — best side profile

# ── Animal manifest ────────────────────────────────────────────────────────────
# Each entry: (sheet_file, char_slot, species, variant, state)
# state: "walk" | "eat"
# Walk slots produce both StateIdle (middle frame held) and StateWalking (all 3).
# Eat slots produce StateEating (all 3 frames).

MANIFEST = [
    # ── Dogs (animals1.png, slots 0-3 = walk) ─────────────────────────────────
    ('animals1.png', 0, 'dog', 'beagle',  'walk'),
    ('animals1.png', 1, 'dog', 'golden',  'walk'),
    ('animals1.png', 2, 'dog', 'scottie', 'walk'),
    ('animals1.png', 3, 'dog', 'brown',   'walk'),

    # ── Cats (animals1.png, slots 4-7 = walk) ─────────────────────────────────
    ('animals1.png', 4, 'cat', 'bicolor',  'walk'),
    ('animals1.png', 5, 'cat', 'gray',     'walk'),
    ('animals1.png', 6, 'cat', 'tabby',    'walk'),
    ('animals1.png', 7, 'cat', 'siamese',  'walk'),

    # ── Cows (animals2.png, slot 0 = walk, slot 4 = eat) ──────────────────────
    ('animals2.png', 0, 'cow',  'holstein', 'walk'),
    ('animals2.png', 4, 'cow',  'holstein', 'eat'),

    # ── Bulls (animals2.png, slot 1 = walk, slot 5 = eat) ─────────────────────
    ('animals2.png', 1, 'bull', 'brown', 'walk'),
    ('animals2.png', 5, 'bull', 'brown', 'eat'),

    # ── Pigs (animals2.png, slot 2 = walk) ────────────────────────────────────
    ('animals2.png', 2, 'pig',    'pink',   'walk'),
    ('animals2.png', 6, 'piglet', 'pink',   'walk'),

    # ── Goats (animals2.png, slots 3, 7 = walk) ───────────────────────────────
    ('animals2.png', 3, 'goat', 'white', 'walk'),
    ('animals2.png', 7, 'goat', 'dark',  'walk'),

    # ── Sheep (animals3.png, slots 0-3 = walk) ────────────────────────────────
    ('animals3.png', 0, 'sheep', 'woolly',  'walk'),
    ('animals3.png', 1, 'sheep', 'fluffy',  'walk'),
    ('animals3.png', 2, 'sheep', 'black',   'walk'),
    ('animals3.png', 3, 'sheep', 'gray',    'walk'),

    # ── Chicks & Chickens (animals3.png, slots 4-6 = walk) ────────────────────
    ('animals3.png', 4, 'chick',   'yellow', 'walk'),
    ('animals3.png', 5, 'chicken', 'brown',  'walk'),
    ('animals3.png', 6, 'chicken', 'white',  'walk'),

    # ── Turkey (animals3.png, slot 7 = walk) ──────────────────────────────────
    ('animals3.png', 7, 'turkey', 'brown', 'walk'),

    # ── Horses (horses.png, slots 0-3 = walk, slots 4-7 = eat) ───────────────
    ('horses.png', 0, 'horse', 'palomino',  'walk'),
    ('horses.png', 1, 'horse', 'bay',       'walk'),
    ('horses.png', 2, 'horse', 'gray',      'walk'),
    ('horses.png', 3, 'horse', 'dark',      'walk'),
    ('horses.png', 4, 'horse', 'palomino',  'eat'),
    ('horses.png', 5, 'horse', 'bay',       'eat'),
    ('horses.png', 6, 'horse', 'gray',      'eat'),
    ('horses.png', 7, 'horse', 'dark',      'eat'),
]


def extract_frames(img, char_slot, direction_row):
    """Extract all 3 walk frames for a given character slot and direction."""
    col = char_slot % CHARS_PER_ROW
    row = char_slot // CHARS_PER_ROW
    bx = col * CHAR_W
    by = row * CHAR_H
    dy = direction_row * FRAME_H

    frames = []
    for fx in range(3):
        x = bx + fx * FRAME_W
        y = by + dy
        frame = img.crop((x, y, x + FRAME_W, y + FRAME_H))
        frames.append(frame)
    return frames


def scale_up(frame, factor=4):
    """Scale a frame up with nearest-neighbor for better chafa rendering."""
    return frame.resize((frame.width * factor, frame.height * factor), Image.NEAREST)


def main():
    if os.path.exists(OUT):
        shutil.rmtree(OUT)
    os.makedirs(OUT)

    # Load sheets once
    sheets = {}
    for sheet_name in ['animals1.png', 'animals2.png', 'animals3.png', 'horses.png']:
        path = os.path.join(ASSETS, sheet_name)
        sheets[sheet_name] = Image.open(path).convert('RGBA')

    counts = {}
    for sheet_name, char_slot, species, variant, anim_state in MANIFEST:
        img = sheets[sheet_name]
        frames = extract_frames(img, char_slot, DIRECTION_ROW)

        key = f'{species}_{variant}'
        state_dir = os.path.join(OUT, key, anim_state)
        os.makedirs(state_dir, exist_ok=True)

        for i, frame in enumerate(frames):
            scaled = scale_up(frame, factor=4)
            out_path = os.path.join(state_dir, f'frame_{i:03d}.png')
            scaled.save(out_path)

        counts[key] = counts.get(key, 0) + len(frames)

        # For walk state also generate idle = middle frame only
        if anim_state == 'walk':
            idle_dir = os.path.join(OUT, key, 'idle')
            os.makedirs(idle_dir, exist_ok=True)
            # Middle frame (index 1) = neutral standing pose
            scale_up(frames[1], factor=4).save(os.path.join(idle_dir, 'frame_000.png'))

    print(f"Sliced {len(MANIFEST)} animal×state combos into {OUT}/")
    for key in sorted(counts):
        animal_dir = os.path.join(OUT, key)
        states = sorted(os.listdir(animal_dir))
        print(f"  {key}: {states}")


if __name__ == '__main__':
    main()
