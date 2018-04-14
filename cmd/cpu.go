package cmd

import (
	"github.com/olorin/nagiosplugin"
	"github.com/soniah/gosnmp"
	"github.com/spf13/cobra"
)

const (
	hpSwitchCPUStat = "1.3.6.1.4.1.11.2.14.11.5.1.9.6.1.0"
)

func cpuCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cpu",
		Short: "Check the CPU usage",
		Long: `Query the switch, retrieve the CPU usage and check it against
the given boundaries.`,
	}

	warn := cmd.Flags().StringP("warning", "w", "", "Warning threshold (required)")
	crit := cmd.Flags().StringP("critical", "c", "", "Critical threshold (required)")

	cmd.MarkFlagRequired("warning")  // nolint: gas
	cmd.MarkFlagRequired("critical") // nolint: gas

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// parse ranges
		warnRange, err := nagiosplugin.ParseRange(*warn)
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		critRange, err := nagiosplugin.ParseRange(*crit)
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// query the SNMP host for the oids
		result, err := gosnmp.Default.Get([]string{
			hpSwitchCPUStat,
		})
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// extract the snmap results
		load := float64(gosnmp.ToBigInt(result.Variables[0].Value).Uint64())

		// instance a new nagios check
		check := nagiosplugin.NewCheck()
		defer check.Finish()

		// add performance data
		err = check.AddPerfDatum("LOAD", "%", load)
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// start out with OK and work your way up to the worst
		state := nagiosplugin.OK

		if warnRange.Check(load) {
			state = nagiosplugin.WARNING
		}

		if critRange.Check(load) {
			state = nagiosplugin.CRITICAL
		}

		// add the result with a nice text
		check.AddResultf(state, "CPU %.1f%%", load)
	}

	return cmd
}
