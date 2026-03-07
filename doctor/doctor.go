package doctor

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Severity indicates how critical a check failure is.
type Severity int

const (
	Info    Severity = iota // nice-to-know, never blocks startup
	Warning                 // works but degraded; operator should fix
	Fatal                   // cannot start safely
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Fatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// CheckResult is the outcome of a single diagnostic check.
type CheckResult struct {
	Name     string
	Severity Severity
	Passed   bool
	Message  string
	FixHint  string
	Fixed    bool
}

// CheckFunc runs a single diagnostic check. If autoFix is true, the check
// should attempt to remediate the problem (e.g. create missing directories).
type CheckFunc func(autoFix bool) CheckResult

type namedCheck struct {
	name string
	fn   CheckFunc
}

// Runner collects checks and executes them in order.
type Runner struct {
	checks []namedCheck
}

// NewRunner returns an empty Runner.
func NewRunner() *Runner {
	return &Runner{}
}

// Add registers a check. Checks run in the order they are added.
func (r *Runner) Add(name string, fn CheckFunc) {
	r.checks = append(r.checks, namedCheck{name: name, fn: fn})
}

// Run executes all registered checks and returns a Report.
func (r *Runner) Run(autoFix bool) *Report {
	rpt := &Report{}
	for _, c := range r.checks {
		result := c.fn(autoFix)
		result.Name = c.name
		rpt.Results = append(rpt.Results, result)
		if result.Fixed || result.Passed {
			rpt.Passed++
		} else {
			switch result.Severity {
			case Warning:
				rpt.Warnings++
			case Fatal:
				rpt.Fatals++
			default:
				rpt.Passed++ // info-level failures are non-blocking
			}
		}
	}
	return rpt
}

// Report summarizes the results of a doctor run.
type Report struct {
	Results  []CheckResult
	Passed   int
	Warnings int
	Fatals   int
}

// LogSummary prints a compact summary suitable for startup logs.
// serviceName is used in the header (e.g. "Zentrale", "Workspace").
func (r *Report) LogSummary(serviceName string) {
	log.Printf("=== %s Doctor ===", serviceName)
	for _, res := range r.Results {
		status := "OK"
		if res.Fixed {
			status = "FIXED"
		} else if !res.Passed {
			status = strings.ToUpper(res.Severity.String())
		}
		log.Printf("  %-26s %s", res.Name+":", status)
		if !res.Passed && !res.Fixed && res.Message != "" {
			log.Printf("    -> %s", res.Message)
		}
	}
	log.Printf("  --- %d passed, %d warnings, %d fatal", r.Passed, r.Warnings, r.Fatals)
}

// PrintTerminal writes colorized output for the CLI doctor subcommand.
// Color is disabled when the NO_COLOR environment variable is set.
func (r *Report) PrintTerminal(w io.Writer, serviceName string) {
	noColor := os.Getenv("NO_COLOR") != ""

	const (
		green  = "\033[32m"
		yellow = "\033[33m"
		red    = "\033[31m"
		cyan   = "\033[36m"
		reset  = "\033[0m"
	)

	colorize := func(code, text string) string {
		if noColor {
			return text
		}
		return code + text + reset
	}

	fmt.Fprintf(w, "%s doctor\n", serviceName)
	fmt.Fprintln(w, strings.Repeat("-", 40))

	for _, res := range r.Results {
		var icon, color string
		switch {
		case res.Fixed:
			icon, color = "FIXD", cyan
		case res.Passed:
			icon, color = "PASS", green
		case res.Severity == Fatal:
			icon, color = "FAIL", red
		case res.Severity == Warning:
			icon, color = "WARN", yellow
		default:
			icon, color = "INFO", cyan
		}
		fmt.Fprintf(w, "  %s %-24s %s\n", colorize(color, "["+icon+"]"), res.Name, res.Message)
		if !res.Passed && !res.Fixed && res.FixHint != "" {
			fmt.Fprintf(w, "         hint: %s\n", res.FixHint)
		}
	}

	fmt.Fprintln(w, strings.Repeat("-", 40))
	summary := fmt.Sprintf("%d passed, %d warnings, %d fatal", r.Passed, r.Warnings, r.Fatals)
	if r.Fatals > 0 {
		fmt.Fprintln(w, colorize(red, summary))
	} else if r.Warnings > 0 {
		fmt.Fprintln(w, colorize(yellow, summary))
	} else {
		fmt.Fprintln(w, colorize(green, summary))
	}
}

// HealthJSON returns a map suitable for inclusion in a /api/health response.
func (r *Report) HealthJSON() map[string]interface{} {
	checks := make([]map[string]interface{}, 0, len(r.Results))
	for _, res := range r.Results {
		c := map[string]interface{}{
			"name":     res.Name,
			"passed":   res.Passed,
			"severity": res.Severity.String(),
		}
		if res.Message != "" {
			c["message"] = res.Message
		}
		checks = append(checks, c)
	}
	status := "healthy"
	if r.Fatals > 0 {
		status = "unhealthy"
	} else if r.Warnings > 0 {
		status = "degraded"
	}
	return map[string]interface{}{
		"status":   status,
		"passed":   r.Passed,
		"warnings": r.Warnings,
		"fatals":   r.Fatals,
		"checks":   checks,
	}
}

// ExitCode returns 0 if no fatal checks failed, 1 otherwise.
func (r *Report) ExitCode() int {
	if r.Fatals > 0 {
		return 1
	}
	return 0
}
