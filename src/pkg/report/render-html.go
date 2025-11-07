package report

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"time"
)

/*
Render the entire HTML report into a buffer.

barRef      -> target duration label (e.g., 12m).
barHeightPx -> pixel height that corresponds to barRef (used to scale bars).
*/
func renderHTMLReport(buf *bytes.Buffer, daySummaries []DaySummary, totals ReportTotals, barRef time.Duration, barHeightPx int, startDate, endDate time.Time) {
	// ---------- precompute ----------
	refSeconds := barRef.Seconds()
	if refSeconds <= 0 {
		refSeconds = 1 // avoid div-by-zero; degenerate but safe
	}
	pxPerSecond := float64(barHeightPx) / refSeconds

	// helper: convert a duration to pixel height (at least 1px if positive)
	segHeight := func(d time.Duration) int {
		if d <= 0 {
			return 0
		}
		h := int(math.Round(d.Seconds() * pxPerSecond))
		if h == 0 {
			h = 1
		}
		return h
	}

	// Precompute per-day container heights and row max for "Time by Day"
	dayContainerHeights := make([]int, len(daySummaries))
	maxDayRowHeight := barHeightPx
	for i, ds := range daySummaries {
		h := segHeight(ds.TotalDuration)
		container := h
		if container < barHeightPx {
			container = barHeightPx
		}
		dayContainerHeights[i] = container
		if container > maxDayRowHeight {
			maxDayRowHeight = container
		}
	}

	// Precompute per-day container heights and row max for "Activity×Time"
	smContainerHeights := make([]int, len(daySummaries))
	maxSmRowHeight := barHeightPx
	for i, ds := range daySummaries {
		h := segHeight(ds.SmoothedActiveTime)
		container := h
		if container < barHeightPx {
			container = barHeightPx
		}
		smContainerHeights[i] = container
		if container > maxSmRowHeight {
			maxSmRowHeight = container
		}
	}

	// task listing (sorted by total desc)
	taskNames := make([]string, 0, len(totals.PerTaskTotals))
	for k := range totals.PerTaskTotals {
		taskNames = append(taskNames, k)
	}
	sort.Slice(taskNames, func(i, j int) bool {
		// pin "Unassigned Time" to the top
		ai, aj := taskNames[i], taskNames[j]
		if ai == "Unassigned Time" && aj != "Unassigned Time" {
			return true
		}
		if aj == "Unassigned Time" && ai != "Unassigned Time" {
			return false
		}

		// sort other cases
		di := totals.PerTaskTotals[ai]
		dj := totals.PerTaskTotals[aj]
		if di == dj {
			return taskNames[i] < taskNames[j]
		}
		return di > dj
	})

	avgActivity := 0.0
	if totals.TotalWorked > 0 {
		avgActivity = (float64(totals.TotalActive) / float64(totals.TotalWorked)) * 100.0
	}
	activityHex := colorToHex(barColorFor(avgActivity))

	// Activity color legend swatches
	hex0 := colorToHex(barColorFor(0))
	hex50 := colorToHex(barColorFor(50))
	hex75 := colorToHex(barColorFor(75))
	hex100 := colorToHex(barColorFor(100))

	// ---------- Chart geometry (classic for ≤14 days, adaptive for >14) ----------
	nDays := len(daySummaries)
	if nDays < 1 {
		nDays = 1
	}
	weeklyMode := nDays <= 14
	showBarLabels := weeklyMode // NEW: hide in-bar labels when period > 2 weeks

	// Outputs we’ll use below
	barW := 28
	pad := 8
	perDaySlot := barW + 2*pad
	chartW := nDays * perDaySlot

	if !weeklyMode {
		// Adaptive geometry for long ranges
		const maxInnerW = 640 // inner width that fits your 760px wrapper
		const minBarW = 1
		const minPad = 1

		perDayW := int(math.Floor(float64(maxInnerW) / float64(nDays)))
		if perDayW < (minBarW + 1*minPad) {
			perDayW = minBarW + 1*minPad
		}
		pad = perDayW / 4 // ~25% of slot as padding
		if pad < minPad {
			pad = minPad
		}
		barW = perDayW - 1*pad
		if barW < minBarW {
			barW = minBarW
		}
		perDaySlot = barW + 1*pad
		chartW = nDays * perDaySlot
		if chartW > maxInnerW {
			chartW = maxInnerW
		}
	}

	// ---------- Label strategy ----------
	// Weekly: show all day names exactly like before (no clipping box).
	// Long ranges: thin or replace with ticks; also clip width to per-day slot so text can’t spill.
	labelStep := 0
	useTicks := false
	labelStyle := "font-family:Arial, sans-serif;font-size:12px;color:#555;padding-top:6px;" // weekly default

	if !weeklyMode {
		useTicks = false
		labelStyle = fmt.Sprintf("font-family:Arial, sans-serif;font-size:12px;color:#555;padding-top:6px;width:%dpx;overflow:hidden;white-space:nowrap;line-height:1;", perDaySlot)
	}

	// ---------- HTML ----------
	fmt.Fprintf(buf, `<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <title>%s</title>
  </head>
  <body style="margin:0;padding:0;background:#fafafa;">
    <!-- Wrapper -->
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:760px;margin:0 auto;background:#fff;border-collapse:collapse;">
      <tr>
        <td style="font-family:Arial, sans-serif;color:#222;font-size:16px;padding:12px 18px;border-bottom:1px solid #eee;">
          %s
        </td>
      </tr>

      <!-- Summary Section -->
      <tr>
        <td align="center" style="padding:14px 8px;">
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
            <tr valign="middle">
              <td align="center" style="padding:0 16px;">
                <div style="font-family:Arial, sans-serif;font-size:13px;color:#666;padding-bottom:15px;padding-top:0px;">Total Worked</div>
                <div style="font-family:Arial, sans-serif;font-size:38px;color:#111;font-weight:bold;">%s</div>
              </td>
              <td align="center" style="padding:0 16px;">
                <div style="font-family:Arial, sans-serif;font-size:13px;color:#666;padding-bottom:4px;">Avg Activity</div>
                %s
                <div style="font-family:Arial, sans-serif;font-size:24px;color:#111;font-weight:bold;margin-top:4px;">%.1f%%</div>
              </td>
            </tr>
          </table>
        </td>
      </tr>

      <!-- Tasks in period (vertical list, centered) -->
      <tr>
        <td align="center" style="padding:4px 12px 10px 12px;">
          <div style="font-family:Arial, sans-serif;font-size:14px;color:#444;padding-bottom:6px;font-weight:bold;">Tasks in period</div>
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
`,
		reportTitle(startDate, endDate),
		reportTitle(startDate, endDate),
		formatDuration(totals.TotalWorked),
		buildSquares10HTML(avgActivity, activityHex),
		avgActivity,
	)

	for i, name := range taskNames {
		dur := totals.PerTaskTotals[name]
		c := taskColorHex(i, name)
		fmt.Fprintf(buf, `            <tr>
              <td style="padding:4px 10px;font-family:Arial, sans-serif;font-size:13px;color:#333;vertical-align:middle;text-align:center;">
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr>
                    <td style="background:%s;width:12px;height:12px;line-height:0;font-size:0;">&nbsp;</td>
                    <td style="padding-left:8px;">%s&nbsp;<span style="color:#666;">— %s</span></td>
                  </tr>
                </table>
              </td>
            </tr>
`, c, esc(name), esc(formatDuration(dur)))
	}

	fmt.Fprintf(buf, `          </table>
        </td>
      </tr>

      <!-- Time by Day (stacked per task) -->
      <tr>
        <td align="center" style="padding:15px 0 10px 0;">
          <div style="font-family:Arial, sans-serif;color:#222;font-size:14px;font-weight:bold;">Time by Day (%s baseline)</div>
        </td>
      </tr>

      <!-- Centered chart wrapper -->
      <tr>
        <td align="center" style="padding:2px 0 0 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" width="%d" style="border-collapse:collapse;">
            <tr>
              <td style="padding:0 20px;">
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr valign="bottom">
`, esc(formatDuration(barRef)), chartW)

	for dayIdx, dsum := range daySummaries {
		containerH := dayContainerHeights[dayIdx]
		dayTotalH := segHeight(dsum.TotalDuration)
		if dayTotalH > containerH {
			dayTotalH = containerH
		}
		topSpacer := containerH - dayTotalH
		if topSpacer < 0 {
			topSpacer = 0
		}

		// order tasks by global order, include only present tasks
		dayTasks := make([]string, 0, len(taskNames))
		for _, tname := range taskNames {
			if dsum.TaskDurations[tname] > 0 {
				dayTasks = append(dayTasks, tname)
			}
		}

		// NOTE: the outer <td> has only horizontal padding = pad, so the per-column
		// width is exactly perDaySlot, matching the bar/table inside (no hidden widening).
		fmt.Fprintf(buf, `                    <!-- Day bar -->
                    <td align="center" style="padding:0 %dpx;">
                      <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                        <tr><td style="background:#e0e0e0;position:relative;width:%dpx;height:%dpx;line-height:0;font-size:0;">
                          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:%dpx;">
`, pad, barW, containerH, barW)

		if topSpacer > 0 {
			fmt.Fprintf(buf, `                            <tr><td style="height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
`, topSpacer)
		}
		for segIdx, tname := range dayTasks {
			seg := segHeight(dsum.TaskDurations[tname])
			if seg <= 0 {
				continue
			}
			fmt.Fprintf(buf, `                            <tr><td style="background:%s;height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
`, taskColorHex(segIdx, tname), seg)
		}

		// Hours label inside bar (bottom-center) — now conditional
		hoursLabel := fmt.Sprintf("%.1fh", dsum.TotalDuration.Hours())
		hoursLabelHTML := ""
		if showBarLabels {
			hoursLabelHTML = fmt.Sprintf(
				`<div style="position:absolute;left:0;right:0;bottom:2px;text-align:center;">
                   <span style="font-family:Arial, sans-serif;font-size:11px;color:#000;">%s</span>
                 </div>`, esc(hoursLabel))
		}

		fmt.Fprintf(buf, `                          </table>
                          %s
                        </td></tr>
                      </table>
`, hoursLabelHTML)

		// Label (weekly: full text; long range: thinned/tick/hidden)
		var labelHTML string
		if weeklyMode {
			labelHTML = fmt.Sprintf(`<div style="%s">%s</div>`, labelStyle, esc(weekdayShort(dsum.Date)))
		} else {
			if labelStep != 0 && (dayIdx%labelStep == 0) {
				if useTicks {
					labelHTML = `<div style="height:8px;margin:4px auto 0;line-height:0;">
                                   <div style="width:0;height:8px;border-left:1px solid #aaa;margin:0 auto;"></div>
                                 </div>`
				} else {
					labelHTML = fmt.Sprintf(`<div style="%s">%s</div>`, labelStyle, esc(weekdayShort(dsum.Date)))
				}
			} else {
				labelHTML = fmt.Sprintf(`<div style="%s"></div>`, labelStyle)
			}
		}

		fmt.Fprintf(buf, `                      %s
                    </td>
`, labelHTML)
	}

	fmt.Fprintf(buf, `                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>

      <!-- Legend for tasks -->
      <tr>
        <td align="center" style="padding:6px 0 10px 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
            <tr>
`)
	for idx, name := range taskNames {
		if idx > 0 {
			fmt.Fprintf(buf, `              <td style="width:12px;">&nbsp;</td>
`)
		}
		fmt.Fprintf(buf, `              <td style="background:%s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 0 0 6px;">%s</td>
`, taskColorHex(idx, name), esc(name))
	}
	fmt.Fprintf(buf, `            </tr>
          </table>
        </td>
      </tr>

      <!-- Activity × Time -->
      <tr>
        <td align="center" style="padding:15px 0 10px 0;">
          <div style="font-family:Arial, sans-serif;color:#222;font-size:14px;font-weight:bold;">Activity × Time (%s baseline)</div>
        </td>
      </tr>

      <!-- Centered chart wrapper -->
      <tr>
        <td align="center" style="padding:2px 0 6px 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" width="%d" style="border-collapse:collapse;">
            <tr>
              <td>
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr valign="bottom">
`, esc(formatDuration(barRef)), chartW)

	for dayIdx, dsum := range daySummaries {
		dayPct := 0.0
		if dsum.TotalDuration > 0 {
			dayPct = (float64(dsum.TotalActive) / float64(dsum.TotalDuration)) * 100.0
		}
		hex := colorToHex(barColorFor(dayPct))

		containerH := smContainerHeights[dayIdx]
		h := segHeight(dsum.SmoothedActiveTime)
		if h > containerH {
			h = containerH
		}
		top := containerH - h
		if top < 0 {
			top = 0
		}

		// Activity % label for this day — now conditional
		pctLabel := fmt.Sprintf("%.0f%%", dayPct)
		pctLabelHTML := ""
		if showBarLabels {
			pctLabelHTML = fmt.Sprintf(
				`<div style="position:absolute;left:0;right:0;bottom:2px;text-align:center;">
                   <span style="font-family:Arial, sans-serif;font-size:11px;color:#000;">%s</span>
                 </div>`, esc(pctLabel))
		}

		fmt.Fprintf(buf, `                    <td align="center" style="padding:0 %dpx;">
                      <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                        <tr><td style="background:#e0e0e0;position:relative;width:%dpx;height:%dpx;line-height:0;font-size:0;">
                          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:%dpx;">
                            <tr><td style="height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
                            <tr><td style="background:%s;height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
                          </table>
                          %s
                        </td></tr>
                      </table>
`, pad, barW, containerH, barW, top, hex, h, pctLabelHTML)

		// Label (same rules as above)
		var labelHTML string
		if weeklyMode {
			labelHTML = fmt.Sprintf(`<div style="%s">%s</div>`, labelStyle, esc(weekdayShort(dsum.Date)))
		} else {
			if labelStep != 0 && (dayIdx%labelStep == 0) {
				if useTicks {
					labelHTML = `<div style="height:8px;margin:4px auto 0;line-height:0;">
                                   <div style="width:0;height:8px;border-left:1px solid #aaa;margin:0 auto;"></div>
                                 </div>`
				} else {
					labelHTML = fmt.Sprintf(`<div style="%s">%s</div>`, labelStyle, esc(weekdayShort(dsum.Date)))
				}
			} else {
				labelHTML = fmt.Sprintf(`<div style="%s"></div>`, labelStyle)
			}
		}
		fmt.Fprintf(buf, `                      %s
                    </td>
`, labelHTML)
	}

	fmt.Fprintf(buf, `                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>

      <!-- Activity × Time legend -->
      <tr>
        <td align="center" style="padding:2px 0 20px 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
            <tr>
              <td style="background:%[1]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">0%%</td>

              <td style="background:%[2]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">50%%</td>

              <td style="background:%[3]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">75%%</td>

              <td style="background:%[4]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 0 0 6px;">100%%</td>
            </tr>
          </table>
        </td>
      </tr>

    </table> <!-- end white wrapper -->

	<!-- Footer OUTSIDE content area -->
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
      <tr>
        <td align="center" style="padding:20px 0;">
          <div style="font-family:Arial, sans-serif;font-size:12px;color:#888;">
            Generated by Work Tracker
          </div>
        </td>
      </tr>
    </table>

  </body>
</html>
`, hex0, hex50, hex75, hex100)
}
