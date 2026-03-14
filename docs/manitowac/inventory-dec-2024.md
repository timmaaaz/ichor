# Inventory Snapshot — December 2024

Source: `docs/Dec 2024.xlsx` (Sheet1, 152 rows, 6 columns: Part Number, Description, Count, [Spools], [Est. Bobbins])

## Manufacturing Flow

The spreadsheet is organized by manufacturing stage — this is a **BOM snapshot**, not just an inventory count. The stages reveal the full relay production process:

```
Raw Material → Plastics → Coil Winding → Coil Riveting → Contact Riveting → Assembly → Calibration → Finished Relays
```

---

## Stage Breakdown

### 1. PLASTICS (Inbound)

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119644 | Covers, Black (1000/Box) | 15,120 |
| 119629 | Covers, White (1000/Box) | 52,000 |
| 119643 | Covers, DanFoss (1000/Box) | 52,680 |
| *(none)* | Covers, White KS | 4,830 |
| 119537 | Bases, Black (2000/Box) | 12,000 |
| 119535 | Bases, White (2000/Box) | 227,000 |
| 119559 | Bobbins (6000/Box, 3000/Bin) | 153,000 |

**Note:** "White KS" covers have no part number. Client confirmed everything *should* have a part number — this is a gap to resolve.

---

### 2. COIL WINDING (WIP / Processing)

Wire is tracked in two units: **spools** (physical count) and **estimated bobbins** (spools × 700, a formula-derived approximation).

| Part Number | Description | Spools | Est. Bobbins |
|-------------|-------------|--------|--------------|
| 073050 | Wire 0.05 gauge (37.5 lb/Spool) | 46 | ~32,200 |
| 073056 | Wire 0.056 gauge (38.5 lb/Spool) | 59 | ~41,300 |
| 073063 | Wire 0.063 gauge (39.5 lb/Spool) | 98 | ~68,600 |
| 073090 | Wire 0.09 gauge (41.5 lb/Spool) | 26 | ~18,200 |
| 119582 | Pins | 70,000 | — |

**Unfinished Coils (350/tray) — WIP, no part numbers assigned:**

| Coil # | Count |
|--------|-------|
| #2 | 8,750 |
| #3 | 5,600 |
| #4 | 8,400 |
| #5 | 700 |
| #6 | 8,050 |
| #7 | 7,000 |
| #9 | 3,500 |
| Imported Coil #8 | 16,170 |

Also: Mylar Tape (31213, qty 233), Wire Leads (Long/Short × Green/White/Blue/Red — no part numbers).

---

### 3. COIL RIVETING (WIP / Processing)

**Frames** — two sources tracked separately:

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119554 | Frames, MR&P (1000/bin) | 0 |
| *(none)* | Frames, Italy (1000/bin) | 280,200 |

**Cores** — four sources:

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119578 | Cores, China (3375/bin) | 174,000 |
| *(none)* | Cores, MR&P (3375/bin) | 22,000 |
| *(none)* | Cores, Italy (3375/bin) | 27,000 |
| *(none)* | Cores, Mitotec | 10,200 |

**Shading Coils:**

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119573 | Shading Coils, MR&P (15,000/Bin) | 237,000 |
| *(none)* | Shading Coils, Italy (15,000/Bin) | 35,000 |

**Other:**

| Part Number | Description | Count |
|-------------|-------------|-------|
| 209504 | Eyelets | 70,000 |
| 119507 | Screws | 60,000 |

**Finished Coils (200/tray) — post-riveting, no part numbers:**

| Coil # | Count |
|--------|-------|
| #2 | 15,400 |
| #3 | 7,200 |
| #4 | 7,200 |
| #5 | 400 |
| #6 | 3,800 |
| #7 | 4,400 |
| #8 | 1,000 |
| #9 | 7,600 |

---

### 4. CONTACT RIVETING (WIP / Processing)

**Terminals** — three types, each with MR&P and Italy sources:

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119680 | Terminals, Long MR&P (7,500/bin) | 0 |
| *(none)* | Terminals, Long Italy | 95,000 |
| 119681 | Terminals, Short MR&P (10,000/bin) | 0 |
| *(none)* | Terminals, Short Italy | 64,000 |
| 119682 | Terminals, Optional MR&P (3,000/bin) | 0 |
| *(none)* | Terminals, Optional Italy | 271,000 |

**Contacts:**

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119567 | Contacts, Standard | 180,000 |
| 119574 | Contacts, High Capacity | 40,000 |

**Other:**

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119568 | Rivets | 225,000 |
| 903105 | Washers, Standard | 100,000 |
| 119580 | Washers, High Capacity (30,000/Bin) | 250,000 |
| 119581 | Blades, MR&P (10,000/Bin) | 0 |
| *(none)* | Blades, Italy | 75,000 |

**Finished Bases** (no part numbers):

| Description | Count |
|-------------|-------|
| Finished Bases - 3 Terminal (35 Amp) | 11,250 |
| Finished Bases - 4 Terminal (35 Amp) | 5,100 |
| Finished Bases - 3 Terminal (50 Amp) | 4,950 |
| Finished Bases - 4 Terminal (50 Amp) | 5,100 |
| Finished Bases - K.S. | 9,150 |
| Finished Bases (Screw w/o screws) | 2,400 |
| Finished Bases (Screw w/screws) | 150 |
| Bases w/o Blades (Terminal) | 1,800 |
| Bases w/o Blades (Screw) | 1,400 |

---

### 5. ASSEMBLY + CALIBRATION

**Armatures:**

| Part Number | Description | Count |
|-------------|-------------|-------|
| *(none)* | Armatures, Bent Standard (1,500/Bin) | 18,000 |
| 119546 | Armatures, Unbent Standard | 45,500 |
| *(none)* | Armatures, Unbent Nickel | 37,500 |

**Other assembly components:**

| Part Number | Description | Count |
|-------------|-------------|-------|
| 119610 | Actuators, White | 65,000 |
| *(none)* | Actuators, Black | 13,000 |

**Springs (no part numbers, 7 types):**

| Spring # | Count |
|----------|-------|
| 29 (12,500/unit) | 10,000 |
| 30 (12,500/unit) | 170,000 |
| 31 (13,750/unit) | 156,250 |
| 32 (12,500/unit) | 175,000 |
| 33 (20,000/unit) | 160,000 |
| 34 | 37,000 |
| 35 (10,000/unit) | 40,000 |

**Calibration WIP and Finished Output:**

| Description | Count |
|-------------|-------|
| Unfinished Relays (100/Tray) | 18,200 |
| **Finished Relays (100/Box)** | **71,000** |

---

### 6. RAW MATERIAL (Inbound / Stock)

All items have no part numbers in this snapshot.

| Material | Count/Unit |
|----------|------------|
| White Resin | 16,500 |
| Black Resin | 8,100 |
| Steel Bars | 11,000 |
| Steel for Frames | 12,235 |
| Bronze for Hi-Cap Washer | 0 |
| Phos. Bronze for Blade | 3,114 |
| Armature Material | 4,637 |
| Opt. Term Material | 2,653 |
| Short Term Material | 5,586 |
| Long Term Material | 9,499 |
| Shade Coil Material | 1,145 |
| Wire Solder | 360 |
| Bar Solder | 240 |
| Lead Wire, Green | 25,000 |
| Lead Wire, White | 30,000 |
| Lead Wire, Blue | 15,000 |
| Lead Wire, Red | 25,000 |

---

## Part Number Coverage

| Stage | Total Items | Have Part Number | Missing Part Number |
|-------|-------------|-----------------|---------------------|
| Plastics | 7 | 6 | 1 (White KS) |
| Coil Winding | 13 | 5 | 8 (unfinished coils, wire leads) |
| Coil Riveting | 18 | 5 | 13 (Italy source items, finished coils) |
| Contact Riveting | 20 | 8 | 12 (Italy source items, bases) |
| Assembly | 12 | 2 | 10 (armatures, actuators, springs) |
| Raw Material | 17 | 0 | 17 |
| **Total** | **87** | **26** | **61 (~70%)** |

---

## Key Observations

1. **~70% of line items lack a part number** — mostly WIP, raw material, and Italy-sourced items.
2. **Wire is tracked in two units simultaneously** — spools (physical) and estimated bobbins (spools × 700). No UOM field exists in the current schema.
3. **Coil numbers (#2–#9) are product variants**, not just sizes — they represent different relay specifications.
4. **"Bobbins" appears in two contexts**: plastic bobbins (119559, incoming component) and wire bobbins (winding output, ~700 per wire spool).
5. **MR&P stock is consistently 0** across most categories in December — all current stock is Italy-sourced.
6. **Finished Relays** are the sole outbound product (71,000 finished + 18,200 in calibration).
