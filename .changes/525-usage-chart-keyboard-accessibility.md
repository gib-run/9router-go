### fix(usage): make the usage chart reachable from the keyboard

The usage chart drew a set of stacked areas and a tooltip that opened under
the pointer. The drawn areas carry no text a screen reader can read, and a
tooltip that only opens on hover is unreadable to anyone driving the page from
the keyboard — the numbers the chart shows existed nowhere else.

The bucket overlays are now focusable and announce themselves, and the arrow
keys walk the same points the mouse reveals. The tooltip is the same node in
both cases and carries `role="status"`, so a screen reader hears the point
the keyboard just moved to.

**The same numbers are also available as a table.** The chart's figures are
repeated in a `<details>` disclosure, one row per bucket, with the day (or
hour) label in a row header. A collapsed `<details>` hides its content from
the accessibility tree, so the disclosure is opened the way a keyboard user
reaches it before the table is queried.

**The table's rows are keyed by index, not by label.** A daily bucket is
labelled "Jan 2", so a period long enough to cross a year boundary repeats
it — and Svelte throws `each_key_duplicate` on the second occurrence, which
takes the whole usage page down rather than just the table. The index is
unique by construction, and it is the same key the plotted points use.

The e2e suite covers both halves: the table exposes one row per bucket, and
ArrowRight moves the tooltip to the next one.
