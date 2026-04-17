package stats

import (
	"time"

	"github.com/GerardoFC8/claumeter/internal/usage"
)

type FilterPreset int

const (
	FilterAll FilterPreset = iota
	FilterToday
	FilterYesterday
	FilterLast7Days
	FilterLast30Days
	FilterLast90Days
	FilterThisWeek
	FilterThisMonth
)

var allFilters = []FilterPreset{
	FilterAll,
	FilterToday,
	FilterYesterday,
	FilterLast7Days,
	FilterLast30Days,
	FilterLast90Days,
	FilterThisWeek,
	FilterThisMonth,
}

func (p FilterPreset) Label() string {
	switch p {
	case FilterAll:
		return "All time"
	case FilterToday:
		return "Today"
	case FilterYesterday:
		return "Yesterday"
	case FilterLast7Days:
		return "Last 7 days"
	case FilterLast30Days:
		return "Last 30 days"
	case FilterLast90Days:
		return "Last 90 days"
	case FilterThisWeek:
		return "This week"
	case FilterThisMonth:
		return "This month"
	}
	return "?"
}

func (p FilterPreset) Next() FilterPreset {
	for i, f := range allFilters {
		if f == p {
			return allFilters[(i+1)%len(allFilters)]
		}
	}
	return FilterAll
}

func (p FilterPreset) Prev() FilterPreset {
	for i, f := range allFilters {
		if f == p {
			return allFilters[(i-1+len(allFilters))%len(allFilters)]
		}
	}
	return FilterAll
}

// Range returns [from, to) in local time. Zero values mean unbounded.
func (p FilterPreset) Range(now time.Time) (time.Time, time.Time) {
	loc := now.Location()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	startOfTomorrow := startOfToday.AddDate(0, 0, 1)

	switch p {
	case FilterAll:
		return time.Time{}, time.Time{}
	case FilterToday:
		return startOfToday, startOfTomorrow
	case FilterYesterday:
		return startOfToday.AddDate(0, 0, -1), startOfToday
	case FilterLast7Days:
		return startOfToday.AddDate(0, 0, -6), startOfTomorrow
	case FilterLast30Days:
		return startOfToday.AddDate(0, 0, -29), startOfTomorrow
	case FilterLast90Days:
		return startOfToday.AddDate(0, 0, -89), startOfTomorrow
	case FilterThisWeek:
		wd := int(now.Weekday())
		if wd == 0 {
			wd = 7
		}
		monday := startOfToday.AddDate(0, 0, -(wd - 1))
		return monday, startOfTomorrow
	case FilterThisMonth:
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return firstOfMonth, startOfTomorrow
	}
	return time.Time{}, time.Time{}
}

func (p FilterPreset) Apply(data usage.Data) usage.Data {
	if p == FilterAll {
		return data
	}
	from, to := p.Range(time.Now())
	if from.IsZero() && to.IsZero() {
		return data
	}

	inRange := func(t time.Time) bool {
		t = t.Local()
		if !from.IsZero() && t.Before(from) {
			return false
		}
		if !to.IsZero() && !t.Before(to) {
			return false
		}
		return true
	}

	out := usage.Data{
		Events:   make([]usage.Event, 0, len(data.Events)),
		Prompts:  make([]usage.Prompt, 0, len(data.Prompts)),
		ToolUses: make([]usage.ToolUse, 0, len(data.ToolUses)),
	}
	for _, e := range data.Events {
		if inRange(e.Timestamp) {
			out.Events = append(out.Events, e)
		}
	}
	for _, pm := range data.Prompts {
		if inRange(pm.Timestamp) {
			out.Prompts = append(out.Prompts, pm)
		}
	}
	for _, tu := range data.ToolUses {
		if inRange(tu.Timestamp) {
			out.ToolUses = append(out.ToolUses, tu)
		}
	}
	return out
}
