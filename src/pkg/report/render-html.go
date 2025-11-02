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

	// Chart geometry for centering (per-day width = bar width + left/right padding)
	const barW = 28
	const pad = 8
	perDayW := barW + 2*pad
	chartW := len(daySummaries) * perDayW // used to center the inner table

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

      <!-- Centered chart wrapper with top-left label INSIDE (doesn't affect centering) -->
      <tr>
        <td align="center" style="padding:2px 0 0 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" width="%d" style="border-collapse:collapse;">
            <tr>
              <td>
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr valign="bottom">
`, esc(formatDuration(barRef)), chartW)

	for i, dsum := range daySummaries {
		containerH := dayContainerHeights[i]
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

		fmt.Fprintf(buf, `                    <!-- Day bar -->
                    <td align="center" style="padding:0 %dpx;">
                      <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                        <tr><td style="background:#e0e0e0;width:%dpx;height:%dpx;line-height:0;font-size:0;">
                          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:%dpx;">
`, pad, barW, containerH, barW)

		if topSpacer > 0 {
			fmt.Fprintf(buf, `                            <tr><td style="height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
`, topSpacer)
		}
		for i, tname := range dayTasks {
			seg := segHeight(dsum.TaskDurations[tname])
			if seg <= 0 {
				continue
			}
			fmt.Fprintf(buf, `                            <tr><td style="background:%s;height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
`, taskColorHex(i, tname), seg)
		}

		fmt.Fprintf(buf, `                          </table>
                        </td></tr>
                      </table>
                      <div style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding-top:6px;">%s</div>
                    </td>
`, esc(weekdayShort(dsum.Date)))
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
          <div style="font-family:Arial, sans-serif;color:#222;font-size:14px;font-weight:bold;">Activity × Time</div>
        </td>
      </tr>

      <!-- Centered chart wrapper with top-left label for the smoothed chart -->
      <tr>
        <td align="center" style="padding:2px 0 6px 0;">
          <table role="presentation" cellpadding="0" cellspacing="0" width="%d" style="border-collapse:collapse;">
            <tr>
              <td>
                <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 auto;">
                  <tr valign="bottom">
`, chartW)

	for i, dsum := range daySummaries {
		dayPct := 0.0
		if dsum.TotalDuration > 0 {
			dayPct = (float64(dsum.TotalActive) / float64(dsum.TotalDuration)) * 100.0
		}
		hex := colorToHex(barColorFor(dayPct))

		containerH := smContainerHeights[i]
		h := segHeight(dsum.SmoothedActiveTime)
		if h > containerH {
			h = containerH
		}
		top := containerH - h
		if top < 0 {
			top = 0
		}

		fmt.Fprintf(buf, `                    <td align="center" style="padding:0 %dpx;">
                      <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;">
                        <tr><td style="background:#e0e0e0;width:%dpx;height:%dpx;line-height:0;font-size:0;">
                          <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:%dpx;">
                            <tr><td style="height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
                            <tr><td style="background:%s;height:%dpx;line-height:0;font-size:0;">&nbsp;</td></tr>
                          </table>
                        </td></tr>
                      </table>
                      <div style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding-top:6px;">%s</div>
                    </td>
`, pad, barW, containerH, barW, top, hex, h, esc(weekdayShort(dsum.Date)))
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
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">%[5]s × 0%%</td>

              <td style="background:%[2]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">%[5]s × 50%%</td>

              <td style="background:%[3]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 12px 0 6px;">%[5]s × 75%%</td>

              <td style="background:%[4]s;width:10px;height:10px;line-height:0;font-size:0;">&nbsp;</td>
              <td style="font-family:Arial, sans-serif;font-size:12px;color:#555;padding:0 0 0 6px;">%[5]s × 100%%</td>
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
`, hex0, hex50, hex75, hex100, esc(formatDuration(barRef)))
}
