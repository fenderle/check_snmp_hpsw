package cmd

import (
	"github.com/olorin/nagiosplugin"
	"github.com/soniah/gosnmp"
	"github.com/spf13/cobra"
)

const (
	hpGlobalMemTotalBytes = "1.3.6.1.4.1.11.2.14.11.5.1.1.2.2.1.1.5.1"
	hpGlobalMemFreeBytes  = "1.3.6.1.4.1.11.2.14.11.5.1.1.2.2.1.1.6.1"
	hpGlobalMemAllocBytes = "1.3.6.1.4.1.11.2.14.11.5.1.1.2.2.1.1.7.1"
)

func createMemoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Check the memory usage",
		Long: `Query the switch, retrieve the memory allocation and check it
against the given boundaries.`,
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
			hpGlobalMemTotalBytes,
			hpGlobalMemFreeBytes,
			hpGlobalMemAllocBytes,
		})
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// extract values from the snmp response
		memAlloc := gosnmp.ToBigInt(result.Variables[2].Value).Uint64()
		memFree := gosnmp.ToBigInt(result.Variables[1].Value).Uint64()
		memTotal := gosnmp.ToBigInt(result.Variables[0].Value).Uint64()
		pctUsed := float64(memAlloc) / float64(memTotal) * 100.0

		// instance a new nagios check
		check := nagiosplugin.NewCheck()
		defer check.Finish()

		// add performance data
		err = check.AddPerfDatum("USED", "KB", float64(memAlloc/1024))
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		err = check.AddPerfDatum("FREE", "KB", float64(memFree/1024))
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		err = check.AddPerfDatum("TOTAL", "KB", float64(memTotal/1024))
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// start out with OK and work your way up to the worst
		state := nagiosplugin.OK

		if warnRange.Check(pctUsed) {
			state = nagiosplugin.WARNING
		}

		if critRange.Check(pctUsed) {
			state = nagiosplugin.CRITICAL
		}

		// add the result with a nice text
		check.AddResultf(state, "Memory %.1f%% (%d kB) used", pctUsed, memAlloc/1024)
	}

	return cmd
}
