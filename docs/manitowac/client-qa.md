# Client Q&A — Manitowac Discovery

Questions asked during initial inventory analysis. Answers updated as they come in.

---

## Part Numbers

**Q: Are there a lot of parts with no part number (like "White KS"), or is this rare?**
- Everything *should* have a part number — we can make it required
- Items in assembly may not have a part number yet → we can assign one for them
- **Decision:** Part number required for finished/purchased parts; WIP items need assigned SKUs

---

## Display / UI

**Q: Single scrollable page or tabbed by part/category?**
- Make it a setting so users can choose both views

---

## Pins

**Q: Does part number 119582 apply to all the different pin variants?**
- Pins go into bobbins (plastic bobbins, the winding component)
- Unfinished coils = bobbins on the shelf that are in-progress toward assembly
- **Context:** "Bobbins" appears in two meanings — plastic bobbins (inbound component 119559) and wire bobbins (output of coil winding, ~700 per wire spool)

---

## Categories

**Q: Do we actually want to categorize things?**
- Yes, categories are desired
- No need for variants (variants as a separate concept were ruled out)

---

## Wire / Coil Tracking

**Q: What is the non-"spools" number for coil winding wire?**
- Bobbins of wire (the wound output)
- ~700 bobbins per spool is an approximation
- The spreadsheet uses a formula: `=Spools × 700` for estimated bobbin count

**Q: Spool numbers don't seem to add up — why?**
- *Answer pending*

---

## Unfinished Coils

**Q: What are the part numbers for unfinished coils? Do they even have one?**
- They don't currently have part numbers
- We can assign them
- **Decision:** Assign internal part numbers for WIP coil items (#2–#9)

---

## Frames — MR&P vs Italy

**Q: Are MR&P and Italy frames variants (implying source matters) or different products?**
- They are **different** (source matters — they are not freely interchangeable)
- **Answer:** Technically interchangeable on the floor, but significant quality difference between sources.
- **Decision:** One product + two `supplier_products` records + lot tracking. Quality handled at lot level via `lot_trackings.quality_status`. No schema changes needed.

---

## Finished Coils

**Q: Is "Finished coils (200/Tray)" an outgoing item? How many were created vs used on finished products?**
- *Answer pending*

---

## Coil Numbers (#2–#9)

**Q: Are coil numbers (#2–#9) separate products, or are they variants of one "coil" product?**
- The spreadsheet tracks them as separate line items (Unfinished Coil #2, Finished Coil #2, etc.) with no part numbers
- They appear at both the unfinished (post-winding) and finished (post-riveting) stages
- Do #2 and #3 represent different relay specs, different wire gauges, different dimensions — or just different production run labels?
- *Answer pending — determines whether we assign separate SKUs per coil number or model them another way*

---

## Inbound vs Outbound

**Q: Which items are incoming, which are outbound?**
- *Answer pending — see manufacturing stages in [inventory-dec-2024.md](inventory-dec-2024.md) for current mapping*
- Finished Relays (100/Box) are clearly the outbound product

---

## Manufacturing Stages

**Q: Stages in addition to categories?**
- Yes, stages are wanted:
  - Inbound
  - Received
  - Processing → tied to order only sometimes
  - Assembly
    - Calibration
    - QA
  - Outbound
    - In stock vs on order
- **Schema impact:** No "stage" field exists in Ichor today — this is a schema gap (see [schema-decisions.md](schema-decisions.md))

---

## Armatures — Bent vs Unbent

**Q: Are "Armatures (Bent)" and "Armatures (Unbent)" the same physical part at different WIP stages, or two genuinely different products?**
- Bending sounds like a processing step applied to an unbent armature
- If same part: "bent" = the armature after a processing operation → modeled as zone/stage, one product
- If different: separate SKUs with different specs (material, dimensions, tolerances)
- *Answer pending — affects whether bending is a stage transition or a separate product*

---

## Imported Coil #8

**Q: "Imported Coil #8" is listed separately from the other unfinished coils. Is this the same spec as regular Coil #8 but sourced externally, or a different product entirely?**
- If same spec, different source → one product + supplier record (same model as Italy/MR&P frames)
- If different spec → separate SKU
- *Answer pending*

---

## Wire Leads vs Lead Wire

**Q: Are "Wire Leads" (pre-cut colored wires in Coil Winding) and "Lead Wire" (bulk raw wire in Raw Material) the same material at different processing stages?**
- If yes: raw lead wire gets cut to length → becomes a wire lead. This is a WIP relationship — raw material consumed to produce a component.
- If no: they are unrelated materials that happen to both be wire
- *Answer pending — affects how these products are classified and whether a BOM relationship exists between them*

---

## Springs (#29–#35)

**Q: Are spring numbers (#29, #30, #31, etc.) different product specifications (separate SKUs) or just batch/run labels for the same spring?**
- 7 types listed with different per-bin counts (12,500 / 13,750 / 20,000)
- Different bin sizes suggest different physical dimensions
- *Answer pending — same structural question as coil numbers*

---

## "KS" Suffix

**Q: What does "KS" mean on "Covers, White KS" and "Finished Bases - K.S."?**
- Customer specification? Size variant? Industry abbreviation?
- Both items have no part number
- *Answer pending — affects product classification and whether KS is a category, a customer tag, or something else*

---

## Duplicate Part Number 119568

**Q: Part number 119568 appears twice in the spreadsheet — once on "Rivets" (qty 225,000) and once as the header label for the "BASES" section. Which is correct?**
- Almost certainly a spreadsheet data entry error
- *Answer pending — need correct part number for Bases section header, if it has one*

---

## MR&P Supplier Status

**Q: MR&P stock shows 0 across nearly every category in December (Frames, Terminals, Blades, Contacts). Is MR&P still an active supplier, or have you fully switched to Italy-sourced parts?**
- Affects whether MR&P supplier records are active or should be flagged as inactive/historical
- *Answer pending*

---

## Finished Bases — 7 Variants

**Q: The 7 finished base variants (3/4 Terminal × 35/50 Amp × K.S. × Screw types) all have no part numbers. Are all 7 tracked and reported separately, or are some interchangeable?**
- If tracked separately → 7 distinct products, each needs an assigned SKU
- If some are interchangeable → fewer SKUs, variant logic at lot level
- *Answer pending*

---

## Calibration — Separate Zone or Same Floor?

**Q: Is "Calibration" a physically separate area from the rest of Assembly, or does it happen in the same space?**
- If separate area → gets its own zone with `stage = calibration`
- If same floor → calibration is a process state, not a location (needs explicit stage field on lot instead)
- *Answer pending — affects zone setup for Manitowac*

---

## Mylar Tape — Component or Consumable?

**Q: Is Mylar Tape (119507) a component that ends up in the finished product, or a consumable used during the winding process and not present in the final relay?**
- Component → `inventory_type = component`, tracked through BOM
- Consumable → `inventory_type = consumable`, tracked by usage rate, not by unit in product
- *Answer pending*
